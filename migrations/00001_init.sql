-- +goose Up
CREATE TABLE metric_gauges
(
    id    bigint GENERATED ALWAYS AS IDENTITY,
    name  VARCHAR(256) PRIMARY KEY,
    value DOUBLE PRECISION NOT NULL
);

CREATE TABLE metric_counters
(
    id    bigint GENERATED ALWAYS AS IDENTITY,
    name  VARCHAR(256) PRIMARY KEY,
    value BIGINT NOT NULL
);

-- +goose Down
DROP TABLE metric_counters;
DROP TABLE metric_gauges;