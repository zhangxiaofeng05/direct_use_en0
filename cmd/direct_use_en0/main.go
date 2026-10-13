package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/zhangxiaofeng05/direct_use_en0"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/http_proxy"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/mixed_proxy"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/txthinking_socks5"
)

func main() {
	port := flag.Int("port", 20808, "port to listen on")
	interfaceName := flag.String("interfaceName", "en0", "network interface name")
	proxyType := flag.String("proxyType", "mixed", "proxy type: mixed, socks5 or http")

	// DNS flags — can also be set via DNS_SERVERS / DNS_MODE env vars
	dnsServers := flag.String("dnsServers", envOr("DNS_SERVERS", "8.8.8.8,1.1.1.1"),
		"comma-separated DNS server list (e.g. 8.8.8.8,1.1.1.1)")
	dnsMode := flag.String("dnsMode", envOr("DNS_MODE", "udp"),
		"DNS mode: udp, tcp, doh, or dot")
	flag.Parse()

	servers := parseServers(*dnsServers)
	resolver, err := direct_use_en0.NewResolver(*interfaceName, servers, *dnsMode)
	if err != nil {
		log.Fatalf("Failed to create DNS resolver: %v", err)
	}

	log.Printf("DNS resolver: mode=%s servers=%v interface=%s", *dnsMode, servers, *interfaceName)

	switch *proxyType {
	case "mixed":
		mixed_proxy.Run(*port, *interfaceName, resolver)
	case "socks5":
		txthinking_socks5.Run(*port, *interfaceName, resolver)
	case "http":
		http_proxy.Run(*port, *interfaceName, resolver)
	default:
		log.Fatal("proxy type must be mixed, socks5 or http")
	}
}

// parseServers splits a comma-separated list of DNS servers and trims
// whitespace from each entry.
func parseServers(s string) []string {
	parts := strings.Split(s, ",")
	servers := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			servers = append(servers, p)
		}
	}
	return servers
}

// envOr returns the value of the named environment variable, or def if
// the variable is unset or empty.
func envOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}
