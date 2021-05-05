// +build android

package internet

import (
	"context"
	"net"
)

const SystemDNS = "8.8.8.8:53"

func init() {
	NewSystemResolver = func() *net.Resolver {
		return &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, network, SystemDNS)
			},
		}
	}
}
