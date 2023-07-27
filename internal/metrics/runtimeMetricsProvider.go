package metrics

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
)

type RuntimeMetricsProviderConfig interface {
	MetricsList() []string
}

type runtimeMetricsProvider struct {
	metrics []Metric
}

func NewRuntimeMetricsProvider(config RuntimeMetricsProviderConfig) MetricsProvider {
	metrics := []Metric{}
	for _, metricName := range config.MetricsList() {
		metrics = append(metrics, NewGaugeMetric(metricName))
	}

	return &runtimeMetricsProvider{metrics: metrics}
}

func (p *runtimeMetricsProvider) Update(context.Context) error {
	fmt.Println("Collect runtime metrics was start")
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	for _, metric := range p.metrics {
		metricName := metric.GetName()
		metricValue, err := getFieldValue(&stats, metricName)
		if err != nil {
			fmt.Printf("Error! Get %v runtime metric value: %v was failure.", metricName, err.Error())
			return err
		}

		metric.SetValue(metricValue)
		fmt.Printf("%V metric was update with %v value.", metricName, metric.GetStringValue())
	}
	return nil
}

func (p *runtimeMetricsProvider) GetMetrics() []Metric {
	return p.metrics
}

func getFieldValue(stats *runtime.MemStats, fieldName string) (float64, error) {
	r := reflect.ValueOf(stats)
	f := reflect.Indirect(r).FieldByName(fieldName)

	value, ok := convertValue(f)
	if !ok {
		return value, fmt.Errorf("field name %v was not found", fieldName)
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
