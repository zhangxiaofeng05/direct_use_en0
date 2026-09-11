## direct use en0
support ipv4 and ipv6. both tcp and udp are supported.

can configure a front proxy for v2rayn or add routes to the routing table.


### install
```bash
go install github.com/zhangxiaofeng05/direct_use_en0/cmd/direct_use_en0@dev
```

### run
```bash
direct_use_en0 -port 20808
```

`-proxyType` supports `mixed` (default), `socks5` and `http`. `mixed` serves
SOCKS5 and HTTP proxy clients on the same port, the protocol is detected
per connection. `-proxyType socks5` or `-proxyType http` run a single
protocol server.

```bash
curl -x socks5://127.0.0.1:20808 https://api.ip.sb/ip
curl -x http://127.0.0.1:20808 https://api.ip.sb/ip
```
SOCKS5 is recommended.

### terminal use proxy
```bash
# function to enable terminal proxy
function set_proxy() {
  # HTTPS proxy (recommended)
  # export HTTPS_PROXY=https://proxy.example.com:8080
  # HTTP proxy (if HTTPS not available)
  # export HTTP_PROXY=http://proxy.example.com:8080

  # export HTTP_PROXY=http://127.0.0.1:20808
  export HTTP_PROXY=socks5://127.0.0.1:20808
  export HTTPS_PROXY=$HTTP_PROXY
  export ALL_PROXY=$HTTP_PROXY
  echo -e "Proxy is ON"
}
# function to disable the terminal proxy
function unset_proxy(){
    unset HTTP_PROXY HTTPS_PROXY ALL_PROXY
    echo -e "Proxy is OFF"
}
# Bypass proxy for local server (required)
export NO_PROXY=localhost,127.0.0.1,::1
```

look at the routing table to see if the default gateway is en0
```bash
netstat -rn
```

### add routes to the routing table
add ipv4 route to the routing table to use en0 as the default gateway
```bash
sudo route -n add -host ipv4 192.168.1.1
```
replace ipv4 with your ipv4 address
replace 192.168.1.1 with your default gateway


add ipv6 route to the routing table to use en0 as the default gateway
```bash
sudo route -n add -inet6 ipv6 2001:db8::1%en0
```
replace ipv6 with your ipv6 address
replace 2001:db8::1%en0 with your default gateway


look at the routing table to see if the ipv4 route is added successfully
```bash
route -n get ipv4
```
replace ipv4 with your ipv4 address


look at the routing table to see if the ipv6 route is added successfully
```bash
route -n get -inet6 ipv6
```
replace ipv6 with your ipv6 address


delete the ipv4 route from the routing table
```bash
sudo route delete -host ipv4
```
replace ipv4 with your ipv4 address


delete the ipv6 route from the routing table
```bash
sudo route delete -inet6 ipv6
```
replace ipv6 with your ipv6 address
