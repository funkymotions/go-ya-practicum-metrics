package main

import (
	"flag"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
)

var endpoint *env.Endpoint = &env.Endpoint{Hostname: "localhost", Port: 8080}
var reportInterval uint
var pollInterval uint

func parseFlags() {
	flag.Var(endpoint, "a", "set endpoint (host:port)")
	flag.UintVar(&reportInterval, "r", 2, "set report interval (seconds)")
	flag.UintVar(&pollInterval, "p", 10, "set poll interval (seconds)")
	flag.Parse()
}
