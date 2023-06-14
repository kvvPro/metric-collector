package memstorage

import (
	"errors"
	"metric-collector/internal/metrics"
	"metric-collector/internal/storage"
	"strconv"
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

func (s *MemStorage) Update(t string, n string, v string) error {
	if t == storage.MetricTypeGauge {
		if fval, err := strconv.ParseFloat(v, 64); err == nil {
			s.Gauges[n] = fval
		}
	} else if t == storage.MetricTypeCounter {
		if ival, err := strconv.ParseInt(v, 10, 64); err == nil {
			s.Counters[n] += ival
		}
	} else {
		return errors.New("uknown metric type")
	}
	return nil
}

func (s *MemStorage) GetValue(t string, n string) (any, error) {
	var val any
	var exists bool

	if t == storage.MetricTypeGauge {
		val, exists = s.Gauges[n]
	} else if t == storage.MetricTypeCounter {
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
