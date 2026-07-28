## direct use en0
tcp proxy, can configure a front proxy for v2rayn.

only support ipv4.

### install
```bash
go install github.com/zhangxiaofeng05/direct_use_en0/cmd/direct_use_en0@dev
```

### run
```bash
direct_use_en0 -port 20808
```

look at the routing table to see if the default gateway is en0
```bash
netstat -rn
```


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
