package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/MlDenis/prometheus_wannabe/internal/converter"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/storage"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/types"
)

type inMemoryStorage struct {
	metricsByType map[string]map[string]metrics.Metric
	lock          sync.RWMutex
}

func NewInMemoryStorage() storage.MetricsStorage {
	return &inMemoryStorage{
		metricsByType: map[string]map[string]metrics.Metric{},
		lock:          sync.RWMutex{},
	}
}

func (s *inMemoryStorage) AddMetricValues(ctx context.Context, metricList []metrics.Metric) ([]metrics.Metric, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	result := make([]metrics.Metric, len(metricList))

	for i, metric := range metricList {
		metricType := metric.GetType()
		typedMetrics, ok := s.metricsByType[metricType]
		if !ok {
			typedMetrics = map[string]metrics.Metric{}
			s.metricsByType[metricType] = typedMetrics
		}

		metricName := metric.GetName()
		currentMetric, ok := typedMetrics[metricName]
		if ok {
			currentMetric.SetValue(metric.GetValue())
		} else {
			currentMetric = metric
			typedMetrics[metricName] = currentMetric
		}
		result[i] = currentMetric
	}

	return result, nil
}

func (s *inMemoryStorage) GetMetricValues(context.Context) (map[string]map[string]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	metricValues := map[string]map[string]string{}
	for metricsType, metricsList := range s.metricsByType {
		values := map[string]string{}
		metricValues[metricsType] = values

		for metricName, metric := range metricsList {
			values[metricName] = metric.GetStringValue()
		}
	}

	return metricValues, nil
}

func (s *inMemoryStorage) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	metricsByName, ok := s.metricsByType[metricType]
	if !ok {
		return nil, fmt.Errorf("get metric with type %s: %w", metricType, metrics.ErrMetricNotFound)
	}

	metric, ok := metricsByName[metricName]
	if !ok {
		return nil, fmt.Errorf("metrics with name %v and types %v not found: %w", metricName, metricType, metrics.ErrMetricNotFound)
	}

	return metric, nil
}

func (s *inMemoryStorage) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.metricsByType = map[string]map[string]metrics.Metric{}

	for metricType, metricsByType := range metricValues {
		metricFactory := types.NewGaugeMetric
		if metricType == "counter" {
			metricFactory = types.NewCounterMetric
		} else if metricType != "gauge" {
			return fmt.Errorf("handle backup metric with type '%s': %w", metricType, metrics.ErrUnknownMetricType)
		}

		for metricName, metricValue := range metricsByType {
			value, err := converter.ToFloat64(metricValue)
			if err != nil {
				return fmt.Errorf("parse float metric value: %w", err)
			}

			metricsList, ok := s.metricsByType[metricType]
			if !ok {
				metricsList = map[string]metrics.Metric{}
				s.metricsByType[metricType] = metricsList
			}

			currentMetric, ok := metricsList[metricName]
			if !ok {
				currentMetric = metricFactory(metricName)
				metricsList[metricName] = currentMetric
			}

			currentMetric.SetValue(value)
		}
	}

	return nil
}
