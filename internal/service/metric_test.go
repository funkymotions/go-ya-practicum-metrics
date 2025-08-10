package service

import (
	"reflect"
	"testing"

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
			}
			s.repo.(*metricRepoStub).On("SetCounter", tt.args.name, int64(1)).Return(nil)
			err := s.SetCounter(tt.args.name, tt.args.rawValue)
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}
