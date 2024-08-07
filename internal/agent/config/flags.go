package config

import (
	"flag"
	"log"
)

func parseFlags(args []string) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.StringVar(&Options.Address, "a", hostDefault+":"+portDefault, "server address to post metric values")
	fs.IntVar(&Options.ReportInterval, "r", 10, "report interval in second to post metric values on server")
	fs.IntVar(&Options.PollInterval, "p", 2, "metric values refreshing interval in second")
	fs.StringVar(&Options.Key, "k", "", "secret key to sign request")
	fs.IntVar(&Options.RateLimit, "l", 0, "limit sequential requests to server")

	err := fs.Parse(args)
	if err != nil {
		log.Print(err)
	}
}
