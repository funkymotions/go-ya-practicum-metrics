package main

import (
	"crypto/rsa"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/agent"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/logger"
	"github.com/funkymotions/go-ya-practicum-metrics/internal/utils"
	"go.uber.org/zap"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	l, err := logger.NewLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	if err != nil {
		log.Fatal("failed to create logger")
	}
	maxRetrySendCount := 3
	options := env.ParseAgentOptions()
	var pubKey *rsa.PublicKey
	if *options.CryptoKey != "" {
		pubKey, err = utils.ReadRSAPublicKeyFromFile(*options.CryptoKey)
		if err != nil {
			log.Fatalf("failed to read RSA public key: %v", err)
		}
	}
	agent := agent.NewAgent(&agent.Config{
		Logger: l,
		PubKey: pubKey,
		MetricURL: url.URL{
			Scheme: "http",
			Host:   *options.Endpoint,
			Path:   "/updates/",
		},
		Client: &http.Client{
			Timeout: 200 * time.Millisecond,
		},
		PollInterval:   time.Duration(*options.PollInterval) * time.Second,
		ReportInterval: time.Duration(*options.ReportInterval) * time.Second,
		MaxRetries:     maxRetrySendCount,
		RateLimit:      *options.RateLimit,
		Hashing: struct {
			Key        *string
			HeaderName string
		}{
			Key:        options.Key,
			HeaderName: "hashsha256",
		},
	})

	log.Printf("%s", utils.GetAppMetaInfo(buildVersion, buildDate, buildCommit))
	agent.Launch()
}
