package env

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
)

type GenericParameter interface {
	int | uint | string | bool
}

type agentConfigJSON struct {
	Endpoint       *string `json:"address"`
	ReportInterval *uint   `json:"report_interval"`
	PollInterval   *uint   `json:"poll_interval"`
	RateLimit      *int    `json:"rate_limit"`
	Key            *string `json:"key"`
	CryptoKey      *string `json:"crypto_key"`
}

type serverConfigJSON struct {
	Endpoint        *string `json:"address"`
	StoreInterval   *uint   `json:"store_interval"`
	FileStoragePath *string `json:"store_file"`
	Restore         *bool   `json:"restore"`
	DatabaseDSN     *string `json:"database_dsn"`
	Key             *string `json:"key"`
	AuditFile       *string `json:"audit_file"`
	AuditURL        *string `json:"audit_url"`
	CryptoKey       *string `json:"crypto_key"`
	TrustedSubnet   *string `json:"trusted_subnet"`
}

type Variables struct {
	Endpoint        *string `env:"ADDRESS"`
	ReportInterval  *uint   `env:"REPORT_INTERVAL"`
	PollInterval    *uint   `env:"POLL_INTERVAL"`
	StoreInterval   *uint   `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDSN     *string `env:"DATABASE_DSN"`
	Key             *string `env:"KEY"`
	RateLimit       *int    `env:"RATE_LIMIT"`
	AuditFile       *string `env:"AUDIT_FILE"`
	AuditURL        *string `env:"AUDIT_URL"`
	CryptoKey       *string `env:"CRYPTO_KEY"`
	ConfigFile      *string `env:"CONFIG"`
	TrustedSubnet   *string `env:"TRUSTED_SUBNET"`
}

func ParseAgentOptions() *Variables {
	var envVars Variables
	var endpointFlag = &Endpoint{Hostname: "localhost", Port: 8080}
	var reportInterval = new(uint)
	var pollInterval = new(uint)
	var key = new(string)
	var rateLimit = new(int)
	var cryptoKey = new(string)
	var configFlag = new(string)
	if err := env.Parse(&envVars); err != nil {
		log.Fatal(err)
	}

	flag.Var(endpointFlag, "a", "set endpoint (host:port)")
	flag.UintVar(reportInterval, "r", 10, "set report interval (seconds)")
	flag.UintVar(pollInterval, "p", 2, "set poll interval (seconds)")
	flag.StringVar(key, "k", "", "set key used for hashing")
	flag.IntVar(rateLimit, "l", 0, "set rate limit (requests per second), 0 means no limit")
	flag.StringVar(cryptoKey, "crypto-key", "", "RSA public key path")
	flag.StringVar(configFlag, "config", "", "set path to config file")
	flag.Parse()

	var configJSON agentConfigJSON
	if *configFlag != "" {
		file, err := os.ReadFile(*configFlag)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal(file, &configJSON)
	}

	e := endpointFlag.String()
	endpointParam := chooseParameter(envVars.Endpoint, &e, configJSON.Endpoint)
	reportIntervalParam := chooseParameter(envVars.ReportInterval, reportInterval, configJSON.ReportInterval)
	pollIntervalParam := chooseParameter(envVars.PollInterval, pollInterval, configJSON.PollInterval)
	keyParam := chooseParameter(envVars.Key, key, configJSON.Key)
	rateLimitParam := chooseParameter(envVars.RateLimit, rateLimit, configJSON.RateLimit)
	cryptoKeyParam := chooseParameter(envVars.CryptoKey, cryptoKey, configJSON.CryptoKey)

	return &Variables{
		Endpoint:       &endpointParam,
		ReportInterval: &reportIntervalParam,
		PollInterval:   &pollIntervalParam,
		Key:            &keyParam,
		RateLimit:      &rateLimitParam,
		CryptoKey:      &cryptoKeyParam,
	}
}

func ParseServerOptions() *Variables {
	var envVars Variables
	var endpointFlag = &Endpoint{Hostname: "localhost", Port: 8080}
	var storeInterval = new(uint)
	var fileStoragePath = new(string)
	var restore = new(bool)
	var dsn = new(string)
	var key = new(string)
	var auditFile = new(string)
	var auditURL = new(string)
	var cryptoKey = new(string)
	var configFlag = new(string)
	var trustedSubnet = new(string)
	if err := env.Parse(&envVars); err != nil {
		log.Fatal(err)
	}

	flag.UintVar(storeInterval, "i", 300, "set store interval (seconds)")
	flag.StringVar(fileStoragePath, "f", "tmp/metrics-db.json", "set file storage path")
	flag.BoolVar(restore, "r", false, "set restore")
	flag.Var(endpointFlag, "a", "set endpoint (host:port)")
	flag.StringVar(dsn, "d", "", "set database dsn")
	flag.StringVar(key, "k", "", "set key used for hashing")
	flag.StringVar(auditFile, "audit-file", "", "set audit file path")
	flag.StringVar(auditURL, "audit-url", "", "set audit service URL")
	flag.StringVar(cryptoKey, "crypto-key", "", "RSA private key path")
	flag.StringVar(configFlag, "config", "", "set path to config file")
	flag.StringVar(trustedSubnet, "t", "", "set trusted subnet mask")
	flag.Parse()

	var configJSON serverConfigJSON
	if *configFlag != "" {
		file, err := os.ReadFile(*configFlag)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal(file, &configJSON)
	}

	e := endpointFlag.String()
	endpointParam := chooseParameter(envVars.Endpoint, &e, configJSON.Endpoint)
	storeIntervalParam := chooseParameter(envVars.StoreInterval, storeInterval, configJSON.StoreInterval)
	fileStoragePathParam := chooseParameter(envVars.FileStoragePath, fileStoragePath, configJSON.FileStoragePath)
	restoreParam := chooseParameter(envVars.Restore, restore, configJSON.Restore)
	databaseDSNParam := chooseParameter(envVars.DatabaseDSN, dsn, configJSON.DatabaseDSN)
	keyParam := chooseParameter(envVars.Key, key, configJSON.Key)
	auditFileParam := chooseParameter(envVars.AuditFile, auditFile, configJSON.AuditFile)
	auditURLParam := chooseParameter(envVars.AuditURL, auditURL, configJSON.AuditURL)
	cryptoKeyParam := chooseParameter(envVars.CryptoKey, cryptoKey, configJSON.CryptoKey)
	trustedSubnetParam := chooseParameter(envVars.TrustedSubnet, trustedSubnet, configJSON.TrustedSubnet)

	return &Variables{
		Endpoint:        &endpointParam,
		StoreInterval:   &storeIntervalParam,
		FileStoragePath: &fileStoragePathParam,
		Restore:         &restoreParam,
		DatabaseDSN:     &databaseDSNParam,
		Key:             &keyParam,
		AuditFile:       &auditFileParam,
		AuditURL:        &auditURLParam,
		CryptoKey:       &cryptoKeyParam,
		TrustedSubnet:   &trustedSubnetParam,
	}
}

// chooseParameter selects the marameter valuse based on the priority:
// environment variable > command-line flag > config file > zero value
// cmd flag pointer always non-nil but may point to zero value of T or a default value,
// so we need to check for that case explicitly
func chooseParameter[T GenericParameter](envVal *T, flagVal *T, jsonVal *T) T {
	var zero T
	switch {
	case envVal != nil:
		return *envVal
	case flagVal != nil && *flagVal != zero:
		return *flagVal
	case jsonVal != nil:
		return *jsonVal
	default:
		return zero
	}
}
