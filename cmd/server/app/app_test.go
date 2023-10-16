package app

import (
	"context"
	"reflect"
	"testing"

	"github.com/kvvPro/metric-collector/cmd/server/config"
	"github.com/kvvPro/metric-collector/internal/metrics"
	st "github.com/kvvPro/metric-collector/internal/storage"
	"github.com/kvvPro/metric-collector/internal/storage/memstorage"
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
			}
			if err := srv.AddMetric(context.Background(), tt.args.metricType, tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("Server.AddMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		store    st.Storage
		settings config.ServerFlags
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
				settings: config.ServerFlags{
					Address:         "localhost:8080",
					StoreInterval:   200,
					FileStoragePath: "/tmp/val.txt",
					Restore:         true,
				},

				// dbconn:        "user=postgres password=postgres host=localhost port=5432 dbname=postgres sslmode=disable",
				// storageType:   "db",
			},
			want: &Server{
				storage: &memstorage.MemStorage{
					Gauges:   make(map[string]float64),
					Counters: make(map[string]int64),
				},
				Address:         "localhost:8080",
				StoreInterval:   200,
				FileStoragePath: "/tmp/val.txt",
				Restore:         true,
				// DBConnection:    "user=postgres password=postgres host=localhost port=5432 dbname=postgres sslmode=disable",
				StorageType: "memory",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServer(&tt.args.settings)
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_GetMetricValue(t *testing.T) {
	type fields struct {
		storage st.Storage
		Port    string
	}
	type args struct {
		metricType string
		metricName string
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
			srv := &Server{
				storage: tt.fields.storage,
			}
			got, err := srv.GetMetricValue(context.Background(), tt.args.metricType, tt.args.metricName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.GetMetricValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.GetMetricValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_GetAllMetrics(t *testing.T) {
	type fields struct {
		storage st.Storage
		Port    string
	}
	tests := []struct {
		name   string
		fields fields
		want   []st.Metric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Server{
				storage: tt.fields.storage,
			}
			if got, err := srv.GetAllMetricsNew(context.Background()); !reflect.DeepEqual(got, tt.want) || err != nil {
				t.Errorf("Server.GetAllMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkGetAllMetrics(b *testing.B) {
	memst := memstorage.NewMemStorage()
	srv := &Server{
		storage: &memst,
	}
	for i := 0; i < b.N; i++ {
		srv.GetAllMetricsNew(context.Background())
	}
}

func BenchmarkGetMetricValue(b *testing.B) {
	memst := memstorage.NewMemStorage()
	srv := &Server{
		storage: &memst,
	}
	for i := 0; i < b.N; i++ {
		srv.GetMetricValue(context.Background(), "gauge", "mem_usage")
	}
}

func BenchmarkUpdateMetric(b *testing.B) {
	memst := memstorage.NewMemStorage()
	srv := &Server{
		storage: &memst,
	}
	val := 10.7
	for i := 0; i < b.N; i++ {
		srv.AddMetricNew(context.Background(), metrics.Metric{ID: "cpu", MType: "gauge", Value: &val})
	}
}
