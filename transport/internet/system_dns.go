package internet

import (
	"context"
	"net"
)

const SystemDNS = "8.8.8.8:53"

type DNSResolverFunc func() *net.Resolver

var NewDNSResolver DNSResolverFunc = func() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, network, address)
		},
	}
}
