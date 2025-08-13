package agent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	env := &env.Endpoint{Hostname: "localhost", Port: 8080}
	gaugeURL := url.URL{
		Scheme: "http",
		Host:   env.String(),
		Path:   "/gauge",
	}
	counterURL := url.URL{
		Scheme: "http",
		Host:   env.String(),
		Path:   "/counter",
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
					GaugeURL:   gaugeURL,
					CounterURL: counterURL,
				},
			},
			want: &agent{
				config: &Config{
					GaugeURL:   gaugeURL,
					CounterURL: counterURL,
				},
				gaugeMetrics:   make(map[string]interface{}),
				counterMetrics: make(map[string]interface{}),
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
	var isServerWasCalled = false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		isServerWasCalled = true
	}))

	testServerURL, _ := url.Parse(ts.URL)
	defer ts.Close()
	type fields struct {
		config         *Config
		gaugeMetrics   map[string]interface{}
		counterMetrics map[string]interface{}
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
				gaugeMetrics:   map[string]interface{}{"test_metric_name": 42},
				counterMetrics: map[string]interface{}{"test_counter": 100},
				config: &Config{
					Client: &http.Client{},
					GaugeURL: url.URL{
						Scheme: "http",
						Host:   testServerURL.Host,
						Path:   "/update/gauge",
					},
					CounterURL: url.URL{
						Scheme: "http",
						Host:   testServerURL.Host,
						Path:   "/update/counter",
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
				config:         tt.fields.config,
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
			}
			go m.sendMetrics(tt.args.stop)
			time.Sleep(150 * time.Millisecond)
			require.Equal(t, tt.want, isServerWasCalled)
		})
	}
}

func Test_agent_prepareURL(t *testing.T) {
	type fields struct {
		config         *Config
		gaugeMetrics   map[string]interface{}
		counterMetrics map[string]interface{}
	}
	type args struct {
		name       string
		value      interface{}
		metricType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "should prepare gauge URL",
			fields: fields{
				config: &Config{
					CounterURL: url.URL{
						Scheme: "http",
						Host:   "localhost:8080",
						Path:   "/update/counter",
					},
					GaugeURL: url.URL{
						Scheme: "http",
						Host:   "localhost:8080",
						Path:   "/update/gauge",
					},
				},
			},
			args: args{
				name:       "test_metric_name",
				value:      42,
				metricType: gauge,
			},
			want: "http://localhost:8080/update/gauge/test_metric_name/42",
		},
		{
			name: "should prepare counter URL",
			fields: fields{
				config: &Config{
					CounterURL: url.URL{
						Scheme: "http",
						Host:   "localhost:8080",
						Path:   "/update/counter",
					},
					GaugeURL: url.URL{
						Scheme: "http",
						Host:   "localhost:8080",
						Path:   "/update/gauge",
					},
				},
			},
			args: args{
				name:       "test_counter",
				value:      100,
				metricType: counter,
			},
			want: "http://localhost:8080/update/counter/test_counter/100",
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			m := &agent{
				config:         tt.fields.config,
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
			}
			require.Equal(t, tt.want, m.prepareURL(tt.args.name, tt.args.value, tt.args.metricType))
		})
	}
}

func Test_agent_collectMetrics(t *testing.T) {
	type fields struct {
		URL            string
		config         *Config
		gaugeMetrics   map[string]interface{}
		counterMetrics map[string]interface{}
		mu             sync.Mutex
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
				gaugeMetrics:   make(map[string]interface{}),
				counterMetrics: make(map[string]interface{}),
				config: &Config{
					Client:         &http.Client{},
					PollInterval:   50 * time.Millisecond,
					ReportInterval: 100 * time.Second,
				},
				mu: sync.Mutex{},
			},
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			m := &agent{
				config:         tt.fields.config,
				gaugeMetrics:   tt.fields.gaugeMetrics,
				counterMetrics: tt.fields.counterMetrics,
				mu:             sync.Mutex{},
			}
			go m.collectMetrics(tt.args.stop)
			time.Sleep(100 * time.Millisecond)
			result := len(m.gaugeMetrics) + len(m.counterMetrics)

			require.NotEqual(t, result, 0, "Expected metrics to be collected")
		})
	}
}
