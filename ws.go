package main

import (
	"context"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"golang.org/x/net/websocket"
)

func webSocketLoop(ctx context.Context, dialer *net.Dialer, webSocketEchoURL string, connectTimeCh chan<- time.Time, writeTimeCh chan<- time.Time) {
	tick := time.Tick(4 * time.Second)
	for {
		runConnection(ctx, dialer, webSocketEchoURL, connectTimeCh, writeTimeCh)
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-tick:
		}
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
			log.Println("Connection", webSocketEchoURL, "context expired", myCtx.Err())
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
