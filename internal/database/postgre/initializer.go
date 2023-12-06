package postgre

import (
	"context"
	"database/sql"
	"github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MlDenis/prometheus_wannabe/internal/logger"
)

func initDB(ctx context.Context, connectionString string) (*sql.DB, error) {
	logrus.Info("Initialize database schema")

	conn, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, logger.WrapError("open db connection", err)
	}

	err = conn.PingContext(ctx)
	if err != nil {
		return nil, logger.WrapError("ping db connection", err)
	}

	return conn, nil
}

func getOrCreateMetricTypeId(ctx context.Context, conn *sql.DB, typeName string) (int, error) {
	var typeId int
	err := conn.QueryRowContext(ctx, "SELECT id FROM metricType WHERE name = $1", typeName).Scan(&typeId)
	if err != nil {
		if err == sql.ErrNoRows {
			err = conn.QueryRowContext(ctx, "INSERT INTO metricType(name) VALUES ($1) RETURNING id", typeName).Scan(&typeId)
		}
		if err != nil {
			return 0, err
		}
	}
	return typeId, nil
}

func getOrCreateMetricId(ctx context.Context, conn *sql.DB, metricTypeName string, metricName string) (int, error) {
	var metricId int
	metricTypeId, err := getOrCreateMetricTypeId(ctx, conn, metricTypeName)
	if err != nil {
		return 0, err
	}
	err = conn.QueryRowContext(ctx, "SELECT id FROM metric WHERE name = $1 AND typeId = $2", metricName, metricTypeId).Scan(&metricId)
	if err != nil {
		if err == sql.ErrNoRows {
			err = conn.QueryRowContext(ctx, "INSERT INTO metric(name, typeId) VALUES ($1, $2) RETURNING id", metricName, metricTypeId).Scan(&metricId)
		}
		if err != nil {
			return 0, err
		}
	}
	return metricId, nil
}

func updateOrCreateMetric(ctx context.Context, conn *sql.DB, metricTypeName string, metricName string, metricValue float64) error {
	metricId, err := getOrCreateMetricId(ctx, conn, metricTypeName, metricName)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, "UPDATE metric SET value = $1 WHERE id = $2", metricValue, metricId)
	return err
}
