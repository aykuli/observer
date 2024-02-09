package main

import "flag"

var addr string

func parseFlags() {
	flag.StringVar(&addr, "a", ":8080", "address to run server on")
	flag.Parse()
}
