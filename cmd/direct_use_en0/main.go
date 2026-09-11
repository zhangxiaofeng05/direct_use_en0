package main

import (
	"flag"
	"log"

	"github.com/zhangxiaofeng05/direct_use_en0/internal/http_proxy"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/mixed_proxy"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/txthinking_socks5"
)

func main() {
	port := flag.Int("port", 30808, "port to listen on")
	interfaceName := flag.String("interfaceName", "en0", "network sinterface name")
	proxyType := flag.String("proxyType", "mixed", "proxy type: mixed, socks5 or http")
	flag.Parse()

	switch *proxyType {
	case "mixed":
		mixed_proxy.Run(*port, *interfaceName)
	case "socks5":
		// armon_go_socks5.Run(*port)
		// things_go_go_socks5.Run(*port)
		txthinking_socks5.Run(*port, *interfaceName)
	case "http":
		http_proxy.Run(*port, *interfaceName)
	default:
		log.Fatal("proxy type must be mixed, socks5 or http")
	}
}
