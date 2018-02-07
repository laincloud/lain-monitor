package main

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const (
	collectInterval = 3 * time.Minute
)

func runDaemon(ctx context.Context, graphite *Graphite, logger *zap.Logger) {
	ticker := time.NewTicker(collectInterval)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			logger.Info("collectDockerReservedMemory...", zap.Time("now", now))
			collectDockerReservedMemory(graphite, logger)
			logger.Info("collectDockerReservedMemory done.", zap.Time("now", now))
		case <-ctx.Done():
			logger.Info("runDaemon() has been cancelled.")
			return
		}
	}
}
