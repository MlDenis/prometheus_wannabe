-- +goose Up
CREATE TABLE IF NOT EXISTS metricType (
                                          id SMALLSERIAL PRIMARY KEY,
                                          name TEXT
);

CREATE TABLE IF NOT EXISTS metric (
                                      id SERIAL PRIMARY KEY,
                                      name TEXT,
                                      typeId SMALLSERIAL,
                                      value DOUBLE PRECISION
);

CREATE UNIQUE INDEX IF NOT EXISTS metric_name_type_idx ON metric (name, typeId);

-- +goose Down
DROP INDEX IF EXISTS metric_name_type_idx;
DROP TABLE IF EXISTS metric;
DROP TABLE IF EXISTS metricType;
