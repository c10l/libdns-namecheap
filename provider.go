package namecheap

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/libdns/libdns"
	"gotest.tools/assert/cmp"
)

// Provider implements the libdns interfaces for Namecheap
type Provider struct {
	Sandbox bool

	APIUser string
	APIKey  string

	http.Client
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

	var apiHostsToSet []*APIHost
	for _, existingRecord := range existingRecords {
		// Keep each existing host to the authoritative list
		existingAPIHost := convertLibdnsRecordToAPIHost(existingRecord)
		apiHostsToSet = append(apiHostsToSet, existingAPIHost)
	}

	for _, newRecord := range records {
		for _, existingRecord := range existingRecords {
			// Return an error if trying to append an existing record
			comp := cmp.Equal(convertLibdnsRecordToAPIHost(newRecord), convertLibdnsRecordToAPIHost(existingRecord))
			if comp().Success() {
				return nil, fmt.Errorf("already exists: %+v", newRecord)
			}
		}

		// Add new host to authoritative list
		newAPIHost := convertLibdnsRecordToAPIHost(newRecord)
		apiHostsToSet = append(apiHostsToSet, newAPIHost)

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

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deletedRecords []libdns.Record

	sld, tld, err := convertZoneToSLDAndTLD(zone)
	if err != nil {
		return nil, err
	}

	existingRecords, err := p.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	var apiHostsToSet []*APIHost
	for _, existingRecord := range existingRecords {
		// Keep each existing host to the authoritative list
		existingAPIHost := convertLibdnsRecordToAPIHost(existingRecord)
		apiHostsToSet = append(apiHostsToSet, existingAPIHost)
	}

	for i, recordToDelete := range records {
		apiHostToDelete := convertLibdnsRecordToAPIHost(recordToDelete)
		for j, existingAPIHost := range apiHostsToSet {
			if *existingAPIHost == *apiHostToDelete {
				apiHostsToSet[j] = nil
				deletedRecords = append(deletedRecords, recordToDelete)
				continue
			}
		}
		if len(deletedRecords) < i+1 {
			return nil, fmt.Errorf("not found: %+v", recordToDelete)
		}
	}

	// Delete nil values from list
	for i, apiHost := range apiHostsToSet {
		if apiHost == nil {
			apiHostsToSet[i] = apiHostsToSet[len(apiHostsToSet)-1] // Copy last element to index i.
			apiHostsToSet = apiHostsToSet[:len(apiHostsToSet)-1]   // Truncate slice.
		}
	}

	err = p.setHosts(ctx, APISetHostsRequest{
		SLD:   sld,
		TLD:   tld,
		Hosts: apiHostsToSet,
	})
	if err != nil {
		return nil, err
	}

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
