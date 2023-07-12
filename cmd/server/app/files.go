package app

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/kvvPro/metric-collector/internal/metrics"
)

func (srv *Server) SaveToFile() error {
	// сериализуем структуру в JSON формат
	data, err := json.MarshalIndent(srv.storage.GetAllMetricsNew(), "", "   ")
	if err != nil {
		return err
	}
	// сохраняем данные в файл
	return os.WriteFile(srv.FileStoragePath, data, 0666)
}

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
