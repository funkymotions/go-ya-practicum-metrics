package agent

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

type agent struct {
	config         *Config
	gaugeMetrics   map[string]interface{}
	counterMetrics map[string]interface{}
	mu             sync.Mutex
}

type Config struct {
	Client         *http.Client
	PollInterval   time.Duration
	ReportInterval time.Duration
	GaugeURL       url.URL
	CounterURL     url.URL
}

func NewAgent(cfg *Config) *agent {
	return &agent{
		config:         cfg,
		gaugeMetrics:   make(map[string]interface{}),
		counterMetrics: make(map[string]interface{}),
	}
}

func (m *agent) Launch() {
	log.Printf(
		"Starting agent with endpoints: %s\n%s\n",
		m.config.GaugeURL.String(),
		m.config.CounterURL.String(),
	)
	log.Printf("Report Interval: %s\n", m.config.ReportInterval.String())
	log.Printf("Poll Interval: %s\n", m.config.PollInterval.String())
	stop := make(chan struct{})
	go m.collectMetrics(stop)
	go m.sendMetrics(stop)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	close(stop)
}

func (m *agent) sendMetrics(stop chan struct{}) {
	ticker := time.NewTicker(m.config.ReportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Printf("Sending metrics...\n")
			m.mu.Lock()
			for name, value := range m.gaugeMetrics {
				gaugeURL := m.prepareURL(name, value, gauge)
				resp, err := m.config.Client.Post(
					gaugeURL,
					"text/plain",
					strings.NewReader(fmt.Sprintf("%v", value)),
				)
				if err != nil {
					log.Printf("Error sending gauge metric %s: %v\n", name, err)
					continue
				}
				resp.Body.Close()
			}

			for name, value := range m.counterMetrics {
				counterURL := m.prepareURL(name, value, counter)
				resp, err := m.config.Client.Post(
					counterURL,
					"text/plain",
					strings.NewReader(fmt.Sprintf("%v", value)),
				)
				if err != nil {
					log.Printf("Error sending counter metric %s: %v\n", name, err)
					continue
				}
				resp.Body.Close()
			}
			m.mu.Unlock()
		case <-stop:
			return
		}
	}
}

func (m *agent) prepareURL(name string, value interface{}, metricType string) string {
	val := fmt.Sprintf("%v", value)
	if metricType == gauge {
		return m.config.GaugeURL.JoinPath(name, val).String()
	}
	return m.config.CounterURL.JoinPath(name, val).String()
}

func (m *agent) collectMetrics(stop chan struct{}) {
	ticker := time.NewTicker(m.config.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Printf("Collecting metrics...\n")
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			m.mu.Lock()
			m.gaugeMetrics["Alloc"] = memStats.Alloc
			m.gaugeMetrics["BuckHashSys"] = memStats.BuckHashSys
			m.gaugeMetrics["Frees"] = memStats.Frees
			m.gaugeMetrics["GCCPUFraction"] = memStats.GCCPUFraction
			m.gaugeMetrics["GCSys"] = memStats.GCSys
			m.gaugeMetrics["HeapAlloc"] = memStats.HeapAlloc
			m.gaugeMetrics["HeapIdle"] = memStats.HeapIdle
			m.gaugeMetrics["HeapInuse"] = memStats.HeapInuse
			m.gaugeMetrics["HeapObjects"] = memStats.HeapObjects
			m.gaugeMetrics["HeapReleased"] = memStats.HeapReleased
			m.gaugeMetrics["HeapSys"] = memStats.HeapSys
			m.gaugeMetrics["LastGC"] = memStats.LastGC
			m.gaugeMetrics["Lookups"] = memStats.Lookups
			m.gaugeMetrics["MCacheInuse"] = memStats.MCacheInuse
			m.gaugeMetrics["MCacheSys"] = memStats.MCacheSys
			m.gaugeMetrics["MSpanInuse"] = memStats.MSpanInuse
			m.gaugeMetrics["MSpanSys"] = memStats.MSpanSys
			m.gaugeMetrics["Mallocs"] = memStats.Mallocs
			m.gaugeMetrics["NextGC"] = memStats.NextGC
			m.gaugeMetrics["NumForcedGC"] = memStats.NumForcedGC
			m.gaugeMetrics["NumGC"] = memStats.NumGC
			m.gaugeMetrics["OtherSys"] = memStats.OtherSys
			m.gaugeMetrics["PauseTotalNs"] = memStats.PauseTotalNs
			m.gaugeMetrics["StackInuse"] = memStats.StackInuse
			m.gaugeMetrics["StackSys"] = memStats.StackSys
			m.gaugeMetrics["Sys"] = memStats.Sys
			m.gaugeMetrics["TotalAlloc"] = memStats.TotalAlloc
			m.gaugeMetrics["NumGC"] = memStats.NumGC
			m.gaugeMetrics["RandomValue"] = rand.Intn(1000)
			val, ok := m.counterMetrics["PollCount"]
			if ok {
				m.counterMetrics["PollCount"] = val.(int) + 1
			} else {
				m.counterMetrics["PollCount"] = 1
			}
			m.mu.Unlock()
		case <-stop:
			return
		}
	}
}
