package app

import (
	"reflect"
	"testing"

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
			if err := srv.AddMetric(tt.args.metricType, tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("Server.AddMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		store         st.Storage
		address       string
		storeInterval int
		filePath      string
		restore       bool
		dbconn        string
		// storageType   string
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
				address:       "localhost:8080",
				storeInterval: 200,
				filePath:      "/tmp/val.txt",
				restore:       true,
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
			got, err := NewServer(tt.args.address, tt.args.storeInterval, tt.args.filePath, tt.args.restore, tt.args.dbconn)
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
			got, err := srv.GetMetricValue(tt.args.metricType, tt.args.metricName)
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
			if got := srv.GetAllMetrics(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.GetAllMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
