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
	hp := strings.Split(s, ":")

	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}

	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}

	addr.Host = hp[0]
	addr.Port = port

	return nil
}

var addr = ServerAddr{"localhost", 8080}

func parseFlags() {
	flag.Var(&addr, "a", "server address to run on")
	flag.Parse()
}
