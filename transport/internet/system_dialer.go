package internet

import (
	"context"
	"syscall"
	"time"

	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/session"
)

var (
	effectiveSystemDialer SystemDialer = &DefaultSystemDialer{}
)

type SystemDialer interface {
	Dial(ctx context.Context, source, destination net.Destination, sockopt *SocketConfig) (net.Conn, error)
}

type DefaultSystemDialer struct {
	controllers []controller
}

func (d *DefaultSystemDialer) Dial(ctx context.Context, src, dest net.Destination, sockopt *SocketConfig) (net.Conn, error) {
	if src.Address != nil && src.Network != dest.Network {
		return nil, newError("invalid src or dest")
	}

	if dest.Network == net.Network_UDP && !HasBindAddr(sockopt) {
		return HandleDialUDP(ctx, src, dest, sockopt)
	}

	return HandleDial(ctx, src, dest, sockopt, d.controllers)
}

func ResolveNetAddr(addr net.Destination) (net.Addr, error) {
	// addr.Network must not be nil
	// as requires src.Network = dest.Network
	if addr.Address == nil {
		addr = net.AnyDestination(addr.Network)
	}

	/* if addr.Address == nil {
		return nil, newError("empty addr")
	} */

	switch addr.Network {
	case net.Network_TCP:
		return net.ResolveTCPAddr(addr.Network.SystemString(), addr.NetAddr())
	case net.Network_UDP:
		return net.ResolveUDPAddr(addr.Network.SystemString(), addr.NetAddr())
	default:
		return nil, newError("unsupported network")
	}
}

func HasBindAddr(sockopt *SocketConfig) bool {
	return sockopt != nil && len(sockopt.BindAddress) > 0 && sockopt.BindPort > 0
}

func HandleDialUDP(ctx context.Context, src, dest net.Destination, sockopt *SocketConfig) (net.Conn, error) {
	src.Network = dest.Network
	srcAddr, err := ResolveNetAddr(src)
	if err != nil {
		return nil, err
	}

	destAddr, err := ResolveNetAddr(dest)
	if err != nil {
		return nil, err
	}

	packetConn, err := ListenSystemPacket(ctx, srcAddr, sockopt)
	if err != nil {
		return nil, err
	}

	return &packetConnWrapper{
		conn: packetConn,
		dest: destAddr,
	}, nil
}

func HandleDial(ctx context.Context, src, dest net.Destination, sockopt *SocketConfig, controllers []controller) (net.Conn, error) {
	src.Network = dest.Network
	srcAddr, err := ResolveNetAddr(src)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   time.Second * 16,
		LocalAddr: srcAddr,
	}

	if sockopt != nil || len(controllers) > 0 {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if sockopt != nil {
					if err := applyOutboundSocketOptions(network, address, fd, sockopt); err != nil {
						newError("failed to apply socket options").Base(err).WriteToLog(session.ExportIDToError(ctx))
					}
					if dest.Network == net.Network_UDP && HasBindAddr(sockopt) {
						if err := bindAddr(fd, sockopt.BindAddress, sockopt.BindPort); err != nil {
							newError("failed to bind source address to ", sockopt.BindAddress).Base(err).WriteToLog(session.ExportIDToError(ctx))
						}
					}
				}

				for _, ctl := range controllers {
					if err := ctl(network, address, fd); err != nil {
						newError("failed to apply external controller").Base(err).WriteToLog(session.ExportIDToError(ctx))
					}
				}
			})
		}
	}

	return dialer.DialContext(ctx, dest.Network.SystemString(), dest.NetAddr())
}

type packetConnWrapper struct {
	conn net.PacketConn
	dest net.Addr
}

func (c *packetConnWrapper) Close() error {
	return c.conn.Close()
}

func (c *packetConnWrapper) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *packetConnWrapper) RemoteAddr() net.Addr {
	return c.dest
}

func (c *packetConnWrapper) Write(p []byte) (int, error) {
	return c.conn.WriteTo(p, c.dest)
}

func (c *packetConnWrapper) Read(p []byte) (int, error) {
	n, _, err := c.conn.ReadFrom(p)
	return n, err
}

func (c *packetConnWrapper) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *packetConnWrapper) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *packetConnWrapper) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

type SystemDialerAdapter interface {
	Dial(network string, address string) (net.Conn, error)
}

type SimpleSystemDialer struct {
	adapter SystemDialerAdapter
}

func WithAdapter(dialer SystemDialerAdapter) SystemDialer {
	return &SimpleSystemDialer{
		adapter: dialer,
	}
}

func (v *SimpleSystemDialer) Dial(ctx context.Context, src, dest net.Destination, sockopt *SocketConfig) (net.Conn, error) {
	return v.adapter.Dial(dest.Network.SystemString(), dest.NetAddr())
}

// UseAlternativeSystemDialer replaces the current system dialer with a given one.
// Caller must ensure there is no race condition.
//
// v2ray:api:stable
func UseAlternativeSystemDialer(dialer SystemDialer) {
	if dialer == nil {
		dialer = &DefaultSystemDialer{}
	}
	effectiveSystemDialer = dialer
}

// RegisterDialerController adds a controller to the effective system dialer.
// The controller can be used to operate on file descriptors before they are put into use.
// It only works when effective dialer is the default dialer.
//
// v2ray:api:beta
func RegisterDialerController(ctl func(network, address string, fd uintptr) error) error {
	if ctl == nil {
		return newError("nil listener controller")
	}

	dialer, ok := effectiveSystemDialer.(*DefaultSystemDialer)
	if !ok {
		return newError("RegisterListenerController not supported in custom dialer")
	}

	dialer.controllers = append(dialer.controllers, ctl)
	return nil
}
