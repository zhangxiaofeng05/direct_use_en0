package armon_go_socks5

import (
	"fmt"
	"log"

	socks5 "github.com/armon/go-socks5"
	"github.com/zhangxiaofeng05/direct_use_en0"
)

// github.com/armon/go-socks5
func Run(port int) {
	cfg := &socks5.Config{
		Dial: direct_use_en0.Dial,
	}
	server, _ := socks5.New(cfg)

	addr := fmt.Sprintf("%s:%d", direct_use_en0.Ip, port)
	log.Printf("listening on %s", addr)
	log.Fatal(server.ListenAndServe("tcp", addr))
}
