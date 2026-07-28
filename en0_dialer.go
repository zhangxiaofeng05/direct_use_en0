package direct_use_en0

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func NewDialer(iface string) (*net.Dialer, error) {
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}

	d := &net.Dialer{}

	d.Control = func(network, address string, c syscall.RawConn) error {
		var controlErr error

		err := c.Control(func(fd uintptr) {
			index := uint32(ifi.Index)

			// IPv4
			if e := unix.SetsockoptInt(
				int(fd),
				unix.IPPROTO_IP,
				unix.IP_BOUND_IF,
				int(index),
			); e != nil {
				controlErr = e
				return
			}

			// IPv6（如果是 IPv6，可以再设置 IPV6_BOUND_IF）
			_ = unix.SetsockoptInt(
				int(fd),
				unix.IPPROTO_IPV6,
				unix.IPV6_BOUND_IF,
				int(index),
			)
		})

		if err != nil {
			return err
		}

		return controlErr
	}

	return d, nil
}

func Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	d, err := NewDialer(En0InterfaceName)
	if err != nil {
		return nil, err
	}

	return d.DialContext(ctx, network, addr)
}
