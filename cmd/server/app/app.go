package app

import (
	"github.com/kvvPro/metric-collector/internal/metrics"
	st "github.com/kvvPro/metric-collector/internal/storage"
)

type Server struct {
	storage st.Storage
	Host    string
	Port    string
}

func NewServer(store st.Storage, host string, port string) *Server {
	return &Server{
		storage: store,
		Host:    host,
		Port:    port,
	}
}

func (srv *Server) AddMetric(metricType string, metricName string, metricValue string) error {
	err := srv.storage.Update(metricType, metricName, metricValue)
	if err != nil {
		panic(err)
	}
	return nil
}

func (srv *Server) AddMetricNew(m metrics.Metric) error {
	err := srv.storage.UpdateNew(m.MType, m.ID, m.Delta, m.Value)
	if err != nil {
		panic(err)
	}
	return nil
}

func (srv *Server) GetMetricValue(metricType string, metricName string) (any, error) {
	val, err := srv.storage.GetValue(metricType, metricName)
	return val, err
}

func (srv *Server) GetRequestedValues(m []metrics.Metric) []metrics.Metric {
	slice := srv.GetAllMetrics()
	hash := make(map[string]metrics.Metric, 0)

	for _, el := range m {
		hash[el.ID] = el
	}

	result := make([]metrics.Metric, 0)

	for _, el := range slice {
		if val, isExist := hash[el.GetName()]; isExist {
			if el.GetTypeForQuery() == metrics.MetricTypeGauge {
				newValue := el.GetValue().(float64)
				val.Value = &newValue
			}
			if el.GetTypeForQuery() == metrics.MetricTypeCounter {
				newValue := el.GetValue().(int64)
				val.Delta = &newValue
			}
			result = append(result, val)
		}
	}
	return result
}

func (srv *Server) GetAllMetrics() []st.Metric {
	val := srv.storage.GetAllMetrics()
	return val
}
