package main

import (
	"net/http"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/agent"
)

func main() {
	parseFlags()
	agent := agent.NewAgent(&agent.Config{
		Endpoint: endpoint,
		Client: &http.Client{
			Timeout: 2 * time.Second,
		},
		PollInterval:   time.Duration(pollInterval) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
	})

	agent.Launch()
}
