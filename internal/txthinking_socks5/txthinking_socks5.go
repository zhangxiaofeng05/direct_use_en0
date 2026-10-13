package txthinking_socks5

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"

	socks5 "github.com/txthinking/socks5"
	"github.com/zhangxiaofeng05/direct_use_en0"
)

// NewServer creates a SOCKS5 server bound to the given network interface.
// It overrides the package-level socks5.DialTCP / socks5.DialUDP dialers,
// which is a process-wide side effect. When a SOCKS5 client sends a domain
// name (rather than an IP address), it is first resolved via the provided
// resolver before dialing through the bound interface.
func NewServer(port int, iface string, resolver *direct_use_en0.Resolver) (*socks5.Server, error) {
	socks5.DialTCP = func(network, _, raddr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(raddr)
		if err != nil {
			return nil, err
		}
		// Resolve domain names via the custom resolver (or system default)
		// before dialing through the bound interface
		if net.ParseIP(host) == nil {
			var ips []net.IPAddr
			var err error
			if resolver != nil {
				ips, err = resolver.LookupIPAddr(context.Background(), host)
			} else {
				ips, err = net.DefaultResolver.LookupIPAddr(context.Background(), host)
			}
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no addresses found for %s", host)
			}
			host = ips[0].IP.String()
			raddr = net.JoinHostPort(host, port)
		}
		return direct_use_en0.Dial(context.Background(), network, raddr, iface)
	}
	socks5.DialUDP = func(network, laddr, raddr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(raddr)
		if err != nil {
			return nil, err
		}
		// Resolve domain names via the custom resolver (or system default)
		// before dialing through the bound interface
		if net.ParseIP(host) == nil {
			var ips []net.IPAddr
			var err error
			if resolver != nil {
				ips, err = resolver.LookupIPAddr(context.Background(), host)
			} else {
				ips, err = net.DefaultResolver.LookupIPAddr(context.Background(), host)
			}
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no addresses found for %s", host)
			}
			host = ips[0].IP.String()
			raddr = net.JoinHostPort(host, port)
		}
		d, err := direct_use_en0.NewDialer(iface)
		if err != nil {
			return nil, err
		}
		if laddr != "" {
			la, err := net.ResolveUDPAddr(network, laddr)
			if err != nil {
				return nil, err
			}
			d.LocalAddr = la
		}
		return d.DialContext(context.Background(), network, raddr)
	}

	addr := direct_use_en0.Addr(port)
	s, err := socks5.NewClassicServer(addr, direct_use_en0.Ip, "", "", 0, 60)
	if err != nil {
		return nil, err
	}
	s.Handle = &socks5.DefaultHandle{}
	return s, nil
}

func Run(port int, iface string, resolver *direct_use_en0.Resolver) {
	s, err := NewServer(port, iface, resolver)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("SOCKS5 proxy listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe(nil))
}

// ListenAndServeUDP listens on the server's UDP address and serves
// SOCKS5 UDP ASSOCIATE datagrams. Standalone mode gets this via
// socks5.Server.ListenAndServe; mixed mode owns the TCP listener itself
// and uses this for the UDP part.
func ListenAndServeUDP(s *socks5.Server) error {
	addr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return err
	}
	s.UDPConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	for {
		b := make([]byte, 65507)
		n, addr, err := s.UDPConn.ReadFromUDP(b)
		if err != nil {
			return err
		}
		go func(addr *net.UDPAddr, b []byte) {
			d, err := socks5.NewDatagramFromBytes(b)
			if err != nil {
				log.Println(err)
				return
			}
			if d.Frag != 0x00 {
				log.Println("Ignore frag", d.Frag)
				return
			}
			if err := s.Handle.UDPHandle(s, addr, d); err != nil {
				// Ignore errors caused by the client going away (ICMP
				// port unreachable surfaces as ECONNRESET on UDP sockets).
				if !errors.Is(err, syscall.EPIPE) &&
					!errors.Is(err, syscall.ECONNRESET) {
					log.Println(err)
				}
			}
		}(addr, b[0:n])
	}
}
