package storage

import (
	"context"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

type Metric interface {
	GetName() string
	GetType() string
	GetValue() any
	GetTypeForQuery() string
}

type Storage interface {
	Ping(ctx context.Context) error
	Update(t string, n string, v string) error
	UpdateNew(ctx context.Context, t string, n string, delta *int64, value *float64) error
	GetValue(t string, n string) (any, error)
	GetAllMetrics() []Metric
	GetAllMetricsNew(ctx context.Context) ([]*metrics.Metric, error)
}
