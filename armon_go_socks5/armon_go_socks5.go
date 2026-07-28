package armon_go_socks5

import (
	"context"
	"fmt"
	"log"
	"net"
	"syscall"

	socks5 "github.com/armon/go-socks5"
	"golang.org/x/sys/unix"
)

func Run(port int) {
	printInterfaceIndex()

	cfg := &socks5.Config{
		Dial: dial,
	}
	server, _ := socks5.New(cfg)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("listening on %s", addr)
	log.Fatal(server.ListenAndServe("tcp", addr))
}

func newDialer(iface string) (*net.Dialer, error) {
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

func dial(ctx context.Context, network, addr string) (net.Conn, error) {
	d, err := newDialer("en0")
	if err != nil {
		return nil, err
	}

	return d.DialContext(ctx, network, addr)
}

func printInterfaceIndex() {
	vpnIndex, err := getInterfaceIndex("utun0")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("vpnIndex: %d", vpnIndex)
	enIndex, err := getInterfaceIndex("en0")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("enIndex: %d", enIndex)
}

func getInterfaceIndex(name string) (int, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return 0, err
	}
	return iface.Index, nil
}
