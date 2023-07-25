package client

import (
	"context"
	"testing"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

func TestNewClient(t *testing.T) {
	type args struct {
		pollInterval   int
		reportInterval int
		address        string
		contentType    string
		needToHash     bool
		hashKey        string
		//queue          chan []metrics.Metric
		maxWorkerCount int
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				pollInterval:   2,
				reportInterval: 10,
				address:        "http://localhost:8080",
				contentType:    "text/plain",
				needToHash:     true,
				hashKey:        "opa",
				//queue:          nil,
				maxWorkerCount: 2,
			},
			want: &Client{
				pollInterval:   2,
				reportInterval: 10,
				Address:        "http://localhost:8080",
				contentType:    "text/plain",
				needToHash:     true,
				hashKey:        "opa",
				//queue:          nil,
				maxWorkerCount: 2,
			},
			wantErr: false,
		},
		{
			name: "1",
			args: args{
				pollInterval:   4,
				reportInterval: 12,
				address:        "http://localhost:8080",
				contentType:    "opa",
				needToHash:     true,
				hashKey:        "param",
				//queue:          nil,
				maxWorkerCount: 4,
			},
			want: &Client{
				pollInterval:   4,
				reportInterval: 12,
				Address:        "http://localhost:8080",
				contentType:    "opa",
				needToHash:     true,
				hashKey:        "param",
				//queue:          nil,
				maxWorkerCount: 4,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.args.pollInterval, tt.args.reportInterval, tt.args.address,
				tt.args.contentType, tt.args.hashKey, tt.args.maxWorkerCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !assert.Equal(t, got, tt.want) {
			// 	t.Errorf("NewClient() = %v, want %v", got, tt.want)
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewClient() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestClient_ReadMetrics(t *testing.T) {
	type fields struct {
		Metrics        Metrics
		pollInterval   int
		reportInterval int
		address        string
		contentType    string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &Client{
				Metrics:        tt.fields.Metrics,
				pollInterval:   tt.fields.pollInterval,
				reportInterval: tt.fields.reportInterval,
				Address:        tt.fields.address,
				contentType:    tt.fields.contentType,
			}
			cli.ReadMetrics(context.Background())
		})
	}
}

func TestClient_PushMetrics(t *testing.T) {
	type fields struct {
		Metrics        Metrics
		pollInterval   int
		reportInterval int
		address        string
		contentType    string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &Client{
				Metrics:        tt.fields.Metrics,
				pollInterval:   tt.fields.pollInterval,
				reportInterval: tt.fields.reportInterval,
				Address:        tt.fields.address,
				contentType:    tt.fields.contentType,
			}
			cli.PushMetricsJSON(context.Background())
		})
	}
}

func TestClient_ReadSpecificMetrics(t *testing.T) {
	type fields struct {
		Metrics        Metrics
		pollInterval   int
		reportInterval int
		Address        string
		contentType    string
		needToHash     bool
		hashKey        string
		queue          chan []metrics.Metric
		maxWorkerCount int
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := &Client{
				Metrics:        tt.fields.Metrics,
				pollInterval:   tt.fields.pollInterval,
				reportInterval: tt.fields.reportInterval,
				Address:        tt.fields.Address,
				contentType:    tt.fields.contentType,
				needToHash:     tt.fields.needToHash,
				hashKey:        tt.fields.hashKey,
				queue:          tt.fields.queue,
				maxWorkerCount: tt.fields.maxWorkerCount,
			}
			cli.ReadSpecificMetrics(tt.args.ctx)
		})
	}
}
