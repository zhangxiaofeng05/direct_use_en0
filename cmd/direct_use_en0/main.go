package main

import (
	"flag"

	"github.com/zhangxiaofeng05/direct_use_en0/internal/check_ip"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/txthinking_socks5"
)

func main() {
	port := flag.Int("port", 20808, "port to listen on")
	checkIp := flag.Bool("checkIp", false, "check ip")
	interfaceName := flag.String("interfaceName", "en0", "network interface name")
	flag.Parse()

	if *checkIp {
		check_ip.CheckIp(*port, *interfaceName)
		return
	}

	// armon_go_socks5.Run(*port)
	// things_go_go_socks5.Run(*port)
	txthinking_socks5.Run(*port, *interfaceName)

	// direct_use_en0.PrintInterfaceIndex()
}
