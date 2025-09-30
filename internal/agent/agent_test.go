package agent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

func float64Ptr(i int) *float64 {
	f := float64(i)
	return &f
}

func int64Ptr(i int) *int64 {
	i64 := int64(i)
	return &i64
}

func TestNewAgent(t *testing.T) {
	endpoint := "localhost:8080"
	env := &env.Variables{Endpoint: &endpoint}
	metricURL := url.URL{
		Scheme: "http",
		Host:   *env.Endpoint,
		Path:   "/metric",
	}
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name string
		args args
		want *agent
	}{
		{
			name: "should return new agent",
			args: args{
				cfg: &Config{
					MetricURL: metricURL,
				},
			},
			want: &agent{
				config: &Config{
					MetricURL: metricURL,
				},
				metrics: make(map[string]models.Metrics),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := NewAgent(tt.args.cfg)
			require.True(t, reflect.DeepEqual(expected, tt.want))
		})
	}
}

func Test_agent_sendMetrics(t *testing.T) {
	var isServerWasCalled atomic.Value
	isServerWasCalled.Store(false)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		isServerWasCalled.Store(true)
	}))
	testServerURL, _ := url.Parse(ts.URL)
	defer ts.Close()
	type fields struct {
		config  *Config
		metrics map[string]models.Metrics
	}
	type args struct {
		stop chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "should send metrics to server",
			fields: fields{
				metrics: map[string]models.Metrics{
					"test_metric_name": {
						ID:    "test_metric_name",
						MType: models.Gauge,
						Value: float64Ptr(42),
					},
					"test_counter": {
						ID:    "test_counter",
						MType: models.Counter,
						Delta: int64Ptr(100),
					},
				},
				config: &Config{
					Logger: zap.NewNop(),
					Client: &http.Client{},
					MetricURL: url.URL{
						Scheme: "http",
						Host:   testServerURL.Host,
						Path:   "/metric",
					},
					PollInterval:   50 * time.Millisecond,
					ReportInterval: 100 * time.Millisecond,
				},
			},
			args: args{
				stop: make(chan struct{}),
			},
			want: true,
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			m := &agent{
				config:  tt.fields.config,
				metrics: tt.fields.metrics,
			}
			go m.sendMetrics(tt.args.stop)
			time.Sleep(150 * time.Millisecond)
			require.Equal(t, tt.want, isServerWasCalled.Load().(bool))
		})
	}
}

func Test_agent_collectMetrics(t *testing.T) {
	type fields struct {
		URL     string
		config  *Config
		metrics map[string]models.Metrics
	}
	type args struct {
		stop chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "should collect metrics",
			fields: fields{
				metrics: make(map[string]models.Metrics),
				config: &Config{
					Logger:         zap.NewNop(),
					Client:         &http.Client{},
					PollInterval:   50 * time.Millisecond,
					ReportInterval: 100 * time.Second,
				},
			},
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			m := &agent{
				config:  tt.fields.config,
				metrics: tt.fields.metrics,
			}
			go m.collectMetrics(tt.args.stop)
			time.Sleep(100 * time.Millisecond)
			m.mu.Lock()
			result := len(m.metrics)
			m.mu.Unlock()
			require.NotEqual(t, result, 0, "Expected metrics to be collected")
		})
	}
}

type performRequestTestSuite struct {
	suite.Suite
	agent *agent
}

func Test_performRequestTestSuite(t *testing.T) {
	suite.Run(t, new(performRequestTestSuite))
}

func (s *performRequestTestSuite) SetupTest() {
	retries := 3
	s.agent = &agent{
		config: &Config{
			Logger: zap.NewNop(),
			Client: &http.Client{
				Timeout: 1 * time.Second,
			},
			PollInterval:   50 * time.Millisecond,
			ReportInterval: 100 * time.Millisecond,
			MaxRetries:     &retries,
		},
		metrics: map[string]models.Metrics{},
	}
}

func (s *performRequestTestSuite) Test_performRequest() {
	handerOK := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handlerErr := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name:    "should perform request successfully",
			handler: handerOK,
			wantErr: false,
		},
		{
			name:    "should fail to perform request",
			handler: handlerErr,
			wantErr: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			ts := httptest.NewServer(http.HandlerFunc(test.handler))
			err := s.agent.performRequest(ts.URL)
			if test.wantErr {
				s.Require().Error(err)
			} else {
				s.Assert().NoError(err)
			}
			ts.Close()
		})
	}
}
