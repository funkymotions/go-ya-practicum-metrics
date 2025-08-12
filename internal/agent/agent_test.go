package agent

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/funkymotions/go-ya-practicum-metrics/internal/config/env"
	"github.com/stretchr/testify/require"
)

func TestNewAgent(t *testing.T) {
	env := &env.Endpoint{Hostname: "localhost", Port: 8080}
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
					Endpoint: env,
				},
			},
			want: &agent{
				config: &Config{
					Endpoint: env,
				},
				gaugeEndpoint:   "http://" + env.String() + "/gauge",
				counterEndpoint: "http://" + env.String() + "/counter",
				gaugeMetrics:    make(map[string]interface{}),
				counterMetrics:  make(map[string]interface{}),
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

	parsedTestServerURL, _ := url.Parse(ts.URL)
	hostname := parsedTestServerURL.Hostname()
	port, _ := strconv.ParseUint(parsedTestServerURL.Port(), 10, 32)
	defer ts.Close()
	type fields struct {
		config          *Config
		gaugeEndpoint   string
		counterEndpoint string
		gaugeMetrics    map[string]interface{}
		counterMetrics  map[string]interface{}
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
				gaugeMetrics:    map[string]interface{}{"test_metric_name": 42},
				counterMetrics:  map[string]interface{}{"test_counter": 100},
				gaugeEndpoint:   ts.URL + "/update",
				counterEndpoint: ts.URL + "/update",
				config: &Config{
					Client:         &http.Client{},
					Endpoint:       &env.Endpoint{Hostname: hostname, Port: uint(port)},
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
				config:          tt.fields.config,
				gaugeEndpoint:   tt.fields.gaugeEndpoint,
				counterEndpoint: tt.fields.counterEndpoint,
				gaugeMetrics:    tt.fields.gaugeMetrics,
				counterMetrics:  tt.fields.counterMetrics,
			}
			go m.sendMetrics(tt.args.stop)
			time.Sleep(150 * time.Millisecond)
			require.Equal(t, tt.want, isServerWasCalled)
		})
	}
}

func Test_agent_prepareURL(t *testing.T) {
	type fields struct {
		config          *Config
		gaugeEndpoint   string
		counterEndpoint string
		gaugeMetrics    map[string]interface{}
		counterMetrics  map[string]interface{}
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
				gaugeEndpoint: "http://localhost:8080/update/gauge",
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
				counterEndpoint: "http://localhost:8080/update/counter",
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
				config:          tt.fields.config,
				gaugeEndpoint:   tt.fields.gaugeEndpoint,
				counterEndpoint: tt.fields.counterEndpoint,
				gaugeMetrics:    tt.fields.gaugeMetrics,
				counterMetrics:  tt.fields.counterMetrics,
			}
			require.Equal(t, tt.want, m.prepareURL(tt.args.name, tt.args.value, tt.args.metricType))
		})
	}
}

func Test_agent_collectMetrics(t *testing.T) {
	type fields struct {
		URL             string
		config          *Config
		gaugeEndpoint   string
		counterEndpoint string
		gaugeMetrics    map[string]interface{}
		counterMetrics  map[string]interface{}
		mu              sync.Mutex
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
					Endpoint:       &env.Endpoint{Hostname: "localhost", Port: 8080},
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
				config:          tt.fields.config,
				gaugeEndpoint:   tt.fields.gaugeEndpoint,
				counterEndpoint: tt.fields.counterEndpoint,
				gaugeMetrics:    tt.fields.gaugeMetrics,
				counterMetrics:  tt.fields.counterMetrics,
				mu:              sync.Mutex{},
			}
			go m.collectMetrics(tt.args.stop)
			time.Sleep(100 * time.Millisecond)
			result := len(m.gaugeMetrics) + len(m.counterMetrics)

			require.NotEqual(t, result, 0, "Expected metrics to be collected")
		})
	}
}
