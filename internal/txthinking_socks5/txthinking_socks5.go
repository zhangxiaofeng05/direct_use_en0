package txthinking_socks5

import (
	"context"
	"fmt"
	"log"
	"net"

	socks5 "github.com/txthinking/socks5"
	"github.com/zhangxiaofeng05/direct_use_en0"
)

func Run(port int, iface string) {
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

	addr := fmt.Sprintf("%s:%d", direct_use_en0.Ip, port)
	s, err := socks5.NewClassicServer(addr, direct_use_en0.Ip, "", "", 0, 60)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("SOCKS5 proxy listening on %s", addr)
	log.Fatal(s.ListenAndServe(nil))
}
