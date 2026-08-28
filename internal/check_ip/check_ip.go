package check_ip

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/zhangxiaofeng05/com/com_proxy"
	"github.com/zhangxiaofeng05/com/third_party/ip/ipsb"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/txthinking_socks5"
)

func CheckIp(port int, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go txthinking_socks5.Run(port, name)
	time.Sleep(500 * time.Millisecond)

	proxyStr := "socks5://127.0.0.1:" + strconv.Itoa(port)
	proxyClient, err := com_proxy.HttpClient(proxyStr)
	if err != nil {
		log.Fatal(err)
	}

	directClient := &http.Client{Timeout: 10 * time.Second}

	directGeoIp, err := ipsb.GeoIp(ctx, directClient)
	if err != nil {
		log.Fatal(err)
	}
	proxyGeoIp, err := ipsb.GeoIp(ctx, proxyClient)
	if err != nil {
		log.Fatal(err)
	}
	diff := cmp.Diff(directGeoIp, proxyGeoIp)
	if diff == "" {
		log.Printf("same ip: %s country: %s city: %s", directGeoIp.Ip, directGeoIp.Country, directGeoIp.City)
	} else {
		log.Printf("ip is diff: %v", diff)
	}
}
