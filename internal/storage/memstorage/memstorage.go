package memstorage

import (
	"context"
	"errors"
	_ "net/http/pprof"
	"strconv"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() MemStorage {
	return MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemStorage) Update(ctx context.Context, t string, n string, v string) error {
	if t == metrics.MetricTypeGauge {
		if fval, err := strconv.ParseFloat(v, 64); err == nil {
			s.Gauges[n] = fval
		}
	} else if t == metrics.MetricTypeCounter {
		if ival, err := strconv.ParseInt(v, 10, 64); err == nil {
			s.Counters[n] += ival
		}
	} else {
		return errors.New("uknown metric type")
	}
	return nil
}

func (s *MemStorage) UpdateNew(ctx context.Context, t string, n string, delta *int64, value *float64) error {
	if t == metrics.MetricTypeGauge {
		if value == nil {
			val := new(float64)
			s.Gauges[n] = *val
		} else {
			s.Gauges[n] = *value
		}
	} else if t == metrics.MetricTypeCounter {
		if delta == nil {
			val := new(int64)
			s.Counters[n] = *val
		} else {
			s.Counters[n] += *delta
		}
	} else {
		return errors.New("uknown metric type")
	}
	return nil
}

func (s *MemStorage) UpdateBatch(ctx context.Context, m []metrics.Metric) error {
	for _, el := range m {
		if el.MType == metrics.MetricTypeGauge {
			if el.Value == nil {
				val := new(float64)
				s.Gauges[el.ID] = *val
			} else {
				s.Gauges[el.ID] = *(el.Value)
			}
		} else if el.MType == metrics.MetricTypeCounter {
			if el.Delta == nil {
				val := new(int64)
				s.Counters[el.ID] = *val
			} else {
				s.Counters[el.ID] += *(el.Delta)
			}
		} else {
			return errors.New("uknown metric type")
		}
	}

	return nil
}

func (s *MemStorage) GetValue(ctx context.Context, t string, n string) (any, error) {
	var val any
	var exists bool

	if t == metrics.MetricTypeGauge {
		val, exists = s.Gauges[n]
	} else if t == metrics.MetricTypeCounter {
		val, exists = s.Counters[n]
	} else {
		return nil, errors.New("uknown metric type")
	}
	if !exists {
		return nil, errors.New("metric not found")
	}

	return val, nil
}

func (s *MemStorage) GetAllMetricsNew(ctx context.Context) ([]*metrics.Metric, error) {
	m := []*metrics.Metric{}

	for name, val := range s.Counters {
		newVal := val
		c := metrics.NewCommonMetric(name, metrics.MetricTypeCounter, &newVal, nil)
		m = append(m, c)
	}

	for name, val := range s.Gauges {
		newVal := val
		c := metrics.NewCommonMetric(name, metrics.MetricTypeGauge, nil, &newVal)
		m = append(m, c)
	}

	return m, nil
}
