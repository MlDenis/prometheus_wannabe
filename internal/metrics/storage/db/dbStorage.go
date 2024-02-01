package db

import (
	"context"
	"database/sql"

	"github.com/MlDenis/prometheus_wannabe/internal/converter"
	"github.com/MlDenis/prometheus_wannabe/internal/database"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/storage"
)

type dbStorage struct {
	dataBase database.DataBase
}

func NewDBStorage(dataBase database.DataBase) storage.MetricsStorage {
	return &dbStorage{dataBase: dataBase}
}

func (d *dbStorage) AddMetricValues(ctx context.Context, metricsList []metrics.Metric) ([]metrics.Metric, error) {
	dbRecords := make([]*database.DBItem, len(metricsList))
	for i, metric := range metricsList {
		dbRecords[i] = toDBRecord(metric)
	}

	err := d.dataBase.UpdateItems(ctx, dbRecords)
	if err != nil {
		return nil, logger.WrapError("update db record", err)
	}

	return metricsList, nil
}

func (d *dbStorage) GetMetricValues(ctx context.Context) (map[string]map[string]string, error) {
	records, err := d.dataBase.ReadAllItems(ctx)
	if err != nil {
		return nil, logger.WrapError("read all db records", err)
	}

	result := map[string]map[string]string{}
	for _, record := range records {
		if !record.MetricType.Valid {
			return nil, logger.WrapError("read record", metrics.ErrInvalidRecordMetricType)
		}

		metricType := record.MetricType.String
		metricsByType, ok := result[metricType]
		if !ok {
			metricsByType = map[string]string{}
			result[metricType] = metricsByType
		}

		if !record.Name.Valid {
			return nil, logger.WrapError("read record", metrics.ErrInvalidRecordMetricName)
		}
		metricName := record.Name.String

		if !record.Value.Valid {
			return nil, logger.WrapError("read record", metrics.ErrInvalidRecordMetricValue)
		}

		metricsByType[metricName] = converter.FloatToString(record.Value.Float64)
	}

	return result, nil
}

func (d *dbStorage) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	result, err := d.dataBase.ReadItem(ctx, metricType, metricName)
	if err != nil {
		return nil, logger.WrapError("read db record", err)
	}

	return fromDBRecord(result)
}

func (d *dbStorage) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
	records := []*database.DBItem{}
	for metricType, metricsByType := range metricValues {
		for metricName, metricValue := range metricsByType {
			value, err := converter.ToFloat64(metricValue)
			if err != nil {
				return logger.WrapError("parse metric value", err)
			}

			records = append(records, &database.DBItem{
				MetricType: sql.NullString{String: metricType, Valid: true},
				Name:       sql.NullString{String: metricName, Valid: true},
				Value:      sql.NullFloat64{Float64: value, Valid: true},
			})
		}
	}

	err := d.dataBase.UpdateItems(ctx, records)
	if err != nil {
		return logger.WrapError("update records", err)
	}

	return nil
}
