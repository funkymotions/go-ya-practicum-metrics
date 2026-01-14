package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
)

type AuditType uint

const (
	AuditFileType = iota
	AuditServiceType
)

type AuditObserver interface {
	emit(AuditMessage, callback)
	GetID() string
	GetType() AuditType
}

type auditService struct {
	observers map[string]AuditObserver
	filePath  string
	remoteURL string
	events    chan event
	stopChan  <-chan struct{}
	doneChan  chan<- struct{}
	wg        sync.WaitGroup
}

type event struct {
	metrics   any
	ipAddress string
}

type AuditMessage struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Subscriber struct {
	ID   string
	Type AuditType
}

type callback func(m AuditMessage)

func (s *Subscriber) emit(m AuditMessage, cb callback) {
	cb(m)
}

func (s *Subscriber) GetID() string {
	return s.ID
}

func (s *Subscriber) GetType() AuditType {
	return s.Type
}

// can be extendes to pass ...Subscriber args to register multiple observers due to initialization
func NewAuditService(
	filePath string, remoteURL string,
	stopCh <-chan struct{},
	doneCh chan<- struct{},
) *auditService {
	service := new(auditService)
	service.observers = make(map[string]AuditObserver)
	if filePath != "" {
		sub := Subscriber{
			ID:   strconv.FormatInt(int64(AuditFileType), 10),
			Type: AuditFileType,
		}
		service.observers[sub.ID] = &sub
		service.filePath = filePath
	}
	if remoteURL != "" {
		sub := Subscriber{
			ID:   strconv.FormatInt(int64(AuditServiceType), 10),
			Type: AuditServiceType,
		}
		service.observers[sub.ID] = &sub
		service.remoteURL = remoteURL
	}

	// setup worker to process audit events
	service.events = make(chan event, 100)
	service.doneChan = doneCh
	service.stopChan = stopCh
	go service.run()

	return service
}

// Due to metrics can be array or object, using interface{} here to cover both cases
// to suppoet notify both for single and multi updates
// extractMetricNames will handle the extraction of metric names appropriately
func (s *auditService) Notify(metrics any, ipAddress string) {
	select {
	case s.events <- event{metrics: metrics, ipAddress: ipAddress}:
	default:
		// If the channel is full, we can choose to drop the event or log it
		fmt.Println("Audit event channel is full, dropping event")
	}
}

func (s *auditService) run() {
	for {
		select {
		case ev := <-s.events:
			// dispatch a new event and write down a wg
			s.wg.Add(1)
			s.dispatch(ev.metrics, ev.ipAddress)
		case <-s.stopChan:
			// wait for all existing dispatches to finish
			s.wg.Wait()
			s.doneChan <- struct{}{}
			close(s.doneChan)
			return
		}
	}
}

func (s *auditService) dispatch(rawMessage any, ipAddress string) {
	defer s.wg.Done()
	for _, observer := range s.observers {
		metricNames := extractMetricNames(rawMessage)
		m := AuditMessage{
			Timestamp: time.Now().Unix(),
			Metrics:   metricNames,
			IPAddress: ipAddress,
		}
		if observer.GetType() == AuditFileType {
			observer.emit(m, s.writeEvent)
		}
		if observer.GetType() == AuditServiceType {
			observer.emit(m, s.sendEvent)
		}
	}
}

func (s *auditService) writeEvent(m AuditMessage) {
	f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	event, _ := json.Marshal(m)
	if _, err := f.WriteString(string(event) + "\n"); err != nil {
		fmt.Printf("Error writing to audit file: %s\n", err)
	}
}

func (s *auditService) sendEvent(m AuditMessage) {
	data, _ := json.Marshal(m)
	response, err := http.Post(s.remoteURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer response.Body.Close()
}

func extractMetricNames(metrics any) []string {
	metricNames := make([]string, 0)
	switch v := metrics.(type) {
	case models.Metrics:
		metricNames = append(metricNames, v.ID)
	case []models.Metrics:
		for _, item := range v {
			metricNames = append(metricNames, item.ID)
		}
	}
	return metricNames
}
