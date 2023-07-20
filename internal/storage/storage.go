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
	Update(ctx context.Context, t string, n string, v string) error
	UpdateNew(ctx context.Context, t string, n string, delta *int64, value *float64) error
	UpdateBatch(ctx context.Context, m []metrics.Metric) error
	GetValue(ctx context.Context, t string, n string) (any, error)
	GetAllMetrics(ctx context.Context) ([]Metric, error)
	GetAllMetricsNew(ctx context.Context) ([]*metrics.Metric, error)
}
