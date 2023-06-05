package app

import (
	st "metric-collector/internal/storage"
)

type Server struct {
	storage st.Storage
	Port    string
}

func NewServer(store st.Storage, port string) (*Server, error) {
	return &Server{
		storage: store,
		Port:    port,
	}, nil
}

func (srv *Server) AddMetric(metricType string, metricName string, metricValue string) error {
	err := srv.storage.Update(metricType, metricName, metricValue)
	if err != nil {
		panic(err)
	}
	return nil
}
