package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/MlDenis/prometheus_wannabe/internal/test"
)

func TestPeriodicWorker_CloseContext(t *testing.T) {
	wasCalled := false
	ctx, cancel := context.WithCancel(context.Background())

	worker := NewHardWorker(func(context.Context) error {
		wasCalled = true
		return nil
	})

	cancel()
	worker.StartWork(ctx, 1*time.Millisecond)
	assert.False(t, wasCalled)
}

func TestPeriodicWorker_SuccessCall(t *testing.T) {
	wasCalled := false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := NewHardWorker(func(context.Context) error {
		if !wasCalled {
			wasCalled = true
			return test.ErrTest
		}

		cancel()
		return nil
	})

	worker.StartWork(ctx, 1*time.Millisecond)
	assert.True(t, wasCalled)
}
