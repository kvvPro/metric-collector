package client

import (
	"context"
	"testing"

	"github.com/kvvPro/metric-collector/cmd/agent/config"
	"github.com/kvvPro/metric-collector/internal/metrics"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		args    config.ClientFlags
		want    *Client
		wantErr bool
	}{
		{
			name: "1",
			args: config.ClientFlags{
				PollInterval:   2,
				ReportInterval: 10,
				Address:        "http://localhost:8080",
				HashKey:        "opa",
				RateLimit:      2,
			},
			want: &Client{
				pollInterval:   2,
				reportInterval: 10,
				Address:        "http://localhost:8080",
				needToHash:     true,
				hashKey:        "opa",
				// queue:          nil,
				maxWorkerCount: 2,
			},
			wantErr: false,
		},
		{
			name: "1",
			args: config.ClientFlags{
				PollInterval:   4,
				ReportInterval: 12,
				Address:        "http://localhost:8080",
				HashKey:        "param",
				RateLimit:      4,
			},
			want: &Client{
				pollInterval:   4,
				reportInterval: 12,
				Address:        "http://localhost:8080",
				needToHash:     true,
				hashKey:        "param",
				maxWorkerCount: 4,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(&tt.args)
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
				needToHash:     tt.fields.needToHash,
				hashKey:        tt.fields.hashKey,
				queue:          tt.fields.queue,
				maxWorkerCount: tt.fields.maxWorkerCount,
			}
			cli.ReadSpecificMetrics(tt.args.ctx)
		})
	}
}
