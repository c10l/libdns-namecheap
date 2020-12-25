package namecheap

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
)

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
