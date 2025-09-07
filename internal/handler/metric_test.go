package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type metricServiceStub struct {
	mock.Mock
}

func (m *metricServiceStub) SetCounter(name string, value string) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *metricServiceStub) SetGauge(name string, value string) error {
	args := m.Called(name, value)
	return args.Error(0)
}

func (m *metricServiceStub) GetMetric(name string, metricType string) (*models.Metrics, bool) {
	args := m.Called(name, metricType)
	return args.Get(0).(*models.Metrics), args.Bool(1)
}

func (m *metricServiceStub) GetAllMetricsForHTML() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *metricServiceStub) SetMetricByModel(metric *models.Metrics) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *metricServiceStub) GetMetricByModel(metric *models.Metrics) (*models.Metrics, error) {
	args := m.Called(metric)
	return args.Get(0).(*models.Metrics), args.Error(1)
}

func TestNewMetricHandler(t *testing.T) {
	type args struct {
		s metricService
	}
	tests := []struct {
		name string
		args args
		want *metricHandler
	}{
		{
			name: "should create a new metric handler",
			args: args{
				s: &metricServiceStub{},
			},
			want: &metricHandler{
				service: &metricServiceStub{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewMetricHandler(tt.args.s)
			assert.True(t, reflect.DeepEqual(actual, tt.want))
		})
	}
}

func Test_metricHandler_SetMetric(t *testing.T) {
	type fields struct {
		service metricService
	}
	type args struct {
		entry              string
		metricType         string
		metricName         string
		metricVal          string
		serviceReturnValue error
		spyMethodName      string
	}
	type expected struct {
		statusCode int
		body       []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected expected
	}{
		{
			name: "should handle gauge metric",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				entry:              "/update",
				metricType:         "gauge",
				metricName:         "test_metric",
				metricVal:          "42",
				serviceReturnValue: nil,
				spyMethodName:      "SetGauge",
			},
			expected: expected{
				statusCode: http.StatusOK,
				body:       []byte(""),
			},
		},
		{
			name: "should handle metric service error",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				entry:              "/update",
				metricType:         "gauge",
				metricName:         "test_metric",
				metricVal:          "42",
				serviceReturnValue: fmt.Errorf("service error"),
				spyMethodName:      "SetGauge",
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
				body:       []byte(""),
			},
		},
		{
			name: "should handle counter metric",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				entry:              "/update",
				metricType:         "counter",
				metricName:         "test_metric",
				metricVal:          "42",
				serviceReturnValue: nil,
				spyMethodName:      "SetCounter",
			},
			expected: expected{
				statusCode: http.StatusOK,
				body:       []byte(""),
			},
		},
		{
			name: "should handle metric service error",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				entry:              "/update",
				metricType:         "counter",
				metricName:         "test_metric",
				metricVal:          "42",
				serviceReturnValue: fmt.Errorf("service error"),
				spyMethodName:      "SetCounter",
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
				body:       []byte(""),
			},
		},
		{
			name: "should return http 400 error for unknown metric type",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				entry:              "/update",
				metricType:         "unknown",
				metricName:         "test_metric",
				metricVal:          "42",
				serviceReturnValue: nil,
				spyMethodName:      "SetCounter",
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
				body:       []byte(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &metricHandler{
				service: tt.fields.service,
			}
			chi := chi.NewRouter()
			chi.Post("/update/{type}/{name}/{value}", h.SetMetric)
			ts := httptest.NewServer(chi)
			url := fmt.Sprintf(
				"%s/%s/%s/%s",
				tt.args.entry,
				tt.args.metricType,
				tt.args.metricName,
				tt.args.metricVal,
			)
			h.service.(*metricServiceStub).
				On(tt.args.spyMethodName, tt.args.metricName, tt.args.metricVal).
				Return(tt.args.serviceReturnValue)
			resp, err := http.Post(ts.URL+url, "text/plain", nil)
			if err != nil {
				t.Fatalf("failed to make request: %v\n", err)
			}
			assert.Equal(t, tt.expected.statusCode, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			assert.Equal(t, tt.expected.body, body)
			ts.Close()
		})
	}
}

func Test_metricHandler_GetAllMetrics(t *testing.T) {
	type fields struct {
		service metricService
	}
	type args struct {
	}
	type expected struct {
		statusCode int
		body       []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected expected
	}{
		{
			name: "should return all metrics",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{},
			expected: expected{
				statusCode: http.StatusOK,
				body:       []byte("test:1\ntest:1.1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &metricHandler{
				service: tt.fields.service,
			}
			r := chi.NewRouter()
			r.Get("/metrics", h.GetAllMetrics)
			ts := httptest.NewServer(r)
			h.service.(*metricServiceStub).
				On("GetAllMetricsForHTML").
				Return(string(tt.expected.body))
			resp, err := http.Get(ts.URL + "/metrics")
			if err != nil {
				t.Fatalf("failed to make request: %v\n", err)
			}
			assert.Equal(t, tt.expected.statusCode, resp.StatusCode)
			assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			assert.Equal(t, tt.expected.body, body)
			ts.Close()
		})
	}
}

func Test_metricHandler_GetMetric(t *testing.T) {
	var metricCounterValue = 1.1
	type fields struct {
		service metricService
	}
	type args struct {
		metricType string
		metricName string
	}
	type expected struct {
		statusCode               int
		body                     []byte
		serviceMetricReturnValue *models.Metrics
		serviceBoolReturnValue   bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected expected
	}{
		{
			name: "should return metric value",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				metricType: "gauge",
				metricName: "test_metric",
			},
			expected: expected{
				statusCode: http.StatusOK,
				body:       []byte("1.1"),
				serviceMetricReturnValue: &models.Metrics{
					ID:    "test_metric",
					Value: &metricCounterValue,
					MType: models.Gauge,
				},
				serviceBoolReturnValue: true,
			},
		},
		{
			name: "should return 404 http code when no metric found",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				metricType: "gauge",
				metricName: "test_metric",
			},
			expected: expected{
				statusCode:               http.StatusNotFound,
				body:                     []byte(""),
				serviceMetricReturnValue: nil,
				serviceBoolReturnValue:   false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &metricHandler{
				service: tt.fields.service,
			}
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", h.GetMetric)
			ts := httptest.NewServer(r)
			path := fmt.Sprintf("/value/%s/%s", tt.args.metricType, tt.args.metricName)
			h.service.(*metricServiceStub).
				On("GetMetric", tt.args.metricName, tt.args.metricType).
				Return(tt.expected.serviceMetricReturnValue, tt.expected.serviceBoolReturnValue)
			resp, err := http.Get(ts.URL + path)
			if err != nil {
				t.Fatalf("failed to make request: %v\n", err)
			}
			assert.Equal(t, tt.expected.statusCode, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			assert.Equal(t, tt.expected.body, body)
			ts.Close()
		})
	}
}
