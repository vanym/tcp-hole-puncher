# tcp-hole-puncher

tcp-hole-puncher is a tool to get connections behind nat

## Usage

### Run tcp-hole-puncher

```
tcp-hole-puncher --bind ":7203"
```
You will get address:port in stdout

### Redirect incoming tcp connections using iptables

```
iptables -t nat -A PREROUTING -p tcp --dport 7203 -j REDIRECT --to-ports 8080
```
Make sure that dport is opened in your router nat
