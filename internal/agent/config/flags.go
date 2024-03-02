package config

import (
	"flag"
	"log"
	"os"
)

var addr = ServerAddr{hostDefault, portDefault}

func parseFlags() {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.Var(&addr, "a", "server address to run on")
	fs.IntVar(&Options.ReportInterval, "r", 10, "report interval in second to post metric values on server")
	fs.IntVar(&Options.PollInterval, "p", 2, "metric values refreshing interval in second")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Print(err)
	}

	Options.Address = addr.String()
}

func (addr *ServerAddr) String() string {
	return addr.Host + ":" + addr.Port
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
