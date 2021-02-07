package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	"golang.org/x/net/websocket"

	"github.com/libp2p/go-reuseport"

	"gortc.io/stun"
)

func main() {
	args := parseCliArgs()
	if !args.trace {
		log.SetOutput(ioutil.Discard)
	}
	if !reuseport.Available() {
		log.Println("Port reuse not available")
		return
	}
	bindAddr, err := net.ResolveTCPAddr("tcp", args.bindAddress)
	if err != nil {
		log.Println("Can't resolve bind address", args.bindAddress, err)
		return
	}
	dialer := &net.Dialer{
		Control:   reuseport.Control,
		LocalAddr: bindAddr,
	}
	ctx := context.Background()
	tick := time.Tick(time.Second)
	for {
		runConnection(ctx, dialer, args.webSocketEchoURL, args.stunAddress)
		<-tick
	}
}

func runConnection(ctx context.Context, dialer *net.Dialer, webSocketEchoURL string, stunAddress string) {
	myCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	config, err := websocket.NewConfig(webSocketEchoURL, webSocketEchoURL)
	if err != nil {
		log.Panicln("Can't create config")
		return
	}
	config.Dialer = dialer
	log.Println("Connecting to websocket")
	conn, err := websocket.DialConfig(config)
	if err != nil {
		log.Println("Can't connect websocket", err)
		return
	}
	defer conn.Close()
	go printAddress(myCtx, dialer, stunAddress)
	go io.Copy(ioutil.Discard, conn)
	tick := time.Tick(time.Second * 5)
	for {
		select {
		case <-myCtx.Done():
			log.Println("Run connection context expired", myCtx.Err())
			return
		case <-tick:
			break
		}
		_, err = conn.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
		if err != nil {
			log.Println("Can't write to websocket", err)
			return
		}
	}
}

func printAddress(ctx context.Context, dialer *net.Dialer, stunAddress string) {
	addr, err := getAddress(ctx, dialer, stunAddress)
	if err != nil {
		log.Println("Can't get address", err)
		return
	}
	fmt.Println(addr)
}

func getAddress(ctx context.Context, dialer *net.Dialer, stunAddress string) (addr net.Addr, err error) {
	log.Println("Connecting to stun")
	conn, err := dialer.DialContext(ctx, "tcp", stunAddress)
	if err != nil {
		log.Println("Can't connect stun", err)
		return
	}
	client, err := stun.NewClient(conn)
	if err != nil {
		log.Println("Can't create stun")
		return
	}
	defer client.Close()
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	log.Println("Doing stun request")
	err2 := client.Do(message, func(res stun.Event) {
		if res.Error != nil {
			err = res.Error
			log.Println("Stun event error", err)
			return
		}
		var xorAddr stun.XORMappedAddress
		if suberr := xorAddr.GetFrom(res.Message); err != nil {
			err = suberr
			log.Println("Can't get address from stun message", err)
			return
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
			log.Println("Stun request context expired", err)
			return
		default:
			break
		}
		addr = &net.TCPAddr{
			IP:   xorAddr.IP,
			Port: xorAddr.Port,
		}
	})
	if err2 != nil {
		err = err2
		log.Println("Stun request error", err)
		return
	}
	return
}
