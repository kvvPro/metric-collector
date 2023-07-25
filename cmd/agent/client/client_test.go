package client

import (
	"context"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	type args struct {
		pollInterval   int
		reportInterval int
		address        string
		contentType    string
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
			},
			want: &Client{
				pollInterval:   2,
				reportInterval: 10,
				Address:        "http://localhost:8080",
				contentType:    "text/plain",
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
			},
			want: &Client{
				pollInterval:   4,
				reportInterval: 12,
				Address:        "http://localhost:8080",
				contentType:    "opa",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.pollInterval, tt.args.reportInterval, tt.args.address, tt.args.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
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
			cli.ReadMetrics()
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
