package things_go_go_socks5

import (
	"fmt"
	"log"
	"os"

	"github.com/things-go/go-socks5"
	"github.com/zhangxiaofeng05/direct_use_en0"
)

// reference: https://github.com/things-go/go-socks5/blob/master/_example/main.go
func Run(port int) {
	// Create a SOCKS5 server
	server := socks5.NewServer(
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
		socks5.WithDial(direct_use_en0.Dial),
	)

	// Create SOCKS5 proxy on localhost port
	network := "tcp"
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("listening on %s", addr)
	if err := server.ListenAndServe(network, addr); err != nil {
		log.Fatal(err)
	}
}
