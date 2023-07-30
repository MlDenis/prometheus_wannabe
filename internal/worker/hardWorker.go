package worker

import (
	"context"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"time"
)

type HardWorker struct {
	workFunc func(ctx context.Context) error
}

func NewHardWorker(workFunc func(ctx context.Context) error) HardWorker {
	return HardWorker{
		workFunc: workFunc,
	}
}

func (w *HardWorker) StartWork(ctx context.Context, inv int) {
	interval := time.Duration(inv) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := w.workFunc(ctx)
			if err != nil {
				logger.ErrorFormat("periodic worker error: %v", err)
			}
		case <-ctx.Done():
			logger.ErrorFormat("periodic worker canceled")
			return
		}
	}
}
