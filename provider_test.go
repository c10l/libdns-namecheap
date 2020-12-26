package namecheap

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/libdns/libdns"
)

func ProviderFactory() Provider {
	return Provider{
		APIKey:  os.Getenv("NAMECHEAP_API_KEY"),
		APIUser: os.Getenv("NAMECHEAP_API_USER"),
		Sandbox: true,
	}
}

func HostFactory() APIHost {
	validTypes := []string{
		APIHostTypeA,
		APIHostTypeAAAA,
		APIHostTypeALIAS,
		APIHostTypeCNAME,
		APIHostTypeMX,
		APIHostTypeMXE,
		APIHostTypeTXT,
		APIHostTypeURL,
		APIHostTypeURL301,
		APIHostTypeFRAME,
	}

	hostType := validTypes[rand.Intn(len(validTypes))]
	apiHost := APIHost{
		Name:    fmt.Sprintf("test-host-%d", rand.Intn(100)),
		Address: fmt.Sprintf("%d.%d.%d.%d", rand.Intn(254)+1, rand.Intn(254)+1, rand.Intn(254)+1, rand.Intn(254)+1),
		TTL:     rand.Intn(59940) + 60,
		Type:    hostType,
	}
	if hostType == APIHostTypeCNAME {
		apiHost.Address = fmt.Sprintf("www%d.example.com", rand.Int())
	}
	if hostType == APIHostTypeMX {
		apiHost.MXPref = 1
		apiHost.Address = fmt.Sprintf("mail%d.example.com", rand.Int())
	}
	return apiHost
}

func TestProvider_GetRecordsDomainNotFound(t *testing.T) {
	provider := ProviderFactory()
	zone := "notfound.com"

	hosts, err := provider.GetRecords(context.TODO(), zone)
	if err.Error() != "{SLD:notfound TLD:com} [Domain name not found]" {
		t.Error(err)
	}
	if len(hosts) > 0 {
		t.Errorf("Expected 0 hosts, got %d: %+v", len(hosts), hosts)
	}
}

func TestProvider_GetRecords(t *testing.T) {
	provider := ProviderFactory()
	zone := "gethosts-0.com"

	// Ensure we actually have 0 hosts
	provider.setHosts(context.TODO(), APISetHostsRequest{SLD: "gethosts-0", TLD: "com"})

	hosts, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		t.Error(err)
	}
	if len(hosts) > 0 {
		t.Errorf("Expected 0 hosts, got %d: %+v", len(hosts), hosts)
	}
}

func TestProvider_SetRecords(t *testing.T) {
	provider := ProviderFactory()
	_, err := provider.SetRecords(context.TODO(), "foo", []libdns.Record{})
	if err.Error() != "not implemented" {
		t.Error(err)
	}
}

func TestProvider_AppendRecords(t *testing.T) {
	provider := ProviderFactory()
	sld := "sethosts"
	tld := "com"
	zone := fmt.Sprintf("%s.%s", sld, tld)

	hosts := []APIHost{HostFactory(), HostFactory()}
	request := APISetHostsRequest{
		SLD:   sld,
		TLD:   tld,
		Hosts: hosts,
	}

	// Set up test with some records
	err := provider.setHosts(context.TODO(), request)
	if err != nil {
		t.Error(err)
	}

	recordsToAppend := []libdns.Record{}

	// Test not appending any records
	appendedRecords, err := provider.AppendRecords(context.TODO(), zone, recordsToAppend)
	if err != nil {
		t.Error(err)
	}
	if len(appendedRecords) != 0 {
		t.Errorf("Expected 0 appendedRecords, got %d: %+v", len(appendedRecords), appendedRecords)
	}

	// Check that the previous hosts have not been deleted
	remainingHosts, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		t.Error(err)
	}
	if len(remainingHosts) != 2 {
		t.Errorf("Expected 2 remainingHosts, got %d: %+v", len(remainingHosts), remainingHosts)
	}

	// Test appending 2 records
	recordsToAppend = append(recordsToAppend, convertAPIHostToLibdnsRecord(HostFactory()))
	recordsToAppend = append(recordsToAppend, convertAPIHostToLibdnsRecord(HostFactory()))
	appendedRecords, err = provider.AppendRecords(context.TODO(), zone, recordsToAppend)
	if err != nil {
		t.Error(err)
	}
	if len(appendedRecords) != len(recordsToAppend) {
		message := fmt.Sprintf("\nPrevious Records: %+v", hosts)
		message += fmt.Sprintf("\nExpected %d appendedRecords, got %d: %+v", len(recordsToAppend), len(appendedRecords), appendedRecords)
		t.Error(message)
	}

	// Check that the appendedRecords match the recordsToAppend
	matches := 0
	for _, recordToAppend := range recordsToAppend {
		for _, appendedRecord := range appendedRecords {
			if recordToAppend == appendedRecord {
				matches++
			}
		}
	}
	if matches != 2 {
		t.Errorf("Expected 2 appendedRecords, got %d", len(appendedRecords))
	}
}
