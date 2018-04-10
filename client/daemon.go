package main

import (
	"context"
	"time"

	"github.com/laincloud/lain-monitor/client/backend"
	"go.uber.org/zap"
)

const (
	collectInterval = 3 * 60
)

func runDaemon(ctx context.Context, bd backend.Backend, logger *zap.Logger) {
	ticker := time.NewTicker(collectInterval * time.Second)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			logger.Info("collectDockerReservedMemory...", zap.Time("now", now))
			collectDockerReservedMemory(bd, collectInterval, logger)
			logger.Info("collectDockerReservedMemory done.", zap.Time("now", now))
		case <-ctx.Done():
			logger.Info("runDaemon() has been cancelled.")
			return
		}
	}
}
