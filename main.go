package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"

	"github.com/libp2p/go-reuseport"

	"github.com/pion/stun"
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

func webSocketLoop(ctx context.Context, dialer *net.Dialer, webSocketEchoURL string, connectTimeCh chan<- time.Time, writeTimeCh chan<- time.Time) {
	tick := time.Tick(4 * time.Second)
	for {
		runConnection(ctx, dialer, webSocketEchoURL, connectTimeCh, writeTimeCh)
		<-tick
	}
}

func runConnection(ctx context.Context, dialer *net.Dialer, webSocketEchoURL string, connectTimeCh chan<- time.Time, writeTimeCh chan<- time.Time) {
	myCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	config, err := websocket.NewConfig(webSocketEchoURL, webSocketEchoURL)
	if err != nil {
		log.Panicln("Can't create config")
		return
	}
	config.Dialer = dialer
	log.Println("Connecting to websocket", webSocketEchoURL)
	conn, err := websocket.DialConfig(config)
	if err != nil {
		log.Println("Can't connect websocket "+webSocketEchoURL, err)
		return
	}
	defer conn.Close()
	connectTimeCh <- time.Now()
	log.Println("Connected to websocket", webSocketEchoURL)
	go io.Copy(io.Discard, conn)
	tick := time.Tick(time.Second * 5)
	for {
		select {
		case <-myCtx.Done():
			log.Println("Run connection context expired", myCtx.Err())
			return
		case <-tick:
			break
		}
		_, err = conn.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10) + "\n"))
		if err != nil {
			log.Println("Can't write to websocket", err)
			return
		}
		select {
		case writeTimeCh <- time.Now():
		default:
		}
	}
}

type stunsHolder []string

func (h *stunsHolder) getAddress(ctx context.Context, dialer *net.Dialer) (net.Addr, time.Time, error) {
	n := len(*h)
	if n == 0 {
		log.Panicln("empty stuns")
		return nil, time.Now(), io.ErrNoProgress
	} else if n == 1 {
		only := (*h)[0]
		addr, err := getAddress(ctx, dialer, only)
		t := time.Now()
		return addr, t, err
	}
	var current atomic.Int32
	const th = 2
	var wg sync.WaitGroup
	var mu sync.Mutex
	var addrs []net.Addr
	bads := make(map[*string]struct{})
	var lastTime time.Time
	for range [th]int{} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var addr net.Addr
			var err error = io.ErrNoProgress
			var subLastTime time.Time
			var subBads []*string
			for err != nil {
				i := int(current.Add(1) - 1)
				if i >= n {
					break
				}
				stunAddress := (*h)[i]
				addr, err = getAddress(ctx, dialer, stunAddress)
				if err != nil {
					addr = nil
					subBads = append(subBads, &(*h)[i])

				} else {
					subLastTime = time.Now()
				}
			}
			mu.Lock()
			defer mu.Unlock()
			for _, v := range subBads {
				bads[v] = struct{}{}
			}
			if err != nil {
				return
			}
			lastTime = subLastTime
			for _, a := range addrs {
				if a.Network() == addr.Network() &&
					a.String() == addr.String() {
					return
				}
			}
			addrs = append(addrs, addr)
		}()
	}
	wg.Wait()
	if len(bads) > 0 {
		sort.Slice(*h, func(i, j int) bool {
			_, bi := bads[&(*h)[i]]
			_, bj := bads[&(*h)[j]]
			if bi == bj {
				return i < j
			} else {
				return bj
			}
		})
	}
	switch len(addrs) {
	default:
		{
			var sb strings.Builder
			sb.WriteString("too many results:")
			for _, addr := range addrs {
				sb.WriteString(" ")
				sb.WriteString(addr.String())
			}
			return nil, lastTime, errors.New(sb.String())
		}
	case 0:
		return nil, lastTime, errors.New("no results")
	case 1:
		return addrs[0], lastTime, nil
	}
}

func getAddress(ctx context.Context, dialer *net.Dialer, stunAddress string) (addr net.Addr, err error) {
	myCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	log.Println("Connecting to stun", stunAddress)
	conn, err := dialer.DialContext(myCtx, "tcp", stunAddress)
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
	log.Println("Doing stun request", stunAddress)
	err2 := client.Do(message, func(res stun.Event) {
		if res.Error != nil {
			err = res.Error
			log.Println("Stun event error", err)
			return
		}
		var xorAddr stun.XORMappedAddress
		if suberr := xorAddr.GetFrom(res.Message); suberr != nil {
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
