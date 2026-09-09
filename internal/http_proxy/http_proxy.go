package http_proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zhangxiaofeng05/direct_use_en0"
)

var (
	ipv4Dialer *net.Dialer
	ipv6Dialer *net.Dialer

	transport *http.Transport
)

func Run(port int, iface string) {
	ipv4, ipv6, err := getInterfaceIPs(iface)
	if err != nil {
		log.Fatalf("get interface IP failed: %v", err)
	}

	if ipv4 != nil {
		log.Printf("IPv4: %s", ipv4)
		ipv4Dialer = &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP: ipv4,
			},
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
	}

	if ipv6 != nil {
		log.Printf("IPv6: %s", ipv6)
		ipv6Dialer = &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP: ipv6,
			},
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
	}

	if ipv4Dialer == nil && ipv6Dialer == nil {
		log.Fatal("no usable IPv4 or IPv6 address found")
	}

	transport = &http.Transport{
		Proxy:               nil,
		DialContext:         dialContext,
		ForceAttemptHTTP2:   false,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,

		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", direct_use_en0.Ip, port)
	server := &http.Server{
		Addr:         addr,
		Handler:      http.HandlerFunc(proxyHandler),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	log.Printf("HTTP proxy listening on %s", addr)

	log.Fatal(server.ListenAndServe())
}

// Get the IPv4 and IPv6 addresses of the specified network interface card.
func getInterfaceIPs(name string) (net.IP, net.IP, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, nil, err
	}

	var ipv4 net.IP
	var ipv6 net.IP

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		ip := ipNet.IP

		if ip.IsLoopback() {
			continue
		}

		// IPv4
		if ip4 := ip.To4(); ip4 != nil {
			if ipv4 == nil {
				ipv4 = ip4
			}
			continue
		}

		// IPv6
		if ip.To16() != nil {
			// ignore Link-Local addresses such as fe80::
			if ip.IsLinkLocalUnicast() {
				continue
			}

			if ipv6 == nil {
				ipv6 = ip
			}
		}
	}

	if ipv4 == nil && ipv6 == nil {
		return nil, nil, errors.New("no usable IP address found")
	}

	return ipv4, ipv6, nil
}

// Select the IPv4 or IPv6 dialer based on the final destination IP address.
func dialContext(
	ctx context.Context,
	network string,
	address string,
) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	// the address is already an IP address
	if ip := net.ParseIP(host); ip != nil {
		return dialIP(ctx, ip, port)
	}

	// resolve all addresses for the domain
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	var lastErr error

	// try IPv6 first
	for _, ipAddr := range ips {
		if ipAddr.IP.To4() == nil {
			if ipv6Dialer == nil {
				continue
			}

			conn, err := ipv6Dialer.DialContext(
				ctx,
				"tcp6",
				net.JoinHostPort(ipAddr.IP.String(), port),
			)

			if err == nil {
				return conn, nil
			}

			lastErr = err
		}
	}

	// try IPv4 next
	for _, ipAddr := range ips {
		if ipAddr.IP.To4() != nil {
			if ipv4Dialer == nil {
				continue
			}

			conn, err := ipv4Dialer.DialContext(
				ctx,
				"tcp4",
				net.JoinHostPort(ipAddr.IP.String(), port),
			)

			if err == nil {
				return conn, nil
			}

			lastErr = err
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, errors.New("no suitable address available")
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleConnect(w, r)
		return
	}

	handleHTTP(w, r)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() {
		http.Error(
			w,
			"proxy request requires absolute URL",
			http.StatusBadRequest,
		)
		return
	}

	r.RequestURI = ""

	removeHopHeaders(r.Header)

	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	removeHopHeaders(resp.Header)

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, _ = io.Copy(w, resp.Body)
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	target := r.Host

	if !strings.Contains(target, ":") {
		target += ":443"
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(
			w,
			"hijacking not supported",
			http.StatusInternalServerError,
		)
		return
	}

	clientConn, rw, err := hijacker.Hijack()
	if err != nil {
		return
	}

	defer clientConn.Close()

	// continue data processing if there are any buffered data
	if rw.Reader.Buffered() > 0 {
		// continue processing buffered data
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	targetConn, err := dialContext(ctx, "tcp", target)
	if err != nil {
		_, _ = clientConn.Write(
			[]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"),
		)
		return
	}

	defer targetConn.Close()

	_, err = clientConn.Write(
		[]byte("HTTP/1.1 200 Connection Established\r\n\r\n"),
	)
	if err != nil {
		return
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		// attack after data loss after the CONNECT request
		_, _ = io.Copy(targetConn, rw.Reader)

		if tcpConn, ok := targetConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()

		_, _ = io.Copy(clientConn, targetConn)

		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()

	wg.Wait()
}

func removeHopHeaders(header http.Header) {
	for _, key := range []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	} {
		header.Del(key)
	}
}

func dialIP(
	ctx context.Context,
	ip net.IP,
	port string,
) (net.Conn, error) {
	if ip.To4() != nil {
		if ipv4Dialer == nil {
			return nil, errors.New("IPv4 dialer not available")
		}

		return ipv4Dialer.DialContext(
			ctx,
			"tcp4",
			net.JoinHostPort(ip.String(), port),
		)
	}

	if ipv6Dialer == nil {
		return nil, errors.New("IPv6 dialer not available")
	}

	return ipv6Dialer.DialContext(
		ctx,
		"tcp6",
		net.JoinHostPort(ip.String(), port),
	)
}
