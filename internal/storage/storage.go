package storage

type Metric interface {
	GetName() string
	GetType() string
	GetValue() any
	GetTypeForQuery() string
}

type Storage interface {
	Update(t string, n string, v string) error
	GetValue(t string, n string) (any, error)
	GetAllMetrics() []Metric
}
