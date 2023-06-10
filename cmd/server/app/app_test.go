package app

import (
	st "metric-collector/internal/storage"
	"metric-collector/internal/storage/memstorage"
	"reflect"
	"testing"
)

func TestServer_AddMetric(t *testing.T) {
	type fields struct {
		storage st.Storage
		Port    string
	}
	type args struct {
		metricType  string
		metricName  string
		metricValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				storage: &memstorage.MemStorage{
					Gauges:   make(map[string]float64),
					Counters: make(map[string]int64),
				},
			},
			args: args{
				metricType:  "counter",
				metricName:  "test1",
				metricValue: "250",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
				Port:    tt.fields.Port,
			}
			if err := srv.AddMetric(tt.args.metricType, tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("Server.AddMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		store st.Storage
		port  string
	}
	tests := []struct {
		name    string
		args    args
		want    *Server
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				store: &memstorage.MemStorage{
					Gauges:   make(map[string]float64),
					Counters: make(map[string]int64),
				},
				port: "8080",
			},
			want: &Server{
				storage: &memstorage.MemStorage{
					Gauges:   make(map[string]float64),
					Counters: make(map[string]int64),
				},
				Port: "8080",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServer(tt.args.store, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
