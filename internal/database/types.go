package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
)

type DataBase interface {
	driver.Pinger
	io.Closer

	UpdateItems(ctx context.Context, records []*DBItem) error
	ReadItem(ctx context.Context, metricType string, metricName string) (*DBItem, error)
	ReadAllItems(ctx context.Context) ([]*DBItem, error)
}

type DBItem struct {
	MetricType sql.NullString
	Name       sql.NullString
	Value      sql.NullFloat64
}
