package storage

type MetricsStorage interface {
	AddGaugeMetric(name string, value float64)
	AddCounterMetric(name string, value int64)
	GetMetric(metricType string, metricName string) (string, bool)
	GetAllMetrics() map[string]map[string]string
}
