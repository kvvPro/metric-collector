package app

import (
	"context"
	"time"

	"github.com/kvvPro/metric-collector/internal/metrics"
	"github.com/kvvPro/metric-collector/internal/storage"
)

type Server struct {
	storage         storage.Storage
	Address         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DBConnection    string
}

func NewServer(store storage.Storage,
	address string,
	storeInterval int,
	filePath string,
	restore bool,
	dbconn string) *Server {
	return &Server{
		storage:         store,
		Address:         address,
		StoreInterval:   storeInterval,
		FileStoragePath: filePath,
		Restore:         restore,
		DBConnection:    dbconn,
	}
}

func (srv *Server) Ping(ctx context.Context) error {
	return srv.storage.Ping(ctx)
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
	if srv.StoreInterval == 0 {
		err = srv.SaveToFile()
		if err != nil {
			Sugar.Infoln("Save to file failed: ", err.Error())
		}
	}

	return nil
}

func (srv *Server) GetMetricValue(metricType string, metricName string) (any, error) {
	val, err := srv.storage.GetValue(metricType, metricName)
	return val, err
}

func (srv *Server) GetRequestedValues(m []metrics.Metric) []metrics.Metric {
	slice := srv.GetAllMetricsNew()
	hash := make(map[string]*metrics.Metric, 0)

	for _, el := range slice {
		hash[el.ID] = el
	}

	result := make([]metrics.Metric, 0)

	for _, el := range m {
		metricID := el.ID
		if _, isExist := hash[metricID]; isExist && el.MType == hash[metricID].MType {
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

func (srv *Server) GetAllMetrics() []storage.Metric {
	val := srv.storage.GetAllMetrics()
	return val
}
func (srv *Server) GetAllMetricsNew() []*metrics.Metric {
	val := srv.storage.GetAllMetricsNew()
	return val
}

func (srv *Server) AsyncSaving() {
	// run if only StoreInterval > 0, if StoreInterval = 0 => sync writing after each update
	// and FileStoragePath != ""
	if srv.StoreInterval > 0 && srv.FileStoragePath != "" {
		for {
			time.Sleep(time.Duration(srv.StoreInterval) * time.Second)

			err := srv.SaveToFile()
			if err != nil {
				Sugar.Infoln("Save to file failed: ", err.Error())
				panic(err)
			}
		}
	}
}

func (srv *Server) RestoreValues() {
	if srv.Restore {
		m, err := srv.ReadFromFile()
		if err != nil {
			Sugar.Infoln("Read values failed: ", err.Error())
			// panic(err)
		}

		for _, el := range m {
			srv.AddMetricNew(el)
		}
	}
}
