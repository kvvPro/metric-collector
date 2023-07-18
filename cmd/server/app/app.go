package app

import (
	"context"
	"errors"
	"time"

	"github.com/kvvPro/metric-collector/internal/metrics"
	"github.com/kvvPro/metric-collector/internal/retry"
	"github.com/kvvPro/metric-collector/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kvvPro/metric-collector/internal/storage/memstorage"
	"github.com/kvvPro/metric-collector/internal/storage/postgres"
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
		newdb, err := postgres.NewPSQLStr(context.Background(), dbconn)
		if err != nil {
			return nil, err
		}
		st = newdb
	} else {
		t = MemStorageType
		newmem := memstorage.NewMemStorage()
		st = &newmem
	}

	if st == nil {
		return nil, errors.New("cannot create storage for server")
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
		return err
	}
	return nil
}

func (srv *Server) AddMetricNew(m metrics.Metric) error {
	err := srv.storage.UpdateNew(context.Background(), m.MType, m.ID, m.Delta, m.Value)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry.Do(func() error {
				return srv.storage.UpdateNew(context.Background(), m.MType, m.ID, m.Delta, m.Value)
			},
				retry.Attempts(3),
				retry.PauseBeforeFirstAttempt(true),
			)
		}

		if err != nil {
			Sugar.Errorln(err)
			return err
		}
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry.Do(func() error {
				return srv.storage.UpdateBatch(context.Background(), m)
			},
				retry.Attempts(3),
				retry.PauseBeforeFirstAttempt(true),
			)
		}

		if err != nil {
			Sugar.Errorln(err)
			return err
		}
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

func (srv *Server) GetRequestedValues(m []metrics.Metric) ([]metrics.Metric, error) {
	slice, err := srv.GetAllMetricsNew()
	if err != nil {
		return nil, err
	}
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

	return result, nil
}

func (srv *Server) GetAllMetrics() []storage.Metric {
	val := srv.storage.GetAllMetrics()
	return val
}
func (srv *Server) GetAllMetricsNew() ([]*metrics.Metric, error) {
	var val []*metrics.Metric
	val, err := srv.storage.GetAllMetricsNew(context.Background())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			err = retry.Do(func() error {
				val, err = srv.storage.GetAllMetricsNew(context.Background())
				return err
			},
				retry.Attempts(3),
				retry.PauseBeforeFirstAttempt(true),
			)
		}

		if err != nil {
			Sugar.Errorln(err)
			return nil, err
		}
	}
	return val, nil
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
			}
		}
	}
}

func (srv *Server) RestoreValues() {
	if srv.Restore && srv.StorageType == "memory" {
		m, err := srv.ReadFromFile()
		if err != nil {
			Sugar.Infoln("Read values failed: ", err.Error())
		}

		for _, el := range m {
			srv.AddMetricNew(el)
		}
	}
}
