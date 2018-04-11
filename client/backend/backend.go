package backend

import (
	"time"

	"go.uber.org/zap"
)

type Metric struct {
	Path      string
	Value     float64
	Tags      map[string]string
	Timestamp time.Time
	Step      int64
}

type Backend interface {
	Send(metrics []*Metric, logger *zap.Logger)
	Close() error
}
