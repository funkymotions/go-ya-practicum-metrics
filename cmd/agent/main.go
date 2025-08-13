package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/agent"
)

func main() {
	parseFlags()
	agent := agent.NewAgent(&agent.Config{
		GaugeURL: url.URL{
			Scheme: "http",
			Host:   endpoint.String(),
			Path:   "/update/gauge",
		},
		CounterURL: url.URL{
			Scheme: "http",
			Host:   endpoint.String(),
			Path:   "/update/counter",
		},
		Client: &http.Client{
			Timeout: 2 * time.Second,
		},
		PollInterval:   time.Duration(pollInterval) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
	})
	agent.Launch()
}
