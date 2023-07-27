package worker

import (
	"context"
	"fmt"
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

func (w *HardWorker) StartWork(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := w.workFunc(ctx)
			if err != nil {
				fmt.Printf("Hard worker error: %v.", err.Error())
			}
		case <-ctx.Done():
			fmt.Println("The worker stopped working.")
			return
		}
	}
}
