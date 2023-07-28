package model

import (
	"github.com/MlDenis/prometheus_wannabe/internal/hash"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/types"
)

type UnknownMetricTypeError struct {
	UnknownType string
}

func (e *UnknownMetricTypeError) Error() string {
	return "unknown metric type: " + e.UnknownType
}

type MetricsConverterConfig interface {
	SignMetrics() bool
}

type MetricsConverter struct {
	signer      *hash.Signer
	signMetrics bool
}

func NewMetricsConverter(conf MetricsConverterConfig, signer *hash.Signer) *MetricsConverter {
	return &MetricsConverter{
		signMetrics: conf.SignMetrics(),
		signer:      signer,
	}
}

func (c *MetricsConverter) ToModelMetric(metric metrics.Metric) (*Metrics, error) {
	modelMetric := &Metrics{
		ID:    metric.GetName(),
		MType: metric.GetType(),
	}

	metricValue := metric.GetValue()
	switch modelMetric.MType {
	case "counter":
		counterValue := int64(metricValue)
		modelMetric.Delta = &counterValue
	case "gauge":
		modelMetric.Value = &metricValue
	default:
		logger.ErrorFormat("unknown metric type: %v", modelMetric.MType)
		return nil, &UnknownMetricTypeError{UnknownType: modelMetric.MType}
	}

	if c.signMetrics {
		signature, err := c.signer.GetSignString(metric)
		if err != nil {
			return nil, logger.WrapError("get signature string", err)
		}

		modelMetric.Hash = signature
	}

	return modelMetric, nil
}

func (c *MetricsConverter) FromModelMetric(modelMetric *Metrics) (metrics.Metric, error) {
	var metric metrics.Metric
	var value float64

	switch modelMetric.MType {
	case "counter":
		if modelMetric.Delta == nil {
			return nil, logger.WrapError("convert metric", metrics.ErrMetricValueMissed)
		}

		metric = types.NewCounterMetric(modelMetric.ID)
		value = float64(*modelMetric.Delta)
	case "gauge":
		if modelMetric.Value == nil {
			return nil, logger.WrapError("convert metric", metrics.ErrMetricValueMissed)
		}

		metric = types.NewGaugeMetric(modelMetric.ID)
		value = *modelMetric.Value
	default:
		logger.ErrorFormat("unknown metric type: %v", modelMetric.MType)
		return nil, &UnknownMetricTypeError{UnknownType: modelMetric.MType}
	}

	metric.SetValue(value)

	if c.signMetrics && modelMetric.Hash != "" {
		ok, err := c.signer.CheckSign(metric, modelMetric.Hash)
		if err != nil {
			return nil, logger.WrapError("check signature", err)
		}

		if !ok {
			return nil, logger.WrapError("check signature", metrics.ErrInvalidSignature)
		}
	}

	return metric, nil
}
