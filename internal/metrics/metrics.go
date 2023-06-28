package metrics

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Counter struct {
	Name  string `json:"id"`
	Type  string `json:"type"`
	Value int64  `json:"delta,omitempty"`
}

type Gauge struct {
	Name  string  `json:"id"`
	Type  string  `json:"type"`
	Value float64 `json:"value,omitempty"`
}

func NewCommonMetric(mname string, mtype string, delta *int64, value *float64) *Metric {
	return &Metric{
		ID:    mname,
		MType: mtype,
		Delta: delta,
		Value: value,
	}
}

func NewCounter(mname string, mtype string, mint int64) *Counter {
	return &Counter{
		Name:  mname,
		Type:  mtype,
		Value: mint,
	}
}

func NewGauge(mname string, mtype string, mfloat float64) *Gauge {
	return &Gauge{
		Name:  mname,
		Type:  mtype,
		Value: mfloat,
	}
}

func (metric *Gauge) GetName() string {
	return metric.Name
}

func (metric *Counter) GetName() string {
	return metric.Name
}

func (metric *Gauge) GetType() string {
	return metric.Type
}

func (metric *Counter) GetType() string {
	return metric.Type
}

func (metric *Gauge) GetValue() any {
	return metric.Value
}

func (metric *Counter) GetValue() any {
	return metric.Value
}

func (metric *Gauge) GetTypeForQuery() string {
	return MetricTypeGauge
}

func (metric *Counter) GetTypeForQuery() string {
	return MetricTypeCounter
}
