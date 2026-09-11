# AGENTS.md

Local proxy (SOCKS5 or HTTP) that forces outbound traffic through a specific network interface (default `en0`) on macOS, bypassing the default route (e.g. a VPN on `utun*`).

## Commands

- Build: `go build ./...`
- Vet: `go vet ./...`
- Run: `go run ./cmd/direct_use_en0 -port 20808 [-interfaceName en0] [-proxyType mixed|socks5|http] [-checkIp]` (`mixed` is the default and serves SOCKS5 + HTTP on one port)
- Smoke test: `go run ./cmd/direct_use_en0 -checkIp` — starts the proxy, then compares direct vs. proxied exit IP via ipsb (needs internet access)
- There are no tests, no linter, and no CI. Build + vet is the only verification.
- Install path used in README: `go install github.com/zhangxiaofeng05/direct_use_en0/cmd/direct_use_en0@dev`

## Platform constraint

macOS-only. `en0_dialer.go` uses `unix.IP_BOUND_IF` / `unix.IPV6_BOUND_IF` (darwin-only socket options). Do not try to make this build on Linux without reworking the dialer.

## Architecture

- Root package `direct_use_en0`: shared dialer (`en0_dialer.go`), constants and the listen-address helper (`const.go`). Proxy listens on `127.0.0.1` only.
- `cmd/direct_use_en0`: flag parsing; entrypoint. Default `proxyType` is `mixed`.
- `internal/mixed_proxy`: mixed-mode server (default). Owns one TCP listener and dispatches per connection by the first byte: `0x05` → SOCKS5, anything else → HTTP (via a one-shot `net.Listener` + `http.Server.Serve`). Also starts the SOCKS5 UDP relay on the same port.
- `internal/txthinking_socks5`: the active SOCKS5 server (TCP + UDP). Works by overriding the package-level `socks5.DialTCP` / `socks5.DialUDP` globals — a process-wide side effect. Exports `NewServer` (server + dialer setup, shared with mixed) and `ListenAndServeUDP` (UDP relay without owning the TCP listener, used by mixed).
- `internal/http_proxy`: HTTP proxy, TCP only (no UDP). Binds via interface local IPs (`LocalAddr`), not `IP_BOUND_IF`. Exports `NewProxy`/`Handler` shared with mixed.
- `internal/armon_go_socks5` and `internal/things_go_go_socks5`: legacy/alternate implementations, not wired into `main.go` (armon's is fully commented out). Don't extend them; `txthinking_socks5` is the current implementation.
- `internal/check_ip`: backs the `-checkIp` flag; imports the txthinking socks5 package.

## Conventions

- Commit messages follow conventional-commit style with scope: `feat(socks5): ...`, `chore(netinterface.go): ...`, `style(cmd): ...`.
