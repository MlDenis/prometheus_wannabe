package db

import (
	"database/sql"
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/database"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/types"
)

func toDBRecord(metric metrics.Metric) *database.DBItem {
	return &database.DBItem{
		MetricType: sql.NullString{String: metric.GetType(), Valid: true},
		Name:       sql.NullString{String: metric.GetName(), Valid: true},
		Value:      sql.NullFloat64{Float64: metric.GetValue(), Valid: true},
	}
}

func fromDBRecord(record *database.DBItem) (metrics.Metric, error) {
	if !record.MetricType.Valid {
		return nil, logger.WrapError("read record", metrics.ErrInvalidRecordMetricType)
	}
	metricType := record.MetricType.String

	if !record.Name.Valid {
		return nil, logger.WrapError("read record", metrics.ErrInvalidRecordMetricName)
	}
	metricName := record.Name.String

	if !record.Value.Valid {
		return nil, logger.WrapError("read record", metrics.ErrInvalidRecordMetricValue)
	}
	value := record.Value.Float64

	var metric metrics.Metric
	switch metricType {
	case "gauge":
		metric = types.NewGaugeMetric(metricName)
	case "counter":
		metric = types.NewCounterMetric(metricName)
	default:
		return nil, logger.WrapError(fmt.Sprintf("read record with type '%s'", metricType), metrics.ErrUnknownMetricType)
	}

	metric.SetValue(value)
	return metric, nil
}
