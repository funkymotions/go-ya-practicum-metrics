package env

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Variables struct {
	Endpoint       string `env:"ADDRESS"`
	ReportInterval uint   `env:"REPORT_INTERVAL"`
	PollInterval   uint   `env:"POLL_INTERVAL"`
}

func ParseOptions() *Variables {
	var envVars Variables
	var endpointFlag *Endpoint = &Endpoint{Hostname: "localhost", Port: 8080}
	var reportInterval = new(uint)
	var pollInterval = new(uint)
	if err := env.Parse(&envVars); err != nil {
		log.Fatal(err)
	}
	flag.Var(endpointFlag, "a", "set endpoint (host:port)")
	flag.UintVar(reportInterval, "r", 10, "set report interval (seconds)")
	flag.UintVar(pollInterval, "p", 2, "set poll interval (seconds)")
	flag.Parse()
	return &Variables{
		Endpoint: func() string {
			if envVars.Endpoint != "" {
				return envVars.Endpoint
			}
			return endpointFlag.String()
		}(),
		ReportInterval: func() uint {
			if envVars.ReportInterval != 0 {
				return envVars.ReportInterval
			}
			return *reportInterval
		}(),
		PollInterval: func() uint {
			if envVars.PollInterval != 0 {
				return envVars.PollInterval
			}
			return *pollInterval
		}(),
	}
}
