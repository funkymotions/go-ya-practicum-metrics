package main

import (
	"flag"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
)

var endpoint *env.Endpoint = &env.Endpoint{Hostname: "localhost", Port: 8080}

func parseFlags() {
	flag.Var(endpoint, "a", "set endpoint (host:port)")
	flag.Parse()
}
