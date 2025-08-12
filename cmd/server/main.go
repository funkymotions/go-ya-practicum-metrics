package main

import (
	"log"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/server"
)

func main() {
	parseFlags()
	err := server.NewServer(endpoint).Run()
	if err != nil {
		log.Printf("Server launch error: %v\n", err)
		panic(err)
	}
}
