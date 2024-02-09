package main

import (
	"flag"
	"time"
)

var addr string
var reportInterval, pollInterval time.Duration

func parseFlags() {
	flag.StringVar(&addr, "a", "8080", "server address to run on")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "report interval in second to post metric values on server")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "metric values refreshing interval in second")

	flag.Parse()
}
