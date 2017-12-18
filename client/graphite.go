package main

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

// Graphite is a graphite client
type Graphite struct {
	addr string
	conn net.Conn
}

func newGraphite(addr string) (*Graphite, error) {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return nil, err
	}
	return &Graphite{
		addr: addr,
		conn: conn,
	}, nil
}

func (g *Graphite) send(hostname, metric string, value interface{}, logger *zap.Logger) {
	if _, err := g.conn.Write([]byte(fmt.Sprintf("%s.%s %v %d\n", hostname, metric, value, time.Now().Unix()))); err != nil {
		logger.Error("graphite.send() failed.", zap.String("hostname", hostname), zap.String("metric", metric), zap.Any("value", value), zap.Error(err))
		return
	}

	logger.Info("graphite.send() succeed.", zap.String("hostname", hostname), zap.String("metric", metric), zap.Any("value", value))
}

func (g *Graphite) close() error {
	return g.conn.Close()
}
