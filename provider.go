package namecheap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
		return "", "", fmt.Errorf("Bad zone: %s. Should be in the format <sld>.<tld>. e.g.: example.com", zone)
	}

	return splitZone[0], splitZone[1], nil
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

	// zoneID, err := p.getZoneID(ctx, zone)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, record := range records {
	// 	newRecord, err := p.createRecord(ctx, zoneID, record)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	newRecord.TTL = time.Duration(newRecord.TTL) * time.Second
	// 	createdRecords = append(createdRecords, newRecord)
	// }

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
