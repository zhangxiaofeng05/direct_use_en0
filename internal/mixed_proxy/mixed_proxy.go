package mixed_proxy

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	socks5 "github.com/txthinking/socks5"
	"github.com/zhangxiaofeng05/direct_use_en0"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/http_proxy"
	"github.com/zhangxiaofeng05/direct_use_en0/internal/txthinking_socks5"
)

// Run starts a proxy that serves both SOCKS5 and HTTP proxy clients on
// the same port. The protocol is detected per connection: SOCKS5 starts
// with the version byte 0x05, HTTP starts with a request method.
func Run(port int, iface string) {
	s, err := txthinking_socks5.NewServer(port, iface)
	if err != nil {
		log.Fatal(err)
	}

	p, err := http_proxy.NewProxy(iface)
	if err != nil {
		log.Fatal(err)
	}

	httpSrv := &http.Server{
		Handler:      p.Handler(),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	addr := direct_use_en0.Addr(port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := txthinking_socks5.ListenAndServeUDP(s); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Mixed SOCKS5/HTTP proxy listening on %s", addr)

	for {
		c, err := l.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(s, httpSrv, c)
	}
}

// handleConn detects the protocol by the first byte of the connection:
// 0x05 means SOCKS5, anything else is treated as HTTP.
func handleConn(s *socks5.Server, httpSrv *http.Server, c *net.TCPConn) {
	var first [1]byte
	if _, err := io.ReadFull(c, first[:]); err != nil {
		_ = c.Close()
		return
	}

	if first[0] == socks5.Ver {
		serveSocks5(s, c, first[0])
		return
	}

	serveHTTP(httpSrv, &prefixConn{Conn: c, prefix: first[:]})
}

// serveSocks5 handles a SOCKS5 connection whose version byte has already
// been consumed. The byte is replayed for the handshake, which reads
// exactly the bytes it needs, so the original connection can then be
// handed to the socks5 handler without losing any data.
func serveSocks5(s *socks5.Server, c *net.TCPConn, version byte) {
	defer c.Close()

	rw := &prefixConn{Conn: c, prefix: []byte{version}}
	if err := s.Negotiate(rw); err != nil {
		log.Println(err)
		return
	}
	r, err := s.GetRequest(rw)
	if err != nil {
		log.Println(err)
		return
	}
	if err := s.Handle.TCPHandle(s, c, r); err != nil {
		log.Println(err)
	}
}

// serveHTTP hands an established connection to an http.Server through a
// one-shot listener. Serve returns after the connection is handed off;
// the connection keeps being served in its own goroutine.
func serveHTTP(httpSrv *http.Server, conn net.Conn) {
	err := httpSrv.Serve(newOneShotListener(conn))
	if err != nil && !errors.Is(err, errListenerClosed) {
		log.Println(err)
	}
}

// prefixConn is a net.Conn wrapper that replays prefix bytes before
// reading from the wrapped connection.
type prefixConn struct {
	net.Conn
	prefix []byte
}

func (c *prefixConn) Read(b []byte) (int, error) {
	if len(c.prefix) > 0 {
		n := copy(b, c.prefix)
		c.prefix = c.prefix[n:]
		return n, nil
	}
	return c.Conn.Read(b)
}

func (c *prefixConn) CloseWrite() error {
	if cw, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return errors.New("CloseWrite not supported")
}

var errListenerClosed = errors.New("listener closed")

// oneShotListener hands a single, already-established connection to
// http.Server.Serve. A second Accept returns errListenerClosed so Serve
// returns while the first connection is still being served.
type oneShotListener struct {
	conn  net.Conn
	ready bool
}

func newOneShotListener(conn net.Conn) *oneShotListener {
	return &oneShotListener{conn: conn}
}

func (l *oneShotListener) Accept() (net.Conn, error) {
	if l.ready {
		return nil, errListenerClosed
	}
	l.ready = true
	return l.conn, nil
}

func (l *oneShotListener) Close() error {
	return nil
}

func (l *oneShotListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}
