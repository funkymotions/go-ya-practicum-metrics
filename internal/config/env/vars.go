package env

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Variables struct {
	Endpoint        *string `env:"ADDRESS"`
	ReportInterval  *uint   `env:"REPORT_INTERVAL"`
	PollInterval    *uint   `env:"POLL_INTERVAL"`
	StoreInterval   *uint   `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
}

func ParseAgentOptions() *Variables {
	var envVars Variables
	var endpointFlag = &Endpoint{Hostname: "localhost", Port: 8080}
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
		Endpoint: func() *string {
			if envVars.Endpoint != nil {
				return envVars.Endpoint
			}
			res := endpointFlag.String()
			return &res
		}(),
		ReportInterval: func() *uint {
			if envVars.ReportInterval != nil {
				return envVars.ReportInterval
			}
			return reportInterval
		}(),
		PollInterval: func() *uint {
			if envVars.PollInterval != nil {
				return envVars.PollInterval
			}
			return pollInterval
		}(),
	}
}

func ParseServerOptions() *Variables {
	var envVars Variables
	var endpointFlag = &Endpoint{Hostname: "localhost", Port: 8080}
	var storeInterval = new(uint)
	var fileStoragePath = new(string)
	var restore = new(bool)
	if err := env.Parse(&envVars); err != nil {
		log.Fatal(err)
	}
	flag.UintVar(storeInterval, "i", 300, "set store interval (seconds)")
	flag.StringVar(fileStoragePath, "f", "tmp/metrics-db.json", "set file storage path")
	flag.BoolVar(restore, "r", false, "set restore")
	flag.Var(endpointFlag, "a", "set endpoint (host:port)")
	flag.Parse()
	return &Variables{
		Endpoint: func() *string {
			if envVars.Endpoint != nil {
				return envVars.Endpoint
			}
			res := endpointFlag.String()
			return &res
		}(),
		StoreInterval: func() *uint {
			if envVars.StoreInterval != nil {
				return envVars.StoreInterval
			}
			return storeInterval
		}(),
		FileStoragePath: func() *string {
			if envVars.FileStoragePath != nil {
				return envVars.FileStoragePath
			}
			return fileStoragePath
		}(),
		Restore: func() *bool {
			if envVars.Restore != nil {
				return envVars.Restore
			}
			return restore
		}(),
	}
}
