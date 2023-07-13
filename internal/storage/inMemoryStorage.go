package storage

import (
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"sort"
	"sync"
)

type MemStorage struct {
	metricNames    []string
	gaugeMetrics   map[string]metrics.Metric
	counterMetrics map[string]metrics.Metric
	lock           sync.RWMutex
}

func NewMemStorage() MetricsStorage {
	return &MemStorage{
		metricNames:    []string{},
		gaugeMetrics:   map[string]metrics.Metric{},
		counterMetrics: map[string]metrics.Metric{},
		lock:           sync.RWMutex{},
	}
}

func (ms *MemStorage) AddGaugeMetric(metricName string, value float64) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	serviceMetricUpdate(ms.gaugeMetrics, &ms.metricNames, metricName, value, metrics.NewGaugeMetric)
}

func (ms *MemStorage) AddCounterMetric(name string, value int64) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	serviceMetricUpdate(ms.counterMetrics, &ms.metricNames, name, float64(value), metrics.NewCounterMetric)
}

func (ms *MemStorage) GetMetric(metricType string, metricName string) (string, bool) {
	// TODO: Add implementation
}

func (ms *MemStorage) GetAllMetrics() map[string]map[string]string {
	// TODO: Add implementation
}

func serviceMetricUpdate(metricsMap map[string]metrics.Metric, keys *[]string, metricName string, value float64, metricFactory func(string) metrics.Metric) {
	currentMetric, ok := metricsMap[metricName]
	if !ok {
		currentMetric = metricFactory(metricName)
		metricsMap[metricName] = currentMetric

		*keys = append(*keys, metricName)
		sort.Strings(*keys)
	}

	currentMetric.SetValue(value)
}
