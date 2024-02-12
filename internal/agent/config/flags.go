package config

import (
	"flag"
)

func (addr *ServerAddr) String() string {
	return addr.Host + addr.Port
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

var addr = ServerAddr{"localhost", "8080"}

func parseFlags() {
	flag.Var(&addr, "a", "server address to run on")
	flag.IntVar(&ReportInterval, "r", 10, "report interval in second to post metric values on server")
	flag.IntVar(&PollInterval, "p", 2, "metric values refreshing interval in second")

	flag.Parse()

	ListenAddr = "http://" + addr.Host + ":" + addr.Port
}
