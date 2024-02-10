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

var addr = ServerAddr{"localhost", 8080}

func parseFlags() {
	flag.Var(&addr, "a", "server address to run on")
	flag.Parse()
}

func (addr *ServerAddr) String() string {
	return fmt.Sprintf("%s:%d", addr.Host, addr.Port)
}

func (addr *ServerAddr) Set(s string) error {
	parsedAddr, err := parseServerAddr(s)
	if err != nil {
		return err
	}

	addr.Host = parsedAddr.Host
	addr.Port = parsedAddr.Port

	return nil
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
		Host: fmt.Sprintf("%s", hp[0]),
		Port: port,
	}, nil
}
