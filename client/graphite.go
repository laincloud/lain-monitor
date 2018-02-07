package main

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

const timeout = 3 * time.Second

// Graphite is a graphite client
type Graphite struct {
	addr string
}

// NewGraphite create a new Graphite
func NewGraphite(addr string) *Graphite {
	return &Graphite{
		addr: addr,
	}
}

// Send sends metric value to hostname
func (g *Graphite) Send(metrics []GraphiteMetric, logger *zap.Logger) {
	conn, err := net.DialTimeout("tcp", g.addr, timeout)
	if err != nil {
		logger.Error("net.DialTimeout() failed.", zap.Error(err))
		return
	}
	defer conn.Close()
	logger.Info("net.DialTimeout() succeed.", zap.String("addr", g.addr))

	for _, m := range metrics {
		if _, err := fmt.Fprintf(conn, "%s %v %d\n", m.Path, m.Value, m.Timestamp.Unix()); err != nil {
			logger.Error("Graphite.Send() failed.", zap.Any("metric", m), zap.Error(err))
			return
		}
	}

	logger.Info("Graphite.Send() succeed.", zap.Any("metrics", metrics))
}

// GraphiteMetric is the metric sent to graphite
type GraphiteMetric struct {
	Path      string
	Value     interface{}
	Timestamp time.Time
}
