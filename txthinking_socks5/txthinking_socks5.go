package txthinking_socks5

import (
	"context"
	"fmt"
	"log"
	"net"

	socks5 "github.com/txthinking/socks5"
	"github.com/zhangxiaofeng05/direct_use_en0"
)

func Run(port int) {
	socks5.DialTCP = func(network, _, raddr string) (net.Conn, error) {
		return direct_use_en0.Dial(context.Background(), network, raddr)
	}

	addr := fmt.Sprintf("%s:%d", direct_use_en0.Ip, port)
	s, err := socks5.NewClassicServer(addr, direct_use_en0.Ip, "", "", 0, 60)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on %s", addr)
	log.Fatal(s.ListenAndServe(nil))
}
