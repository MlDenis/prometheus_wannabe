package types

import (
	"fmt"
	"hash"

	"github.com/MlDenis/prometheus_wannabe/internal/converter"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
)

type gaugeMetric struct {
	name  string
	value float64
}

func NewGaugeMetric(name string) metrics.Metric {
	return &gaugeMetric{
		name: name,
	}
}

func (m *gaugeMetric) GetType() string {
	return "gauge"
}

func (m *gaugeMetric) GetName() string {
	return m.name
}

func (m *gaugeMetric) GetValue() float64 {

	return m.value
}

func (m *gaugeMetric) GetStringValue() string {

	return converter.FloatToString(m.value)
}

func (m *gaugeMetric) SetValue(value float64) float64 {
	m.value = value

	return m.value
}

func (m *gaugeMetric) ResetState() {
}

func (m *gaugeMetric) GetHash(hash hash.Hash) ([]byte, error) {

	_, err := hash.Write([]byte(fmt.Sprintf("%s:gauge:%f", m.name, m.value)))
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
