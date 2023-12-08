package stub

import (
	"context"

	"github.com/MlDenis/prometheus_wannabe/internal/database"
)

type StubDataBase struct{}

func (s *StubDataBase) UpdateItems(context.Context, []*database.DBItem) error {
	return nil
}

func (s *StubDataBase) ReadItem(context.Context, string, string) (*database.DBItem, error) {
	return nil, nil
}

func (s *StubDataBase) ReadAllItems(context.Context) ([]*database.DBItem, error) {
	return nil, nil
}

func (s *StubDataBase) Ping(context.Context) error {
	return nil
}

func (s *StubDataBase) Close() error {
	return nil
}
