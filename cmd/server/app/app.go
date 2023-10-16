// Package app provides functions to work with main server - metric collector
package app

import (
	"context"
	"errors"
	"time"

	"github.com/kvvPro/metric-collector/cmd/server/config"
	"github.com/kvvPro/metric-collector/internal/metrics"
	"github.com/kvvPro/metric-collector/internal/retry"
	"github.com/kvvPro/metric-collector/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kvvPro/metric-collector/internal/storage/memstorage"
	"github.com/kvvPro/metric-collector/internal/storage/postgres"
)

type Server struct {
	// Main storage for metrics, can be memstorage or postgresql type
	storage storage.Storage
	// Address of web server, where app woul be deployed
	// Format - Host[:Port]
	Address string
	// Interval in seconds for backup storage on disk
	StoreInterval int
	// Path to save storage with interval StoreInterval
	FileStoragePath string
	// True if server would restore metrics from backup file
	Restore bool
	// String connection to postgres DB
	// Format - "user=<user> password=<pass> host=<host> port=<port> dbname=<db> sslmode=<true/false>"
	DBConnection string
	// "db" or "memory"
	StorageType string
	// Key for decrypt and encrypt body of requests
	HashKey string
	// True if server would validate hash of all incoming requests
	CheckHash bool
	// True if server accepts encrypted messages from agent
	UseEncryption bool
	// Path to private key RSA
	PrivateKeyPath string
}

const (
	DatabaseStorageType = "db"
	MemStorageType      = "memory"
)

// NewServer creates app instance
func NewServer(settings *config.ServerFlags) (*Server, error) {

	var t string
	var st storage.Storage

	if settings.DBConnection != "" {
		t = DatabaseStorageType
		newdb, err := postgres.NewPSQLStr(context.Background(), settings.DBConnection)
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
		Address:         settings.Address,
		StoreInterval:   settings.StoreInterval,
		FileStoragePath: settings.FileStoragePath,
		Restore:         settings.Restore,
		DBConnection:    settings.DBConnection,
		StorageType:     t,
		HashKey:         settings.HashKey,
		CheckHash:       settings.HashKey != "",
		PrivateKeyPath:  settings.CryptoKey,
		UseEncryption:   settings.CryptoKey != "",
	}, nil
}

// Ping tests connection to db
func (srv *Server) Ping(ctx context.Context) error {
	return srv.storage.Ping(ctx)
}

// AddMetric adds new metric if it doesn't exist, or update existing metric with name metricName
//
// Deprecated: use AddMetricNew
func (srv *Server) AddMetric(ctx context.Context, metricType string, metricName string, metricValue string) error {
	err := srv.storage.Update(ctx, metricType, metricName, metricValue)
	if err != nil {
		return err
	}
	return nil
}

// AddMetricNew adds new metric if it doesn't exist, or update existing metric with name metricName
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

// AddMetricsBatch adds array of new metrics.
// Behavior is the same as in AddMetricNew
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

// GetMetricValue returns current value of metric
func (srv *Server) GetMetricValue(ctx context.Context, metricType string, metricName string) (any, error) {
	val, err := srv.storage.GetValue(ctx, metricType, metricName)
	return val, err
}

// GetRequestedValues returns current values of requested metrics.
// return only already existed metrics
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

// GetAllMetricsNew returns all existed metrics with current values
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

// AsyncSaving backups database to file
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

// RestoreValues restore metrics from file
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
