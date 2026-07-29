package direct_use_en0

import (
	"context"
	"fmt"
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
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				controlErr = err
				return
			}
			ip := net.ParseIP(host)
			if ip == nil {
				controlErr = fmt.Errorf("failed to parse IP: %s", host)
				return
			}

			index := uint32(ifi.Index)

			if ip.To4() != nil {
				if e := unix.SetsockoptInt(
					int(fd),
					unix.IPPROTO_IP,
					unix.IP_BOUND_IF,
					int(index),
				); e != nil {
					controlErr = e
					return
				}
			} else {
				if e := unix.SetsockoptInt(
					int(fd),
					unix.IPPROTO_IPV6,
					unix.IPV6_BOUND_IF,
					int(index),
				); e != nil {
					controlErr = e
					return
				}
			}
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
