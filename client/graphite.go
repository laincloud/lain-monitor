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
	conn net.Conn
}

// NewGraphite create a new Graphite
func NewGraphite(addr string, logger *zap.Logger) (*Graphite, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	logger.Info("Dial graphite addr succeed.", zap.String("addr", addr))

	return &Graphite{
		addr: addr,
		conn: conn,
	}, nil
}

// Send sends metric value to hostname
func (g *Graphite) Send(hostname, metric string, value interface{}, logger *zap.Logger) {
	if _, err := fmt.Fprintf(g.conn, "%s.%s %v %d\n", hostname, metric, value, time.Now().Unix()); err != nil {
		logger.Error("Graphite.Send() failed.", zap.String("hostname", hostname), zap.String("metric", metric), zap.Any("value", value), zap.Error(err))

		if err = g.conn.Close(); err != nil {
			logger.Error("g.conn.Close() failed", zap.Error(err))
		}

		if conn, err := net.DialTimeout("tcp", g.addr, timeout); err != nil {
			logger.Error("net.DialTimeout() failed.", zap.Error(err))
		} else {
			g.conn = conn
		}
		return
	}

	logger.Info("Graphite.Send() succeed.", zap.String("hostname", hostname), zap.String("metric", metric), zap.Any("value", value))
}

// Close closes the underlying connection
func (g *Graphite) Close() error {
	return g.conn.Close()
}
