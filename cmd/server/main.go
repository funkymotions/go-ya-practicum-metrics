package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/server"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	options := env.ParseServerOptions()

	// HTTP server
	app := server.NewApp(options)
	sigChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Printf("%s", utils.GetAppMetaInfo(buildVersion, buildDate, buildCommit))

	// run server in a separate goroutine
	go func() {
		if err := app.Run(); err != nil {
			log.Printf("app launch error: %v\n", err)
			// report the error to the main goroutine instead of exiting here
			errChan <- err
		}
	}()

	select {
	case <-sigChan:
		// received OS signal, proceed to graceful shutdown
		log.Printf("Received termination signal, shutting down...\n")
	case <-errChan:
		// server reported an error, shut down gracefully
		log.Printf("Server error received, shutting down...\n")
	}

	app.Shutdown()
	log.Printf("Servers gracefully exited...\n")
}
