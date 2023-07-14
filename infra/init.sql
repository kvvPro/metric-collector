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
    delta integer NOT NULL,
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