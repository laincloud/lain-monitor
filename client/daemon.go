package main

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const (
	collectInterval = 600 * time.Second
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
			logger.Warn("runDaemon() has been cancelled.")
		}
	}
}
