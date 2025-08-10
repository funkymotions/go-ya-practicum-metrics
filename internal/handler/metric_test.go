package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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

func Test_metricHandler_HandleMetric(t *testing.T) {
	type fields struct {
		service metricService
	}
	type args struct {
		w                  http.ResponseWriter
		entry              string
		metricVal          string
		metricType         string
		metricName         string
		spyMethodName      string
		serviceReturnValue error
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
				w:                  httptest.NewRecorder(),
				entry:              "/update",
				metricVal:          "42",
				metricType:         "gauge",
				metricName:         "test_metric",
				spyMethodName:      "SetGauge",
				serviceReturnValue: nil,
			},
			expected: expected{
				statusCode: http.StatusOK,
				body:       []byte(""),
			},
		},
		{
			name: "should handle counter metric",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				w:                  httptest.NewRecorder(),
				entry:              "/update",
				metricVal:          "42",
				metricType:         "counter",
				metricName:         "test_metric",
				spyMethodName:      "SetCounter",
				serviceReturnValue: nil,
			},
			expected: expected{
				statusCode: http.StatusOK,
				body:       []byte(""),
			},
		},
		{
			name: "should not handle invalid metric type",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				w:                  httptest.NewRecorder(),
				entry:              "/update",
				metricVal:          "42",
				metricType:         "counter-invalid",
				metricName:         "test_metric",
				spyMethodName:      "SetCounter",
				serviceReturnValue: nil,
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
				body:       []byte(""),
			},
		},
		{
			name: "should handle invalid URL with 400",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				w:                  httptest.NewRecorder(),
				entry:              "/update",
				metricVal:          "42/invalid",
				metricType:         "counter-invalid",
				metricName:         "test_metric",
				spyMethodName:      "SetCounter",
				serviceReturnValue: nil,
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
				body:       []byte(""),
			},
		},
		{
			name: "should handle metric service error",
			fields: fields{
				service: &metricServiceStub{},
			},
			args: args{
				w:                  httptest.NewRecorder(),
				entry:              "/update",
				metricVal:          "42",
				metricType:         "gauge",
				metricName:         "test_metric",
				spyMethodName:      "SetGauge",
				serviceReturnValue: fmt.Errorf("service error"),
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
			handler := http.HandlerFunc(h.HandleMetric)
			testServer := httptest.NewServer(handler)
			url := tt.args.entry + "/" + tt.args.metricType + "/" + tt.args.metricName + "/" + tt.args.metricVal
			h.service.(*metricServiceStub).
				On(tt.args.spyMethodName, tt.args.metricName, tt.args.metricVal).
				Return(tt.args.serviceReturnValue)
			resp, err := http.Post(testServer.URL+url, "text/plain", nil)
			if err != nil {
				t.Fatalf("Failed to make request: %v\n", err)
			}
			assert.Equal(t, tt.expected.statusCode, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			assert.Equal(t, body, tt.expected.body)
			testServer.Close()
		})
	}
}
