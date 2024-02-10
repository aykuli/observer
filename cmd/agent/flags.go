package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type ServerAddr struct {
	Host string
	Port int
}

func (addr *ServerAddr) String() string {
	return fmt.Sprintf("%s:%d", addr.Host, addr.Port)
}

func (addr *ServerAddr) Set(s string) error {
	serverAddr, err := parseServerAddr(s)

	if err != nil {
		return err
	}

	addr.Host = serverAddr.Host
	addr.Port = serverAddr.Port
	return nil
}

var addr = ServerAddr{"http://localhost", 8080}
var reportInterval, pollInterval int

func parseFlags() {
	flag.Var(&addr, "a", "server address to run on")
	flag.IntVar(&reportInterval, "r", 10, "report interval in second to post metric values on server")
	flag.IntVar(&pollInterval, "p", 2, "metric values refreshing interval in second")

	flag.Parse()
	listenAddr = fmt.Sprintf("%s:%v", addr.Host, addr.Port)
}

func parseServerAddr(s string) (ServerAddr, error) {
	hp := strings.Split(s, ":")

	if len(hp) != 2 {
		return ServerAddr{}, errors.New("need address in a form host:port")
	}

	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return ServerAddr{}, err
	}

	return ServerAddr{
		Host: fmt.Sprintf("http://%s", hp[0]),
		Port: port,
	}, nil
}
