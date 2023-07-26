package metrics

import (
	"github.com/MlDenis/prometheus_wannabe/internal/converter"
	"sync"
)

type counterMetric struct {
	metricName  string
	metricValue int64
	lock        sync.RWMutex
}

func NewCounterMetric(metricName string) Metric {
	return &counterMetric{
		metricName:  metricName,
		metricValue: 0,
		lock:        sync.RWMutex{},
	}
}

func (m *counterMetric) GetType() string {
	return "counter"
}

func (m *counterMetric) GetName() string {
	return m.metricName
}

func (m *counterMetric) GetStringValue() string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return converter.IntToString(m.metricValue)
}

func (m *counterMetric) SetValue(addedValue float64) {
	m.setMetricValue(m.metricValue + int64(addedValue))
}

func (m *counterMetric) JumpToTheOriginalState() {
	m.setMetricValue(0)
}

func (m *counterMetric) setMetricValue(newMetricValue int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metricValue = newMetricValue
}
