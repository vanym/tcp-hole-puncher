package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/libp2p/go-reuseport"
)

func main() {
	args := parseCliArgs()
	if len(args.webSocketEchoURLs) == 0 {
		args.webSocketEchoURLs = defaultWsUrls[:]
	}
	if len(args.stunAddresses) == 0 {
		args.stunAddresses = defaultStuns[:]
	}
	if !args.trace {
		log.SetOutput(io.Discard)
	}
	if !reuseport.Available() {
		fmt.Fprintln(os.Stderr, "Port reuse not available")
		return
	}
	bindAddr, err := net.ResolveTCPAddr("tcp", args.bindAddress)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't resolve bind address", args.bindAddress, err)
		return
	}
	log.Println("Using address:", bindAddr)
	dialer := &net.Dialer{
		Control:   reuseport.Control,
		LocalAddr: bindAddr,
	}
	ctx := context.Background()
	nurls := len(args.webSocketEchoURLs)
	connectTimeCh := make(chan time.Time, nurls*16)
	writeTimeCh := make(chan time.Time)
	for _, webSocketEchoURL := range args.webSocketEchoURLs {
		go webSocketLoop(ctx, dialer, webSocketEchoURL, connectTimeCh, writeTimeCh)
	}
	stuns := make(stunsHolder, len(args.stunAddresses))
	copy(stuns, args.stunAddresses)
	var lastTime time.Time
	var stunerr error
	tick := time.Tick(time.Second)
	for {
		if stunerr == nil {
			conTime := <-connectTimeCh
			if !conTime.After(lastTime) {
				continue
			}
			time.Sleep(time.Second)
		} else {
			select {
			case <-writeTimeCh:
			case <-connectTimeCh:
			}
			<-tick
		}
		var addr net.Addr
		var checkTime time.Time
		addr, checkTime, stunerr = stuns.getAddress(ctx, dialer)
		if stunerr != nil {
			log.Println("Can't get address", stunerr)
			continue
		}
		lastTime = checkTime
		fmt.Println(addr)
	}
}
