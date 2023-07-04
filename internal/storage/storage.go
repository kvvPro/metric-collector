package storage

import "github.com/kvvPro/metric-collector/internal/metrics"

type Metric interface {
	GetName() string
	GetType() string
	GetValue() any
	GetTypeForQuery() string
}

type Storage interface {
	Update(t string, n string, v string) error
	UpdateNew(t string, n string, delta *int64, value *float64) error
	GetValue(t string, n string) (any, error)
	GetAllMetrics() []Metric
	GetAllMetricsNew() []metrics.Metric
}
