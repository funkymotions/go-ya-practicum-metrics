package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/agent"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
)

func main() {
	options := env.ParseOptions()
	agent := agent.NewAgent(&agent.Config{
		GaugeURL: url.URL{
			Scheme: "http",
			Host:   options.Endpoint,
			Path:   "/update/gauge",
		},
		CounterURL: url.URL{
			Scheme: "http",
			Host:   options.Endpoint,
			Path:   "/update/counter",
		},
		Client: &http.Client{
			Timeout: 2 * time.Second,
		},
		PollInterval:   time.Duration(options.PollInterval) * time.Second,
		ReportInterval: time.Duration(options.ReportInterval) * time.Second,
	})
	agent.Launch()
}
