package app

import (
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

func (srv *Server) GetMetricValue(metricType string, metricName string) (any, error) {
	val, err := srv.storage.GetValue(metricType, metricName)
	return val, err
}

func (srv *Server) GetAllMetrics() []st.Metric {
	val := srv.storage.GetAllMetrics()
	return val
}
