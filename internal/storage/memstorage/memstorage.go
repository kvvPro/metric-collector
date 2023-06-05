package memstorage

import (
	"errors"
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
		if fval, err := strconv.ParseFloat(v, 32); err == nil {
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
