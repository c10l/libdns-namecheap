package namecheap

import (
	"context"
	"testing"
)

func TestProvider_setHosts(t *testing.T) {
	provider := ProviderFactory()
	sld := "sethosts"
	tld := "com"

	host1 := HostFactory()
	host0 := HostFactory()

	request := APISetHostsRequest{
		SLD:   sld,
		TLD:   tld,
		Hosts: []*APIHost{host1, host0},
	}

	err := provider.setHosts(context.TODO(), request)
	if err != nil {
		t.Errorf("%+v", err)
	}

	hosts, err := provider.getHosts(context.TODO(), APIGetHostsRequest{SLD: sld, TLD: tld})
	if err != nil {
		t.Error(err)
	}
	if len(hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d: %+v", len(hosts), hosts)
	}
	if err := host0.MatchRecord(hosts[0]); err != nil {
		t.Error(err)
	}
	if err := host1.MatchRecord(hosts[1]); err != nil {
		t.Error(err)
	}
}
