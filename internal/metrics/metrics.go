package metrics

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
