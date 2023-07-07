package client

import (
	"reflect"
	"testing"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

func TestDeepFields(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want []Metric
	}{
		{
			name: "1",
			arg: struct {
				Inner struct {
					A int64
					B int64
					C int64
				}
				PollCount   int64
				RandomValue float64
			}{
				Inner: struct {
					A int64
					B int64
					C int64
				}{
					1,
					2,
					3,
				},
				PollCount:   4,
				RandomValue: 445.1,
			},
			want: []Metric{
				&metrics.Counter{
					Name:  "A",
					Type:  "int64",
					Value: 1,
				},
				&metrics.Counter{
					Name:  "B",
					Type:  "int64",
					Value: 2,
				},
				&metrics.Counter{
					Name:  "C",
					Type:  "int64",
					Value: 3,
				},
				&metrics.Counter{
					Name:  "PollCount",
					Type:  "int64",
					Value: 4,
				},
				&metrics.Gauge{
					Name:  "RandomValue",
					Type:  "float64",
					Value: 445.1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeepFields(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeepFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMetric(t *testing.T) {
	type args struct {
		mname string
		mtype string
		ival  reflect.Value
	}
	tests := []struct {
		name string
		args args
		want Metric
	}{
		{
			name: "1",
			args: args{
				mname: "ghhgh",
				mtype: "int64",
				ival:  reflect.ValueOf(5),
			},
			want: &metrics.Counter{
				Name:  "ghhgh",
				Type:  "int64",
				Value: 5,
			},
		},
		{
			name: "2",
			args: args{
				mname: "ff",
				mtype: "float64",
				ival:  reflect.ValueOf(4123.09),
			},
			want: &metrics.Gauge{
				Name:  "ff",
				Type:  "float64",
				Value: 4123.09,
			},
		},
		{
			name: "3",
			args: args{
				mname: "adff",
				mtype: "ops",
				ival:  reflect.ValueOf(1113),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMetric(tt.args.mname, tt.args.mtype, tt.args.ival); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}
