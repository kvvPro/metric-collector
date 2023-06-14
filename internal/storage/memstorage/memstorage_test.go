package memstorage

import (
	"metric-collector/internal/storage"
	"reflect"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want MemStorage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_Update(t *testing.T) {
	type fields struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}
	type args struct {
		t string
		n string
		v string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				Gauges:   tt.fields.Gauges,
				Counters: tt.fields.Counters,
			}
			if err := s.Update(tt.args.t, tt.args.n, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	type fields struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}
	tests := []struct {
		name   string
		fields fields
		want   []storage.Metric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				Gauges:   tt.fields.Gauges,
				Counters: tt.fields.Counters,
			}
			if got := s.GetAllMetrics(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.GetAllMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_GetValue(t *testing.T) {
	type fields struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}
	type args struct {
		t string
		n string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MemStorage{
				Gauges:   tt.fields.Gauges,
				Counters: tt.fields.Counters,
			}
			got, err := s.GetValue(tt.args.t, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
