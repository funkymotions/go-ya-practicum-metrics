package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/agent"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {
	l, err := logger.NewLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	if err != nil {
		log.Fatal("failed to create logger")
	}
	options := env.ParseOptions()
	agent := agent.NewAgent(&agent.Config{
		Logger: l,
		MetricURL: url.URL{
			Scheme: "http",
			Host:   options.Endpoint,
			Path:   "/update/",
		},
		Client: &http.Client{
			Timeout: 200 * time.Millisecond,
		},
		PollInterval:   time.Duration(options.PollInterval) * time.Second,
		ReportInterval: time.Duration(options.ReportInterval) * time.Second,
	})
	agent.Launch()
}
