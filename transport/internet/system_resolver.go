package internet

import (
	"context"
	"net"
)

type SystemResolverFunc func() *net.Resolver

var NewSystemResolver SystemResolverFunc = func() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, network, address)
		},
	}
}

type SystemDialerFunc func() *net.Dialer

var NewSystemDialer SystemDialerFunc = func() *net.Dialer {
	return nil
}
