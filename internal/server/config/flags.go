package config

import (
	"flag"
)

var addr = ServerAddr{"localhost", "8080"}

func parseFlags() {
	flag.Var(&addr, "a", "server address to run on")
	flag.Parse()

	ListenAddr = addr.Host + ":" + addr.Port
}

func (addr *ServerAddr) String() string {
	return addr.Host + addr.Port
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
