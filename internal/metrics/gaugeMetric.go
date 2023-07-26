package metrics

import (
	"github.com/MlDenis/prometheus_wannabe/internal/converter"
	"sync"
)

type gaugeMetric struct {
	metricName  string
	metricValue float64
	lock        sync.RWMutex
}

func NewGaugeMetric(metricName string) Metric {
	return &gaugeMetric{
		metricName:  metricName,
		metricValue: 0,
		lock:        sync.RWMutex{},
	}
}

func (m *gaugeMetric) GetType() string {
	return "gauge"
}

func (m *gaugeMetric) GetName() string {
	return m.metricName
}

func (m *gaugeMetric) GetStringValue() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return converter.FloatToString(m.metricValue)
}

func (m *gaugeMetric) SetValue(newValue float64) {
	m.setMetricValue(newValue)
}

func (m *gaugeMetric) JumpToTheOriginalState() {
}

func (m *gaugeMetric) setMetricValue(newValue float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metricValue = newValue
}
