package config

import (
	"flag"
	"log"
	"os"
)

func parseFlags() {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.StringVar(&Options.Address, "a", hostDefault+":"+portDefault, "report interval in second to post metric values on server")
	fs.IntVar(&Options.ReportInterval, "r", 10, "report interval in second to post metric values on server")
	fs.IntVar(&Options.PollInterval, "p", 2, "metric values refreshing interval in second")
	fs.StringVar(&Options.Key, "k", "", "secret key to sign request")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Print(err)
	}
}
