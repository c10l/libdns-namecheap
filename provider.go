package namecheap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for Namecheap
type Provider struct {
	Sandbox bool

	APIUser string
	APIKey  string

	http.Client
}

func convertZoneToSLDAndTLD(zone string) (string, string, error) {
	splitZone := strings.Split(zone, ".")
	if len(splitZone) != 2 {
		return "", "", fmt.Errorf("bad zone: %s. Should be in the format <sld>.<tld>. e.g.: example.com", zone)
	}
	return splitZone[0], splitZone[1], nil
}

func convertLibdnsRecordToAPIHost(record libdns.Record) (*APIHost, error) {
	hostID, err := strconv.Atoi(record.ID)
	if err != nil {
		return nil, err
	}

	apiHost := APIHost{
		HostID:  hostID,
		Name:    record.Name,
		Type:    record.Type,
		Address: record.Value,
		TTL:     int(record.TTL / time.Second),
	}

	return &apiHost, nil
}

func convertAPIHostToLibdnsRecord(apiHost APIHost) libdns.Record {
	return libdns.Record{
		ID:    fmt.Sprint(apiHost.HostID),
		Type:  apiHost.Type,
		Name:  apiHost.Name,
		Value: apiHost.Address,
		TTL:   time.Duration(time.Duration(apiHost.TTL) * time.Second),
	}
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	sld, tld, err := convertZoneToSLDAndTLD(zone)
	if err != nil {
		return nil, err
	}

	records, err := p.getHosts(ctx, APIGetHostsRequest{SLD: sld, TLD: tld})
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var createdRecords []libdns.Record

	sld, tld, err := convertZoneToSLDAndTLD(zone)
	if err != nil {
		return nil, err
	}

	existingRecords, err := p.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	var apiHostsToSet []APIHost
	for _, existingRecord := range existingRecords {

		// Keep each existing host to the authoritative list
		existingAPIHost, err := convertLibdnsRecordToAPIHost(existingRecord)
		if err != nil {
			return nil, err
		}
		apiHostsToSet = append(apiHostsToSet, *existingAPIHost)
	}

	for _, newRecord := range records {
		for _, existingRecord := range existingRecords {
			// Return an error if trying to append an existing record
			if existingRecord.Type == newRecord.Type && existingRecord.Name == newRecord.Name {
				return nil, fmt.Errorf("already exists: %+v", newRecord)
			}
		}

		// Add new host to authoritative list
		newAPIHost, err := convertLibdnsRecordToAPIHost(newRecord)
		if err != nil {
			return nil, err
		}
		apiHostsToSet = append(apiHostsToSet, *newAPIHost)

		// Add each new host to list of newly created records
		createdRecords = append(createdRecords, newRecord)
	}

	// Set the authoritative list of hosts with existing + new records
	if err := p.setHosts(ctx, APISetHostsRequest{
		SLD:   sld,
		TLD:   tld,
		Hosts: apiHostsToSet,
	}); err != nil {
		return nil, err
	}

	return createdRecords, nil
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deletedRecords []libdns.Record

	// zoneID, err := p.getZoneID(ctx, zone)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, record := range records {
	// 	deletedRecord, err := p.deleteRecord(ctx, zoneID, record)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	deletedRecord.TTL = time.Duration(deletedRecord.TTL) * time.Second
	// 	deletedRecords = append(deletedRecords, deletedRecord)
	// }

	return deletedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
// NOTE: Not implemented
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return nil, errors.New("not implemented")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
