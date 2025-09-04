package main

import (
	"log"
	"os"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/server"
)

func main() {
	options := env.ParseServerOptions()
	err := server.NewServer(options).Run()
	if err != nil {
		log.Printf("Server launch error: %v\n", err)
		os.Exit(1)
	}
}
