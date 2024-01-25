package runtime

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/types"

	"github.com/sirupsen/logrus"
)

type runtimeMetricsProviderConfig interface {
	MetricsList() []string
}

type runtimeMetricsProvider struct {
	metrics []metrics.Metric
}

func NewRuntimeMetricsProvider(config runtimeMetricsProviderConfig) metrics.MetricsProvider {
	metricNames := config.MetricsList()
	metricsList := make([]metrics.Metric, len(metricNames))
	for i, metricName := range metricNames {
		metricsList[i] = types.NewGaugeMetric(metricName)
	}

	return &runtimeMetricsProvider{metrics: metricsList}
}

func (p *runtimeMetricsProvider) Update(context.Context) error {
	logrus.Info("Start collect runtime metrics")
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	var err error
	for _, metric := range p.metrics {
		metricName := metric.GetName()
		metricValue, metricErr := getFieldValue(&stats, metricName)
		if metricErr != nil {
			err = logger.WrapError(fmt.Sprintf("get %s runtime metric value", metricName), metricErr)
			logrus.Error(err)
			continue
		}

		metric.SetValue(metricValue)
	}

	return err
}

func (p *runtimeMetricsProvider) GetMetrics() <-chan metrics.Metric {
	result := make(chan metrics.Metric)
	go func() {
		defer close(result)
		for _, metric := range p.metrics {
			result <- metric
		}
	}()

	return result
}

func getFieldValue(stats *runtime.MemStats, fieldName string) (float64, error) {
	r := reflect.ValueOf(stats)
	f := reflect.Indirect(r).FieldByName(fieldName)

	value, ok := convertValue(f)
	if !ok {
		return value, logger.WrapError(fmt.Sprintf("get field with name %s", fieldName), metrics.ErrFieldNameNotFound)
	}

	return value, nil
}

func convertValue(value reflect.Value) (float64, bool) {
	if value.CanInt() {
		return float64(value.Int()), true
	}
	if value.CanUint() {
		return float64(value.Uint()), true
	}
	if value.CanFloat() {
		return value.Float(), true
	}

	return 0, false
}
