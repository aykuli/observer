package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
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

	addr.Host = fmt.Sprintf("http://%s", hp[0])
	addr.Port = port

	return nil
}

var addr = ServerAddr{"http://localhost", 8080}
var reportInterval, pollInterval time.Duration

func parseFlags() {
	flag.Var(&addr, "a", "server address to run on")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "report interval in second to post metric values on server")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "metric values refreshing interval in second")

	flag.Parse()
}
