package namecheap

import (
	"context"

	"github.com/libdns/libdns"
)

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
	Hosts         []APIHost `xml:"Host"`
}

type APIHost struct {
	XMLName xml.Name `xml:"Host"`
	HostId  int      `xml:"HostId,attr"`
	Name    string   `xml:"Name,attr"`
	Type    string   `xml:"Type,attr"`
	Address string   `xml:"Address,attr"`
	MXPref  int      `xml:"MXPref,attr"`
	TTL     int      `xml:"TTL,attr"`
}

type APISetHostsResult struct {
	XMLName   xml.Name `xml:"DomainDNSSetHostsResult"`
	Domain    string   `xml:"Domain,attr"`
	IsSuccess bool     `xml:"IsSuccess,attr"`
}

// Zone ...
type Zone struct {
	TLD string
	SLD string
}

func (z *Zone)

func (p *Provider) getHosts(ctx context.Context) ([]libdns.Record, error) {
	records := []libdns.Record{}
	return records, nil
}
