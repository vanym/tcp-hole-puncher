# tcp-hole-puncher

tcp-hole-puncher is a tool to get connections behind nat

## How it works

App bind port for outgoing connections, make connections to websocket mirror, get address from stun servers

Some websocket mirrors and stun servers used by default, also it can be configured by command line arguments

See `tcp-hole-puncher --help` for more info

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

## Installation
```
git clone https://github.com/vanym/tcp-hole-puncher.git
cd tcp-hole-puncher
make
sudo make install
```

## Usage with scripts

#### Run script

Run script sets up redirection, runs app and passes it output to handler

```
TCP_HOLE_PORT=7203 TCP_HOLE_REDIR_PORT=8080 ./run-hole-maker.sh
```

#### Start stop scripts

Start stop scripts uses `start-stop-daemon` to start and stop run script

```
TCP_HOLE_PORT=7203 TCP_HOLE_REDIR_PORT=8080 ./start-hole-maker.sh
./stop-hole-maker.sh
```

#### Handler script

Handler called by run script when new address appears

Default `handle-address.sh` script put address to address.txt file in script location directory

You can change handler by setting `TCP_HOLE_ADDRESS_HANDLER` environment variable

## Usage with docker

Configure ports as environment variables in `docker-compose.yml` file and start using

```
sudo docker-compose up -d
```

You can copy just `docker-compose.yml` without whole repository and use pre build image [`vanym/tcp-hole-puncher`](https://hub.docker.com/r/vanym/tcp-hole-puncher)
