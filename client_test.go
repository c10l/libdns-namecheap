package namecheap

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

var provider = Provider{
	APIKey:  os.Getenv("NAMECHEAP_API_KEY"),
	APIUser: os.Getenv("NAMECHEAP_API_USER"),
	Sandbox: true,
}

func TestProvider_getHostsNotFound(t *testing.T) {
	sld := "notfound"
	tld := "com"

	hosts, err := provider.getHosts(context.TODO(), APIGetHostsRequest{SLD: sld, TLD: tld})
	if err.Error() != fmt.Sprintf("{SLD:%s TLD:%s} [Domain name not found]", sld, tld) {
		t.Error(err)
	}
	if len(hosts) > 0 {
		t.Errorf("Expected 0 hosts, got %d: %+v", len(hosts), hosts)
	}
}

func TestProvider_getHosts(t *testing.T) {
	sld := "gethosts-0"
	tld := "com"

	// Ensure we actually have 0 hosts
	provider.setHosts(context.TODO(), APISetHostsRequest{SLD: sld, TLD: tld})

	hosts, err := provider.getHosts(context.TODO(), APIGetHostsRequest{SLD: sld, TLD: tld})
	if err != nil {
		t.Error(err)
	}
	if len(hosts) > 0 {
		t.Errorf("Expected 0 hosts, got %d: %+v", len(hosts), hosts)
	}
}

func TestProvider_setHosts(t *testing.T) {
	sld := "sethosts"
	tld := "com"

	host0 := APIHost{
		Name:    fmt.Sprintf("foo%d", rand.Intn(100)),
		Type:    "A",
		Address: fmt.Sprintf("%d.%d.%d.%d", rand.Intn(254)+1, rand.Intn(254)+1, rand.Intn(254)+1, rand.Intn(254)+1),
		TTL:     rand.Intn(59940) + 60,
	}
	host1 := APIHost{
		Name:    fmt.Sprintf("bar%d", rand.Intn(100)),
		Type:    "CNAME",
		Address: "foo.test1.com",
		TTL:     rand.Intn(59940) + 60,
	}

	request := APISetHostsRequest{
		SLD:   sld,
		TLD:   tld,
		Hosts: []APIHost{host0, host1},
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