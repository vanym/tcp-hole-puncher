# tcp-hole-puncher

tcp-hole-puncher is a tool to get connections behind nat

## Usage

### Run tcp-hole-puncher

```
tcp-hole-puncher --bind ":7203"
```
You will get address:port in stdout

### Redirect incoming tcp connections

#### Using iptables
```
iptables -t nat -A PREROUTING -p tcp --dport 7203 -j REDIRECT --to-ports 8080
```

#### Using socat
```
socat tcp-listen:7203,reuseaddr,reuseport,fork tcp-connect:127.0.0.1:8080
```

Make sure that port is opened in your router nat
