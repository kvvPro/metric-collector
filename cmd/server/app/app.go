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
	HashKey         string
	CheckHash       bool
}

const (
	DatabaseStorageType = "db"
	MemStorageType      = "memory"
)

func NewServer(address string,
	storeInterval int,
	filePath string,
	restore bool,
	dbconn string,
	hashKey string) (*Server, error) {

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
		HashKey:         hashKey,
		CheckHash:       hashKey != "",
	}, nil
}

func (srv *Server) Ping(ctx context.Context) error {
	return srv.storage.Ping(ctx)
}

func (srv *Server) AddMetric(ctx context.Context, metricType string, metricName string, metricValue string) error {
	err := srv.storage.Update(ctx, metricType, metricName, metricValue)
	if err != nil {
		return err
	}
	return nil
}

func (srv *Server) AddMetricNew(ctx context.Context, m metrics.Metric) error {
	var err error

	err = retry.Do(func() error {
		return srv.storage.UpdateNew(context.Background(), m.MType, m.ID, m.Delta, m.Value)
	},
		retry.RetryIf(func(errAttempt error) bool {
			var pgErr *pgconn.PgError
			if errors.As(errAttempt, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
				return true
			}
			return false
		}),
		retry.Attempts(3),
		retry.InitDelay(1000*time.Millisecond),
		retry.Step(2000*time.Millisecond),
		retry.Context(ctx),
	)

	if err != nil {
		Sugar.Errorln(err)
		return err
	}

	if srv.StoreInterval == 0 {
		err = srv.SaveToFile(ctx)
		if err != nil {
			Sugar.Infoln("Save to file failed: ", err.Error())
		}
	}

	return nil
}

func (srv *Server) AddMetricsBatch(ctx context.Context, m []metrics.Metric) error {

	var err error
	err = retry.Do(func() error {
		return srv.storage.UpdateBatch(context.Background(), m)
	},
		retry.RetryIf(func(errAttempt error) bool {
			var pgErr *pgconn.PgError
			if errors.As(errAttempt, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
				return true
			}
			return false
		}),
		retry.Attempts(3),
		retry.InitDelay(1000*time.Millisecond),
		retry.Step(2000*time.Millisecond),
		retry.Context(ctx),
	)

	if err != nil {
		Sugar.Errorln(err)
		return err
	}

	if srv.StoreInterval == 0 {
		err = srv.SaveToFile(ctx)
		if err != nil {
			Sugar.Infoln("Save to file failed: ", err.Error())
		}
	}

	return nil
}

func (srv *Server) GetMetricValue(ctx context.Context, metricType string, metricName string) (any, error) {
	val, err := srv.storage.GetValue(ctx, metricType, metricName)
	return val, err
}

func (srv *Server) GetRequestedValues(ctx context.Context, m []metrics.Metric) ([]metrics.Metric, error) {
	slice, err := srv.GetAllMetricsNew(ctx)
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

func (srv *Server) GetAllMetricsNew(ctx context.Context) ([]*metrics.Metric, error) {
	var val []*metrics.Metric
	var err error
	err = retry.Do(func() error {
		val, err = srv.storage.GetAllMetricsNew(context.Background())
		return err
	},
		retry.RetryIf(func(errAttempt error) bool {
			var pgErr *pgconn.PgError
			if errors.As(errAttempt, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
				return true
			}
			return false
		}),
		retry.Attempts(3),
		retry.InitDelay(1000*time.Millisecond),
		retry.Step(2000*time.Millisecond),
		retry.Context(ctx),
	)

	if err != nil {
		Sugar.Errorln(err)
		return nil, err
	}

	return val, nil
}

func (srv *Server) AsyncSaving(ctx context.Context) {
	// run if only StoreInterval > 0, if StoreInterval = 0 => sync writing after each update
	// and FileStoragePath != ""
	if srv.StoreInterval > 0 && srv.FileStoragePath != "" {
		for {
			time.Sleep(time.Duration(srv.StoreInterval) * time.Second)

			err := srv.SaveToFile(ctx)
			if err != nil {
				Sugar.Infoln("Save to file failed: ", err.Error())
			}
		}
	}
}

func (srv *Server) RestoreValues(ctx context.Context) {
	if srv.Restore && srv.StorageType == "memory" {
		m, err := srv.ReadFromFile()
		if err != nil {
			Sugar.Infoln("Read values failed: ", err.Error())
		}

		for _, el := range m {
			srv.AddMetricNew(ctx, el)
		}
	}
}
