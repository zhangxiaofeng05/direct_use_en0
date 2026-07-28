package main

import (
	"flag"

	"github.com/zhangxiaofeng05/direct_use_en0/txthinking_socks5"
)

func main() {
	port := flag.Int("port", 20808, "port to listen on")
	flag.Parse()

	// armon_go_socks5.Run(*port)
	// things_go_go_socks5.Run(*port)
	txthinking_socks5.Run(*port)

	// direct_use_en0.PrintInterfaceIndex()
}
