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
	hash := make(map[string]*metrics.Metric, 0)

	for _, el := range slice {
		hash[el.GetName()] = metrics.NewCommonMetric(el.GetName(), el.GetTypeForQuery(), nil, nil)

		// try to init Value and Delta - to pass the tests
		if el.GetTypeForQuery() == metrics.MetricTypeGauge {
			val := el.GetValue().(float64)
			(hash[el.GetName()]).Value = &(val)
		}
		if el.GetTypeForQuery() == metrics.MetricTypeCounter {
			val := el.GetValue().(int64)
			(hash[el.GetName()]).Delta = &val
		}
	}

	result := make([]metrics.Metric, 0)

	for _, el := range m {
		metricID := el.ID
		if _, isExist := hash[metricID]; isExist {
			// add element with updated value
			result = append(result, *hash[metricID])
		} else {
			// keep requested value
			if el.MType == metrics.MetricTypeCounter {
				el.Delta = new(int64)
			} else {
				el.Value = new(float64)
			}
			result = append(result, el)
		}
	}

	return result
}

func (srv *Server) GetAllMetrics() []st.Metric {
	val := srv.storage.GetAllMetrics()
	return val
}
