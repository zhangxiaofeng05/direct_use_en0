## direct use en0

look at the routing table to see if the default gateway is en0
```
netstat -rn
```

add ipv4 route to the routing table to use en0 as the default gateway
```
sudo route -n add -host ipv4 192.168.1.1
```
replace ipv4 with your ipv4 address
replace 192.168.1.1 with your default gateway

```
sudo route -n add -inet6 ipv6 2001:db8::1%en0
```
replace ipv6 with your ipv6 address
replace 2001:db8::1%en0 with your default gateway

look at the routing table to see if the ipv4 route is added successfully
```
route -n get ipv4
```

look at the routing table to see if the ipv6 route is added successfully
```
route -n get -inet6 ipv6
```

delete the ipv4 route from the routing table
```
sudo route delete -host ipv4
```

delete the ipv6 route from the routing table
```
sudo route delete -inet6 ipv6
```
