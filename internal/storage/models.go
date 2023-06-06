package storage

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

type Counter struct {
	Name  string
	Type  string
	Value int64
}

type Gauge struct {
	Name  string
	Type  string
	Value float64
}
