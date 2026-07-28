package main

import (
	"flag"

	"github.com/zhangxiaofeng05/direct_use_en0/armon_go_socks5"
)

func main() {
	port := flag.Int("port", 1080, "port to listen on")
	flag.Parse()

	armon_go_socks5.Run(*port)
}
