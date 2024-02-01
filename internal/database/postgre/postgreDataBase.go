package postgre

import (
	"context"
	"database/sql"

	"github.com/MlDenis/prometheus_wannabe/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

type PostgresDataaBaseConfig interface {
	GetConnectionString() string
}

type postgresDataBase struct {
	conn *sql.DB
}

func NewPostgresDataBase(ctx context.Context, conf PostgresDataaBaseConfig) (database.DataBase, error) {
	conn, err := initDB(ctx, conf.GetConnectionString())
	if err != nil {
		return nil, err
	}

	return &postgresDataBase{conn: conn}, nil
}

func (p *postgresDataBase) UpdateItems(ctx context.Context, records []*database.DBItem) error {
	return p.callInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		for _, record := range records {
			// statements for stored procedure are stored in a db
			_, err := tx.ExecContext(ctx, "CALL UpdateOrCreateMetric"+"(@metricType, @metricName, @metricValue)", pgx.NamedArgs{
				"metricType":  record.MetricType.String,
				"metricName":  record.Name.String,
				"metricValue": record.Value.Float64})

			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *postgresDataBase) ReadItem(ctx context.Context, metricType string, metricName string) (*database.DBItem, error) {
	result, err := p.callInTransactionResult(ctx, func(ctx context.Context, tx *sql.Tx) ([]*database.DBItem, error) {
		const command = "SELECT mt.name, m.name, m.value " +
			"FROM metric m " +
			"JOIN metricType mt ON m.typeId = mt.id " +
			"WHERE " +
			"	m.name = @metricName " +
			"	and mt.name = @metricType"

		return p.readRecords(ctx, tx, command, pgx.NamedArgs{
			"metricType": metricType,
			"metricName": metricName,
		})
	})

	if err != nil {
		return nil, err
	}

	count := len(result)
	if count == 0 {
		return nil, nil
	}

	if count > 1 {
		logrus.Errorf("More than one metric in logical primary key: %v, %v", metricType, metricName)
	}

	return result[0], nil
}

func (p *postgresDataBase) ReadAllItems(ctx context.Context) ([]*database.DBItem, error) {
	return p.callInTransactionResult(ctx, func(ctx context.Context, tx *sql.Tx) ([]*database.DBItem, error) {
		const command = "SELECT mt.name, m.name, m.value " +
			"FROM metric m " +
			"JOIN metricType mt on m.typeId = mt.id"

		return p.readRecords(ctx, tx, command)
	})
}

func (p *postgresDataBase) Ping(ctx context.Context) error {
	return p.conn.PingContext(ctx)
}

func (p *postgresDataBase) Close() error {
	return p.conn.Close()
}

func (p *postgresDataBase) callInTransaction(ctx context.Context, action func(context.Context, *sql.Tx) error) error {
	_, err := p.callInTransactionResult(ctx, func(ctx context.Context, tx *sql.Tx) ([]*database.DBItem, error) {
		return nil, action(ctx, tx)
	})

	return err
}

func (p *postgresDataBase) callInTransactionResult(ctx context.Context, action func(context.Context, *sql.Tx) ([]*database.DBItem, error)) ([]*database.DBItem, error) {
	tx, err := p.conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	if err != nil {
		return nil, err
	}

	result, err := action(ctx, tx)
	if err != nil {
		rollbackError := tx.Rollback()
		if rollbackError != nil {
			logrus.Errorf("Fail to rollback transaction: %v", rollbackError)
		}

		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		logrus.Errorf("Fail to commit transaction: %v", err)
		return nil, err
	}

	return result, nil
}

func (p *postgresDataBase) readRecords(ctx context.Context, tx *sql.Tx, command string, args ...any) ([]*database.DBItem, error) {
	rows, err := tx.QueryContext(ctx, command, args...)

	if err != nil {
		return nil, err
	}

	result := []*database.DBItem{}
	for rows.Next() {
		var record database.DBItem
		err = rows.Scan(&record.MetricType, &record.Name, &record.Value)
		if err != nil {
			return nil, err
		}

		result = append(result, &record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}
