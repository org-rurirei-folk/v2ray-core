package dns_test

import (
	"testing"

	"github.com/miekg/dns"

	dns_proto "github.com/v2fly/v2ray-core/v4/common/protocol/dns"
)

func TestDNSParseIPQuery(t *testing.T) {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{Name: "google.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}

	m, err := m1.Pack()
	if err != nil {
		t.Errorf("%v", err)
	}

	isIPQuery, domain, id, qType := dns_proto.ParseIPQuery(m)
	if isIPQuery {
		if domain != "google.com." {
			t.Errorf("not specified domain: %s", domain)
		}
	} else {
		t.Error("not ip query")
	}
}
