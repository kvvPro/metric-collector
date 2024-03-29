package postgres

import (
	"context"
	"errors"
	_ "net/http/pprof"

	"github.com/kvvPro/metric-collector/internal/metrics"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	ConnStr string
}

func NewPSQLStr(ctx context.Context, connection string) (*PostgresStorage, error) {
	// init
	init := getInitQuery()
	dbpool, err := pgxpool.New(ctx, connection)
	if err != nil {
		return nil, err
	}

	defer dbpool.Close()

	_, err = dbpool.Exec(ctx, init)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		ConnStr: connection,
	}, nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	dbpool, err := pgxpool.New(ctx, s.ConnStr)
	if err != nil {
		return err
	}

	defer dbpool.Close()

	err = dbpool.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Deprecated: use UpdateNew
func (s *PostgresStorage) Update(ctx context.Context, t string, n string, v string) error {
	return errors.New("func is deprecated")
}

func (s *PostgresStorage) UpdateNew(ctx context.Context, mtype string, mname string, delta *int64, value *float64) error {

	dbpool, err := pgxpool.New(ctx, s.ConnStr)
	if err != nil {
		return err
	}

	defer dbpool.Close()

	transaction, err := dbpool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	err = updateMetric(ctx, dbpool, mtype, mname, delta, value)
	if err != nil {
		return err
	}

	transaction.Commit(ctx)

	return nil
}

func (s *PostgresStorage) UpdateBatch(ctx context.Context, m []metrics.Metric) error {
	dbpool, err := pgxpool.New(ctx, s.ConnStr)
	if err != nil {
		return err
	}

	defer dbpool.Close()

	transaction, err := dbpool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	for _, el := range m {
		err = updateMetric(ctx, dbpool, el.MType, el.ID, el.Delta, el.Value)
		if err != nil {
			return err
		}
	}

	transaction.Commit(ctx)

	return nil
}

func updateMetric(ctx context.Context,
	dbpool *pgxpool.Pool,
	mtype string,
	mname string,
	delta *int64,
	value *float64) error {
	// определим, есть ли запись с такой метрикой
	query := getSearchMetricByNameQuery()

	var metric metrics.Metric
	var metricID int64
	result := dbpool.QueryRow(ctx, query, mname)
	switch err := result.Scan(&metricID, &metric.MType, &metric.ID); err {
	case pgx.ErrNoRows:
		// сначала добавим саму метрику
		insertMetric := getInsertMetricQuery()
		insertRes, err := dbpool.Exec(ctx, insertMetric, mtype, mname)
		if err != nil {
			return err
		}
		if insertRes.RowsAffected() == 0 {
			return errors.New("metric not added")
		}

		// перечитаем id добавленной метрики
		result := dbpool.QueryRow(ctx, query, mname)
		if err := result.Scan(&metricID, &metric.MType, &metric.ID); err != nil {
			return err
		}

		// метрики нет, надо добавлять
		if mtype == metrics.MetricTypeCounter {
			insert := getInsertCounterQuery()
			insertRes, err := dbpool.Exec(ctx, insert, metricID, *delta)
			if err != nil {
				return err
			}
			if insertRes.RowsAffected() == 0 {
				return errors.New("metric not added")
			}
		} else if mtype == metrics.MetricTypeGauge {
			insert := getInsertGaugeQuery()
			insertRes, err := dbpool.Exec(ctx, insert, metricID, *value)
			if err != nil {
				return err
			}
			if insertRes.RowsAffected() == 0 {
				return errors.New("metric not added")
			}
		}

	case nil:
		// метрика есть - апдейтим
		if mtype == metrics.MetricTypeCounter {
			update := getUpdateCounterQuery()
			updateRes, err := dbpool.Exec(ctx, update, *delta, metricID)
			if err != nil {
				return err
			}
			if updateRes.RowsAffected() == 0 {
				return errors.New("metric not added")
			}
		} else if mtype == metrics.MetricTypeGauge {
			update := getUpdateGaugeQuery()
			updateRes, err := dbpool.Exec(ctx, update, *value, metricID)
			if err != nil {
				return err
			}
			if updateRes.RowsAffected() == 0 {
				return errors.New("metric not added")
			}
		}
	default:
		return err
	}

	return nil
}

func getSearchMetricByNameQuery() string {
	return `
	SELECT metrics.id as MetricID,
			metrics.mtype as MetricType,
			metrics.metric_name as MetricName
	FROM
		public.metrics
	WHERE
		metrics.metric_name = $1
	`
}

func getUpdateCounterQuery() string {
	return `
	UPDATE public.counters
	SET delta=delta + $1
	WHERE metric_id=$2;
	`
}

func getUpdateGaugeQuery() string {
	return `
	UPDATE public.gauges
	SET value=$1
	WHERE metric_id=$2;
	`
}

func getInsertMetricQuery() string {
	return `
	INSERT INTO public.metrics(mtype, metric_name)
		VALUES ($1, $2);
	`
}

func getInsertCounterQuery() string {
	return `
	INSERT INTO public.counters(
		metric_id, delta)
		VALUES ($1, $2);
	`
}

func getInsertGaugeQuery() string {
	return `
	INSERT INTO public.gauges(
		metric_id, value)
		VALUES ($1, $2);
	`
}

func getInitQuery() string {
	return `
	-- Table: public.metrics

	-- DROP TABLE IF EXISTS public.metrics;

	CREATE TABLE IF NOT EXISTS public.metrics
	(
		id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 0 MINVALUE 0 MAXVALUE 2147483647 CACHE 1 ),
		mtype character varying NOT NULL,
		metric_name character varying NOT NULL,
		CONSTRAINT metrics_pk PRIMARY KEY (id)
	)

	TABLESPACE pg_default;

	ALTER TABLE IF EXISTS public.metrics
		OWNER to postgres;
	-- Index: metrics_clustered

	-- DROP INDEX IF EXISTS public.metrics_clustered;

	CREATE UNIQUE INDEX IF NOT EXISTS metrics_clustered
		ON public.metrics USING btree
		(id ASC NULLS LAST)
		INCLUDE(id, mtype, metric_name)
		TABLESPACE pg_default;

	ALTER TABLE IF EXISTS public.metrics
		CLUSTER ON metrics_clustered;

	-- Index: metric_name_ind

	-- DROP INDEX IF EXISTS public.metric_name_ind;

	CREATE UNIQUE INDEX IF NOT EXISTS metric_name_ind
		ON public.metrics USING btree
		(metric_name ASC NULLS LAST)
		INCLUDE(id, mtype, metric_name)
		TABLESPACE pg_default;

	-- Table: public.counters

	-- DROP TABLE IF EXISTS public.counters;

	CREATE TABLE IF NOT EXISTS public.counters
	(
		metric_id integer NOT NULL,
		delta bigint NOT NULL,
		CONSTRAINT counters_metrics_id_fk FOREIGN KEY (metric_id)
			REFERENCES public.metrics (id) MATCH SIMPLE
			ON UPDATE NO ACTION
			ON DELETE NO ACTION
	)

	TABLESPACE pg_default;

	ALTER TABLE IF EXISTS public.counters
		OWNER to postgres;

	-- Table: public.gauges

	-- DROP TABLE IF EXISTS public.gauges;

	CREATE TABLE IF NOT EXISTS public.gauges
	(
		metric_id integer NOT NULL,
		value double precision NOT NULL,
		CONSTRAINT gauges_metrics_id_fk FOREIGN KEY (metric_id)
			REFERENCES public.metrics (id) MATCH SIMPLE
			ON UPDATE NO ACTION
			ON DELETE NO ACTION
	)

	TABLESPACE pg_default;

	ALTER TABLE IF EXISTS public.gauges
		OWNER to postgres;
	`
}

// Deprecated: use GetAllMetricsNew
func (s *PostgresStorage) GetValue(ctx context.Context, t string, n string) (any, error) {
	return nil, errors.New("func is deprecated")
}

func (s *PostgresStorage) GetAllMetricsNew(ctx context.Context) ([]*metrics.Metric, error) {
	m := []*metrics.Metric{}

	dbpool, err := pgxpool.New(ctx, s.ConnStr)
	if err != nil {
		return nil, err
	}

	defer dbpool.Close()

	query := `
	SELECT metrics.metric_name as MetricName,
			'counter' as MetricType,
			counters.delta as Delta,
			NULL as Value
	FROM
		public.counters INNER JOIN public.metrics
		ON counters.metric_id = metrics.id
		
	UNION ALL

	SELECT metrics.metric_name as MetricName,
			'gauge' as MetricType,
			NULL as Delta,
			gauges.value as Value
	FROM
		public.gauges INNER JOIN public.metrics
		ON gauges.metric_id = metrics.id
	`
	result, err := dbpool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer result.Close()

	for result.Next() {
		var metric metrics.Metric
		err = result.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			panic(err)
		}
		m = append(m, &metric)
	}

	err = result.Err()
	if err != nil {
		return nil, err
	}

	return m, nil
}
