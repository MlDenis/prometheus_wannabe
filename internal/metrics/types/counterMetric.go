package types

import (
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/converter"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"hash"
)

type counterMetric struct {
	name  string
	value int64
}

func NewCounterMetric(name string) metrics.Metric {
	return &counterMetric{
		name: name,
	}
}

func (m *counterMetric) GetType() string {
	return "counter"
}

func (m *counterMetric) GetName() string {
	return m.name
}

func (m *counterMetric) GetValue() float64 {

	return float64(m.value)
}

func (m *counterMetric) GetStringValue() string {

	return converter.IntToString(m.value)
}

func (m *counterMetric) SetValue(value float64) float64 {
	return m.setValue(m.value + int64(value))
}

func (m *counterMetric) JumpToTheOriginalState() {
	m.setValue(0)
}

func (m *counterMetric) GetHash(hash hash.Hash) ([]byte, error) {

	_, err := hash.Write([]byte(fmt.Sprintf("%s:counter:%d", m.name, m.value)))
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func (m *counterMetric) setValue(value int64) float64 {
	m.value = value
	return float64(m.value)
}
