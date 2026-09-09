package main

import (
	"flag"
	"log"

	"github.com/zhangxiaofeng05/direct_use_en0/internal/check_ip"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/http_proxy"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/txthinking_socks5"
)

func main() {
	port := flag.Int("port", 20808, "port to listen on")
	checkIp := flag.Bool("checkIp", false, "check ip")
	interfaceName := flag.String("interfaceName", "en0", "network sinterface name")
	proxyType := flag.String("proxyType", "socks5", "proxy type: socks5 or http")
	flag.Parse()

	if *checkIp {
		check_ip.CheckIp(*port, *interfaceName)
		return
	}

	switch *proxyType {
	case "socks5":
		// armon_go_socks5.Run(*port)
		// things_go_go_socks5.Run(*port)
		txthinking_socks5.Run(*port, *interfaceName)
	case "http":
		http_proxy.Run(*port, *interfaceName)
	default:
		log.Fatal("proxy type must be socks5 or http")
	}
}
