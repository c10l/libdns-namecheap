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
	SLD   string
	TLD   string
	Hosts []*APIHost
}

type APIResponse struct {
	XMLName          xml.Name           `xml:"ApiResponse"`
	Status           string             `xml:"Status,attr"`
	Errors           []APIError         `xml:"Errors>Error"`
	RequestedCommand string             `xml:"RequestedCommand"`
	CommandResponse  APICommandResponse `xml:"CommandResponse"`
}

type APIError struct {
	Number  int    `xml:"number,attr"`
	Message string `xml:",chardata"`
}

type APICommandResponse struct {
	XMLName        xml.Name          `xml:"CommandResponse"`
	Type           string            `xml:"Type,attr"`
	GetHostsResult APIGetHostsResult `xml:"DomainDNSGetHostsResult"`
	SetHostsResult APISetHostsResult `xml:"DomainDNSSetHostsResult"`
}

type APIGetHostsResult struct {
	XMLName       xml.Name   `xml:"DomainDNSGetHostsResult"`
	Domain        string     `xml:"Domain,attr"`
	IsUsingOurDNS bool       `xml:"IsUsingOurDNS,attr"`
	Hosts         []*APIHost `xml:"host"`
}

type APISetHostsResult struct {
	XMLName   xml.Name `xml:"DomainDNSSetHostsResult"`
	Domain    string   `xml:"Domain,attr"`
	IsSuccess bool     `xml:"IsSuccess,attr"`
}

const APIHostTypeA = "A"
const APIHostTypeAAAA = "AAAA"
const APIHostTypeALIAS = "ALIAS"
const APIHostTypeCAA = "CAA"
const APIHostTypeCNAME = "CNAME"
const APIHostTypeMX = "MX"
const APIHostTypeMXE = "MXE"
const APIHostTypeNS = "NS"
const APIHostTypeTXT = "TXT"
const APIHostTypeURL = "URL"
const APIHostTypeURL301 = "URL301"
const APIHostTypeFRAME = "FRAME"

type APIHost struct {
	XMLName xml.Name `xml:"host"`
	Name    string   `xml:"Name,attr"`
	Type    string   `xml:"Type,attr"`
	Address string   `xml:"Address,attr"`
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

func convertZoneToSLDAndTLD(zone string) (string, string, error) {
	splitZone := strings.Split(zone, ".")
	if len(splitZone) != 2 {
		return "", "", fmt.Errorf("bad zone: %s. Should be in the format <sld>.<tld>. e.g.: example.com", zone)
	}
	return splitZone[0], splitZone[1], nil
}

func convertLibdnsRecordToAPIHost(record libdns.Record) *APIHost {
	return &APIHost{
		Type:    record.Type,
		Name:    record.Name,
		Address: record.Value,
		TTL:     int(record.TTL / time.Second),
	}
}

func convertAPIHostToLibdnsRecord(apiHost APIHost) libdns.Record {
	return libdns.Record{
		Type:  apiHost.Type,
		Name:  apiHost.Name,
		Value: apiHost.Address,
		TTL:   time.Duration(time.Duration(apiHost.TTL) * time.Second),
	}
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
