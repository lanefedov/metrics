CREATE TABLE IF NOT EXISTS metrics (
    metric_type text NOT NULL,
    name text NOT NULL,
    counter_value bigint,
    gauge_value double precision,
    PRIMARY KEY (metric_type, name),
    CHECK (
        (metric_type = 'counter' AND counter_value IS NOT NULL AND gauge_value IS NULL)
        OR
        (metric_type = 'gauge' AND gauge_value IS NOT NULL AND counter_value IS NULL)
    )
);
