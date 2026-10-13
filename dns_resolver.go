package direct_use_en0

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	// DNSModeUDP is standard DNS over UDP (port 53).
	DNSModeUDP = "udp"
	// DNSModeTCP is standard DNS over TCP (port 53).
	DNSModeTCP = "tcp"
	// DNSModeDoH is DNS over HTTPS (port 443).
	DNSModeDoH = "doh"
	// DNSModeDoT is DNS over TLS (port 853).
	DNSModeDoT = "dot"
)

// Resolver is a custom DNS resolver that queries DNS servers through a
// specific network interface. It bypasses the system DNS configuration,
// which may route queries through a VPN or be subject to hijacking.
type Resolver struct {
	iface      string
	servers    []string
	mode       string
	client     *dns.Client
	dialer     *net.Dialer
	httpClient *http.Client
}

// NewResolver creates a custom DNS resolver bound to the given network
// interface. The servers list contains IP addresses of DNS servers.
// The mode must be one of "udp", "tcp", "doh", or "dot".
//
// For DoH mode, the servers are used as hostnames in the URL
// https://<server>/dns-query. For all other modes, the servers are used
// as IP addresses with the default port for the mode.
func NewResolver(iface string, servers []string, mode string) (*Resolver, error) {
	dialer, err := NewDialer(iface)
	if err != nil {
		return nil, fmt.Errorf("create dialer for DNS: %w", err)
	}

	if len(servers) == 0 {
		servers = strings.Split(DefaultDNSServers, ",")
	}

	r := &Resolver{
		iface:   iface,
		servers: servers,
		mode:    mode,
		dialer:  dialer,
	}

	switch mode {
	case DNSModeUDP:
		r.client = &dns.Client{
			Net:     "udp",
			Dialer:  dialer,
			Timeout: 5 * time.Second,
		}
	case DNSModeTCP:
		r.client = &dns.Client{
			Net:     "tcp",
			Dialer:  dialer,
			Timeout: 5 * time.Second,
		}
	case DNSModeDoT:
		r.client = &dns.Client{
			Net:     "tcp-tls",
			Dialer:  dialer,
			Timeout: 5 * time.Second,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
	case DNSModeDoH:
		r.httpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
			Timeout: 10 * time.Second,
		}
	default:
		return nil, fmt.Errorf("unsupported DNS mode: %s", mode)
	}

	return r, nil
}

// LookupIPAddr resolves the given host to IPv4 and IPv6 addresses by
// querying the configured DNS servers through the bound network interface.
// It first queries for A records (IPv4), then AAAA records (IPv6), and
// returns all resolved addresses.
func (r *Resolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IPAddr{{IP: ip}}, nil
	}

	var ips []net.IPAddr
	var lastErr error

	for _, qtype := range []uint16{dns.TypeA, dns.TypeAAAA} {
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(host), qtype)
		m.RecursionDesired = true

		var resp *dns.Msg
		var err error

		for _, server := range r.servers {
			resp, err = r.query(ctx, server, m)
			if err == nil {
				break
			}
			lastErr = err
		}

		if err != nil {
			continue
		}

		for _, answer := range resp.Answer {
			switch record := answer.(type) {
			case *dns.A:
				ips = append(ips, net.IPAddr{IP: record.A})
			case *dns.AAAA:
				ips = append(ips, net.IPAddr{IP: record.AAAA})
			}
		}
	}

	if len(ips) == 0 && lastErr != nil {
		return nil, lastErr
	}

	return ips, nil
}

// query sends a single DNS query to the specified server using the
// configured DNS mode.
func (r *Resolver) query(ctx context.Context, server string, m *dns.Msg) (*dns.Msg, error) {
	switch r.mode {
	case DNSModeUDP, DNSModeTCP, DNSModeDoT:
		addr := net.JoinHostPort(server, r.port())
		resp, _, err := r.client.ExchangeContext(ctx, m, addr)
		return resp, err
	case DNSModeDoH:
		return r.queryDoH(ctx, server, m)
	}
	return nil, fmt.Errorf("unsupported DNS mode: %s", r.mode)
}

// queryDoH sends a DNS query over HTTPS using the POST method with
// the DNS wire format as the request body.
func (r *Resolver) queryDoH(ctx context.Context, server string, m *dns.Msg) (*dns.Msg, error) {
	buf, err := m.Pack()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/dns-query", server)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respMsg := new(dns.Msg)
	err = respMsg.Unpack(body)
	return respMsg, err
}

// port returns the default port for the current DNS mode.
func (r *Resolver) port() string {
	switch r.mode {
	case DNSModeDoT:
		return "853"
	default:
		return "53"
	}
}
