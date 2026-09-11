package txthinking_socks5

import (
	"context"
	"log"
	"net"

	socks5 "github.com/txthinking/socks5"
	"github.com/zhangxiaofeng05/direct_use_en0"
)

// NewServer creates a SOCKS5 server bound to the given network interface.
// It overrides the package-level socks5.DialTCP / socks5.DialUDP dialers,
// which is a process-wide side effect.
func NewServer(port int, iface string) (*socks5.Server, error) {
	socks5.DialTCP = func(network, _, raddr string) (net.Conn, error) {
		return direct_use_en0.Dial(context.Background(), network, raddr, iface)
	}
	socks5.DialUDP = func(network, laddr, raddr string) (net.Conn, error) {
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

func Run(port int, iface string) {
	s, err := NewServer(port, iface)
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
				log.Println(err)
			}
		}(addr, b[0:n])
	}
}
