package main

import (
	"net/http"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/agent"
)

func main() {
	agent := agent.NewAgent(&agent.Config{
		Client: &http.Client{
			Timeout: 2 * time.Second,
		},
		URL:            "http://localhost:8080/update",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	})

	agent.Launch()
}
