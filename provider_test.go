package namecheap

import (
	"context"
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
