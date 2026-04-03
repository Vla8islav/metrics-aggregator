-- +goose Up
CREATE TABLE metric_gauges
(
    name  TEXT PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL
);

CREATE TABLE metric_counters
(
    name  TEXT PRIMARY KEY,
    value BIGINT NOT NULL
);

-- +goose Down
DROP TABLE metric_counters;
DROP TABLE metric_gauges;