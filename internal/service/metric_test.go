package service

import (
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	models "github.com/funkymotions/go-ya-practicum-metrics/internal/model"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type metricRepoStub struct {
	mock.Mock
}

func (m *metricRepoStub) SetGauge(name string, value float64) {
	m.Called(name, value)
}
func (m *metricRepoStub) SetCounter(name string, value int64) {
	m.Called(name, value)
}
func (m *metricRepoStub) GetMetric(name string, metricType string) (*models.Metrics, bool) {
	args := m.Called(name, metricType)
	return args.Get(0).(*models.Metrics), args.Bool(1)
}

func (m *metricRepoStub) GetAllMetrics() map[string]models.Metrics {
	args := m.Called()
	return args.Get(0).(map[string]models.Metrics)
}

func (m *metricRepoStub) SetGaugeIntrospect(name string, value float64) {
	m.Called(name, value)
}

func (m *metricRepoStub) SetCounterIntrospect(name string, value int64) {
	m.Called(name, value)
}

func (m *metricRepoStub) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *metricRepoStub) SetMetricBulk(metrics *[]models.Metrics) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func TestNewMetricService(t *testing.T) {
	type args struct {
		repo metricRepoInterface
	}
	tests := []struct {
		name string
		args args
		want *metricService
	}{
		{
			name: "should create a new metric service",
			args: args{
				repo: &metricRepoStub{},
			},
			want: &metricService{
				re:   regexp.MustCompile(`^\w+$`),
				repo: &metricRepoStub{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewMetricService(tt.args.repo)
			require.True(t, reflect.DeepEqual(actual, tt.want))
		})
	}
}

func Test_metricService_SetCounter(t *testing.T) {
	re := regexp.MustCompile(`^\w+$`)
	type fields struct {
		repo metricRepoInterface
	}
	type args struct {
		name     string
		rawValue string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantError bool
	}{
		{
			name: "should invoke repo SetCounter with no error",
			fields: fields{
				repo: &metricRepoStub{},
			},
			args: args{
				name:     "test_counter",
				rawValue: "1",
			},
			wantError: false,
		},
		{
			name: "should return error calling SetCounter",
			fields: fields{
				repo: &metricRepoStub{},
			},
			args: args{
				name:     "test_counter",
				rawValue: "invalid value",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				repo: tt.fields.repo,
				re:   re,
			}
			s.repo.(*metricRepoStub).On("SetCounterIntrospect", tt.args.name, int64(1)).Return(nil)
			err := s.SetCounter(tt.args.name, tt.args.rawValue)
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func Test_metricService_SetGauge(t *testing.T) {
	re := regexp.MustCompile(`^\w+$`)
	type fields struct {
		repo metricRepoInterface
	}
	type args struct {
		name     string
		rawValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should invoke repo SetGauge with no error",
			fields: fields{
				repo: &metricRepoStub{},
			},
			args: args{
				name:     "test_gauge",
				rawValue: "1.1",
			},
			wantErr: false,
		},
		{
			name: "should return error calling SetGauge",
			fields: fields{
				repo: &metricRepoStub{},
			},
			args: args{
				name:     "test_gauge",
				rawValue: "invalid value",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				repo: tt.fields.repo,
				re:   re,
			}

			s.repo.(*metricRepoStub).On("SetGaugeIntrospect", tt.args.name, float64(1.1)).Return(nil)
			err := s.SetGauge(tt.args.name, tt.args.rawValue)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func Test_metricService_GetAllMetricsForHTML(t *testing.T) {
	var val = 10.5
	var delta int64 = 20
	type fields struct {
		repo metricRepoInterface
	}
	tests := []struct {
		name          string
		fields        fields
		want          string
		repoReturnVal map[string]models.Metrics
	}{
		{
			name: "should return all metrics in HTML format",
			fields: fields{
				repo: &metricRepoStub{},
			},
			want: "20\n10.5\n",
			repoReturnVal: map[string]models.Metrics{
				"metric1": {
					ID:    "metric1",
					Delta: &delta,
					MType: models.Counter,
				},
				"metric2": {
					ID:    "metric2",
					Value: &val,
					MType: models.Gauge,
				},
			},
		},
		{
			name: "should return empty string",
			fields: fields{
				repo: &metricRepoStub{},
			},
			want:          "",
			repoReturnVal: map[string]models.Metrics{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &metricService{
				repo: tt.fields.repo,
			}
			s.repo.(*metricRepoStub).On("GetAllMetrics").Return(tt.repoReturnVal)
			actual := s.GetAllMetricsForHTML()
			actualLines := strings.Split(actual, "\n")
			expectedLines := strings.Split(tt.want, "\n")
			sort.Strings(actualLines)
			sort.Strings(expectedLines)
			assert.Equal(t, expectedLines, actualLines)
		})
	}
}
