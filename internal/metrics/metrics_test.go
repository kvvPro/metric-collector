package metrics

import (
	"reflect"
	"testing"
)

func TestNewCounter(t *testing.T) {
	type args struct {
		mname string
		mtype string
		mint  int64
	}
	tests := []struct {
		name string
		args args
		want *Counter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCounter(tt.args.mname, tt.args.mtype, tt.args.mint); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGauge(t *testing.T) {
	type args struct {
		mname  string
		mtype  string
		mfloat float64
	}
	tests := []struct {
		name string
		args args
		want *Gauge
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGauge(tt.args.mname, tt.args.mtype, tt.args.mfloat); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGauge() = %v, want %v", got, tt.want)
			}
		})
	}
}
