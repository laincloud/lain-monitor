package backend

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
func NewGraphite(addr string) (Backend, error) {
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		return nil, err
	}

	return &Graphite{
		addr: addr,
		conn: conn,
	}, nil
}

// Send sends metric value to hostname
func (g *Graphite) Send(metrics []*Metric, logger *zap.Logger) {
	for _, m := range metrics {
		if _, err := fmt.Fprintf(g.conn, "%s %v %d\n", m.Path, m.Value, m.Timestamp.Unix()); err != nil {
			logger.Error("Graphite.Send() failed.", zap.Any("metric", m), zap.Error(err))
			return
		}
	}

	logger.Info("Graphite.Send() succeed.", zap.Any("metrics", metrics))
}

// Close close the underlying connection
func (g *Graphite) Close() error {
	return g.conn.Close()
}
