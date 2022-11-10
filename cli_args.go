package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type argStringSlice []string

func (a *argStringSlice) Set(s string) error {
	*a = append(*a, s)
	return nil
}

func (a *argStringSlice) String() string {
	return strings.Join(*a, ",")
}

type appOptions struct {
	webSocketEchoURLs argStringSlice
	stunAddresses     argStringSlice
	bindAddress       string
	trace             bool
}

var flagSet *flag.FlagSet

const programName string = "tcp-hole-puncher"
const version string = "0.2.0"

var defaultBind = ":7203"
var defaultWsUrls = [...]string{
	"wss://ws.vi-server.org/mirror",
	"wss://ws.postman-echo.com/raw",
	"wss://ws.ifelse.io/",
	"wss://echo.websocket.events/",
}
var defaultStuns = [...]string{
	"stun.sipnet.net:3478",
	"stun.antisip.com:3478",
	"stun.voipgate.com:3478",
	"stun.nextcloud.com:443",
	"stun.ipfire.org:3478",
	"stun.freeswitch.org:3478",
	"stun.sonetel.net:3478",
}

func parseCliArgs() (args appOptions) {
	flagSet = flag.NewFlagSet(programName, flag.ContinueOnError)
	flagSet.BoolVar(&args.trace, "trace", false, "Trace output")

	flagSet.StringVar(&args.bindAddress, "bind", defaultBind, "Bind address")

	flagSet.Var(&args.webSocketEchoURLs, "ws-url", "Websocket echo server url (default servers if none provided)\n    \""+strings.Join(defaultWsUrls[:], "\"\n    \"")+"\"")
	flagSet.Var(&args.stunAddresses, "stun", "Stun server address (default servers if none provided)\n    \""+strings.Join(defaultStuns[:], "\"\n    \"")+"\"")

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
