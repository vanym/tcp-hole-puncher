package main

import (
	"flag"
	"fmt"
	"os"
)

type appOptions struct {
	webSocketEchoURL string
	stunAddress      string
	bindAddress      string
	trace            bool
}

var flagSet *flag.FlagSet

const programName string = "tcp-hole-puncher"
const version string = "0.1.3"

func parseCliArgs() (args appOptions) {
	flagSet = flag.NewFlagSet(programName, flag.ContinueOnError)
	flagSet.BoolVar(&args.trace, "trace", false, "Trace output")

	flagSet.StringVar(&args.webSocketEchoURL, "ws-url", "wss://ws.vi-server.org/mirror", "Websocket echo server url")
	flagSet.StringVar(&args.bindAddress, "bind-address", ":7203", "Bind address")
	flagSet.StringVar(&args.stunAddress, "stun-address", "stun.sipnet.net:3478", "Stun server address")

	pVersion := flagSet.Bool("version", false, "Print version")

	err := flagSet.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	if err != nil {
		fmt.Printf("Invalid command line: %v. Try '%s -help'.\n", os.Args, programName)
		os.Exit(1)
	}
	if *pVersion {
		fmt.Println(programName, version)
		os.Exit(0)
	}
	return args
}
