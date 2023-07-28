package gopsutil

import (
	"context"
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/test"
	"github.com/stretchr/testify/assert"
	"runtime"
	"strings"
	"testing"
)

func TestGopsutilMetricsProvider_GetMetrics(t *testing.T) {
	expected := []string{
		"FreeMemory",
		"TotalMemory",
	}
	for i := 1; i < runtime.NumCPU()+1; i++ {
		expected = append(expected, fmt.Sprintf("CPUutilization%d", i))
	}

	provider := NewGopsutilMetricsProvider()
	actual := test.ChanToArray(provider.GetMetrics())

	assert.Len(t, expected, len(actual))
	for _, actualMetric := range actual {
		assert.Contains(t, expected, actualMetric.GetName())
		assert.Equal(t, actualMetric.GetStringValue(), "0")
	}
}

func TestGopsutilMetricsProvider_Update(t *testing.T) {
	ctx := context.Background()
	provider := NewGopsutilMetricsProvider()
	assert.NoError(t, provider.Update(ctx))

	actual := test.ChanToArray(provider.GetMetrics())

	cpuChecked := false
	for _, actualMetric := range actual {
		name := actualMetric.GetName()
		if name == "FreeMemory" || name == "TotalMemory" {
			assert.NotEqual(t, actualMetric.GetStringValue(), "0")
		}

		if strings.HasPrefix(name, "CPUutilization") && !cpuChecked {
			cpuChecked = actualMetric.GetStringValue() != "0"
		}
	}
	assert.True(t, cpuChecked)
}
