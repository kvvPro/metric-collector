package memstorage

import (
	"context"
	"errors"
	"strconv"

	"github.com/kvvPro/metric-collector/internal/metrics"
	"github.com/kvvPro/metric-collector/internal/storage"
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

func (s *MemStorage) Update(t string, n string, v string) error {
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

func (s *MemStorage) UpdateNew(t string, n string, delta *int64, value *float64) error {
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

func (s *MemStorage) GetValue(t string, n string) (any, error) {
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

func (s *MemStorage) GetAllMetrics() []storage.Metric {
	m := []storage.Metric{}

	for name, val := range s.Counters {
		c := metrics.NewCounter(name, "int64", val)
		m = append(m, c)
	}

	for name, val := range s.Gauges {
		c := metrics.NewGauge(name, "float64", val)
		m = append(m, c)
	}

	return m
}

func (s *MemStorage) GetAllMetricsNew() []*metrics.Metric {
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

	return m
}
