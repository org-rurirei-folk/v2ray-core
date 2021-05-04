package internet

import (
	"context"
	"testing"
)

func TestSystemResolver(t *testing.T) {
	resolver := NewSystemResolver()
	if ips, err := resolver.LookupIP(context.Background(), "ip", "www.google.com"); err != nil {
		t.Errorf("failed to LookupIP, %v, %v", ips, err)
	}
}
