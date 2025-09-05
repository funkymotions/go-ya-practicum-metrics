package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/server"
)

func main() {
	options := env.ParseServerOptions()
	s := server.NewServer(options)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// run server in a separate goroutine to controll graceful shutdown
	go func() {
		if err := s.Run(); err != nil {
			log.Printf("Server launch error: %v\n", err)
			os.Exit(1)
		}
	}()
	<-sigChan
	s.Shutdown()
}
