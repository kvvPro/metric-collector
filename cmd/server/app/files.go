package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

// SaveToFile saves metrics to file
func (srv *Server) SaveToFile(ctx context.Context) error {
	// сериализуем структуру в JSON формат
	m, err := srv.GetAllMetricsNew(ctx)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "   ")
	if err != nil {
		return err
	}

	// сохраняем данные в файл
	return os.WriteFile(srv.FileStoragePath, data, 0666)
}

// ReadFromFile reads all data from file
func (srv *Server) ReadFromFile() ([]metrics.Metric, error) {
	data, err := os.ReadFile(srv.FileStoragePath)
	if err != nil {
		return nil, err
	}
	m := make([]metrics.Metric, 0)
	reader := bytes.NewReader(data)
	if err := json.NewDecoder(reader).Decode(&m); err != nil {
		Sugar.Infoln("Read from file failed: ", err.Error())
		return nil, err
	}
	return m, nil
}
