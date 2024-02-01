package metrics

import "github.com/MlDenis/prometheus_wannabe/internal/hash"

type Metric interface {
	hash.HashHolder

	GetName() string
	GetType() string
	GetValue() float64
	GetStringValue() string
	SetValue(value float64) float64
	JumpToTheOriginalState()
}
