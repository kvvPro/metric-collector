package app

import (
	"context"
	"time"

	"github.com/kvvPro/metric-collector/internal/metrics"
	"github.com/kvvPro/metric-collector/internal/storage"

	mem "github.com/kvvPro/metric-collector/internal/storage/memstorage"
	db "github.com/kvvPro/metric-collector/internal/storage/postgres"
)

type Server struct {
	storage         storage.Storage
	Address         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DBConnection    string
	StorageType     string
}

const (
	DatabaseStorageType = "db"
	MemStorageType      = "memory"
)

func NewServer(address string,
	storeInterval int,
	filePath string,
	restore bool,
	dbconn string) (*Server, error) {

	var t string
	var st storage.Storage

	if dbconn != "" {
		t = DatabaseStorageType
		newdb, err := db.NewPSQLStr(context.Background(), dbconn)
		if err != nil {
			return nil, err
		}
		st = newdb
	} else {
		t = MemStorageType
		newmem := mem.NewMemStorage()
		st = &newmem
	}

	return &Server{
		storage:         st,
		Address:         address,
		StoreInterval:   storeInterval,
		FileStoragePath: filePath,
		Restore:         restore,
		DBConnection:    dbconn,
		StorageType:     t,
	}, nil
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
	err := srv.storage.UpdateNew(context.Background(), m.MType, m.ID, m.Delta, m.Value)
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

func (srv *Server) AddMetricsBatch(m []metrics.Metric) error {
	err := srv.storage.UpdateBatch(context.Background(), m)
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
	val, err := srv.storage.GetAllMetricsNew(context.Background())
	if err != nil {
		panic(err)
	}
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
	if srv.Restore && srv.StorageType == "memory" {
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
