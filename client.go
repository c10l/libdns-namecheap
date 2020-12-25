package namecheap

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type APIGetHostsRequest struct {
	SLD string
	TLD string
}

type APISetHostsRequest struct {
	SLD       string
	TLD       string
	EmailType string
	Flag      string
	Tag       string
	Hosts     []APIHost
}

type APIResponse struct {
	XMLName          xml.Name           `xml:"ApiResponse"`
	Status           string             `xml:"Status,attr"`
	Errors           []APIError         `xml:"Errors>Error"`
	RequestedCommand string             `xml:"RequestedCommand"`
	CommandResponse  APICommandResponse `xml:"CommandResponse"`
}

type APIError struct {
	Message string `xml:",chardata"`
	Number  int    `xml:"number,attr"`
}

type APICommandResponse struct {
	XMLName        xml.Name          `xml:"CommandResponse"`
	Type           string            `xml:"Type,attr"`
	GetHostsResult APIGetHostsResult `xml:"DomainDNSGetHostsResult"`
	SetHostsResult APISetHostsResult `xml:"DomainDNSSetHostsResult"`
}

type APIGetHostsResult struct {
	XMLName       xml.Name  `xml:"DomainDNSGetHostsResult"`
	Domain        string    `xml:"Domain,attr"`
	IsUsingOurDNS bool      `xml:"IsUsingOurDNS,attr"`
	Hosts         []APIHost `xml:"host"`
}

type APISetHostsResult struct {
	XMLName   xml.Name `xml:"DomainDNSSetHostsResult"`
	Domain    string   `xml:"Domain,attr"`
	IsSuccess bool     `xml:"IsSuccess,attr"`
}

type APIHost struct {
	XMLName xml.Name `xml:"host"`
	HostID  int      `xml:"HostId,attr"`
	Name    string   `xml:"Name,attr"`
	Type    string   `xml:"Type,attr"`
	Address string   `xml:"Address,attr"`
	MXPref  int      `xml:"MXPref,attr"`
	TTL     int      `xml:"TTL,attr"`
}

// MatchRecord compares the APIHost with a libdns.Record. It returns an error if the records don't match.
func (h *APIHost) MatchRecord(other libdns.Record) error {
	switch {
	case h.Type != other.Type:
		return fmt.Errorf("Wrong Type: %s", other.Type)
	case h.Name != other.Name:
		return fmt.Errorf("Wrong Name: %s", other.Name)
	case h.Address != strings.TrimSuffix(other.Value, "."):
		return fmt.Errorf("Wrong Address: %s", other.Value)
	case time.Duration(h.TTL)*time.Second != other.TTL:
		return fmt.Errorf("Wrong TTL %d", other.TTL)
	default:
		return nil
	}
}

func (p *Provider) buildURL(command string) string {
	var host string
	if p.Sandbox {
		host = "api.sandbox.namecheap.com"
	} else {
		host = "api.namecheap.com"
	}
	commonParameters := fmt.Sprintf("ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.dns.%s&ClientIp=0.0.0.0", p.APIUser, p.APIKey, p.APIUser, command)
	return fmt.Sprintf("https://%s/xml.response?%s", host, commonParameters)
}

func parseResponse(apiResponse *APIResponse, body []byte) []error {
	if err := xml.Unmarshal(body, &apiResponse); err != nil {
		log.Fatalln(err)
	}
	if len(apiResponse.Errors) > 0 {
		errorsList := new([]error)
		for _, e := range apiResponse.Errors {
			*errorsList = append(*errorsList, errors.New(e.Message))
		}
		return *errorsList
	}
	return nil
}

func (p *Provider) getHosts(ctx context.Context, params APIGetHostsRequest) ([]libdns.Record, error) {
	var records []libdns.Record
	url := fmt.Sprintf("%s&SLD=%s&TLD=%s", p.buildURL("getHosts"), params.SLD, params.TLD)

	resp, err := http.Post(url, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	apiResponse := new(APIResponse)
	if err := parseResponse(apiResponse, body); err != nil {
		return nil, fmt.Errorf("%+v %+v", params, err)
	}

	for _, host := range apiResponse.CommandResponse.GetHostsResult.Hosts {
		record := libdns.Record{
			ID:    fmt.Sprint(host.HostID),
			Type:  host.Type,
			Name:  host.Name,
			Value: host.Address,
			TTL:   time.Second * time.Duration(host.TTL),
		}
		records = append(records, record)
	}

	return records, nil
}

func (p *Provider) setHosts(ctx context.Context, params APISetHostsRequest) error {
	url := fmt.Sprintf("%s&SLD=%s&TLD=%s", p.buildURL("setHosts"), params.SLD, params.TLD)

	for i, host := range params.Hosts {
		url = fmt.Sprintf("%s&HostName%d=%s", url, i, host.Name)
		url = fmt.Sprintf("%s&Address%d=%s", url, i, host.Address)
		url = fmt.Sprintf("%s&RecordType%d=%s", url, i, host.Type)
		url = fmt.Sprintf("%s&TTL%d=%d", url, i, host.TTL)
		if host.Type == "MX" {
			url = fmt.Sprintf("%s&MXPref%d=%d", url, i, host.MXPref)
		}
	}

	resp, err := http.Post(url, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	apiResponse := new(APIResponse)
	if err := parseResponse(apiResponse, body); err != nil {
		return fmt.Errorf("%+v %+v", params, err)
	}

	return nil
}
