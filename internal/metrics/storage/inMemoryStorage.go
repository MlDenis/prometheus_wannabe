package storage

import (
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"sync"
)

type InMemoryStorage struct {
	metricsByType map[string]map[string]metrics.Metric
	lock          sync.RWMutex
}

func NewInMemoryStorage() MetricsStorage {
	return &InMemoryStorage{
		metricsByType: map[string]map[string]metrics.Metric{},
		lock:          sync.RWMutex{},
	}
}

func (ms *InMemoryStorage) AddGaugeMetric(metricName string, metricValue float64) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	ms.serviceMetricUpdate("gauge", metricName, metricValue, metrics.NewGaugeMetric)
}

func (ms *InMemoryStorage) AddCounterMetric(metricName string, metricValue int64) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	ms.serviceMetricUpdate("counter", metricName, float64(metricValue), metrics.NewCounterMetric)
}

func (ms *InMemoryStorage) GetMetric(metricType string, metricName string) (string, bool) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	metricsByName, ok := ms.metricsByType[metricType]
	if !ok {
		fmt.Printf("Metrics with type %v not found", metricType)
		return "", false
	}

	metric, ok := metricsByName[metricName]
	if !ok {
		fmt.Printf("Metrics with name %v and type %v not found", metricName, metricType)
		return "", false
	}

	return metric.GetStringValue(), true
}

func (ms *InMemoryStorage) GetAllMetrics() map[string]map[string]string {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	metricValues := map[string]map[string]string{}
	for metricsType, metricsList := range ms.metricsByType {
		values := map[string]string{}
		metricValues[metricsType] = values

		for metricName, metric := range metricsList {
			values[metricName] = metric.GetStringValue()
		}
	}

	return metricValues
}

func (ms *InMemoryStorage) serviceMetricUpdate(metricType string, name string, value float64, metricFactory func(string) metrics.Metric) {
	metricsList, ok := ms.metricsByType[metricType]
	if !ok {
		metricsList = map[string]metrics.Metric{}
		ms.metricsByType[metricType] = metricsList
	}

	currentMetric, ok := metricsList[name]
	if !ok {
		currentMetric = metricFactory(name)
		metricsList[name] = currentMetric
	}

	currentMetric.SetValue(value)
}
