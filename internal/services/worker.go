package services

import (
	"context"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/logger"
)

// Worker ...
type Worker struct{}

// LaunchWorker launches a goroutine that will initialize the worker and execute tasks
func LaunchWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker received shutdown signal, finishing up...")
			time.Sleep(500 * time.Millisecond)
			logger.Info("Worker finished cleanup.")
			return
		default:
		}
		logger.Info("Worker is running")
		time.Sleep(time.Second)
	}
}
