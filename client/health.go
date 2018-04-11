package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/laincloud/lain-monitor/client/backend"
	"go.uber.org/zap"
)

const (
	TIMEOUT              = 5
	HEALTHCHECK_INTERVAL = 60

	deploydMetric = "lain.deployd.health"
	deploydURL    = "http://deployd.lain:9003/api/status"

	consoleMetric = "lain.console.health"
	consoleURL    = "http://console.lain/"

	etcdMetric = "lain.etcd.health"
	etcdURL    = "http://etcd.lain:4001/health"

	swarmMetric = "lain.swarm.health"
	swarmURL    = "http://swarm.lain:2376/_ping"

	consulMetric = "lain.consul.health"
	consulAddr   = "consul.lain:8500"
)

type HealthChecker interface {
	Check(logger *zap.Logger) *backend.Metric
}

type urlHealthChecker struct {
	metricName string
	url        string
}

type etcdHealthChecker struct {
	metricName string
	url        string
}

type consulHealthChecker struct {
	metricName string
	addr       string
	client     *api.Client
}

func _buildPacket(name string, isAlive bool) *backend.Metric {
	health := 1
	if isAlive == false {
		health = 0
	}
	return &backend.Metric{
		Path:      name,
		Value:     float64(health),
		Tags:      map[string]string{"cluster": cfg.ClusterName},
		Timestamp: time.Now(),
		Step:      HEALTHCHECK_INTERVAL,
	}
}

func newConsulHealthChecker(metric, addr string) (HealthChecker, error) {
	config := &api.Config{
		Address:   addr,
		Scheme:    "http",
		Transport: cleanhttp.DefaultPooledTransport(),
	}
	consulClient, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &consulHealthChecker{
		metricName: metric,
		addr:       addr,
		client:     consulClient,
	}, nil
}

func (ckr *urlHealthChecker) Check(logger *zap.Logger) *backend.Metric {
	timeout := TIMEOUT * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(ckr.url)
	isAlive := true
	if err != nil || resp.StatusCode != 200 {
		isAlive = false
	}
	return _buildPacket(ckr.metricName, isAlive)
}

func (ckr *etcdHealthChecker) Check(logger *zap.Logger) *backend.Metric {
	timeout := TIMEOUT * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(ckr.url)
	isAlive := true
	if err != nil || resp.StatusCode != 200 {
		isAlive = false
		return _buildPacket(ckr.metricName, isAlive)
	}
	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.Error("failed to unmarshal from etcd resp.", zap.Error(err))
		isAlive = false
		return _buildPacket(ckr.metricName, isAlive)
	}
	if health, ok := data["health"]; !ok || health == "false" {
		isAlive = false
	}
	return _buildPacket(ckr.metricName, isAlive)
}

func (ckr *consulHealthChecker) Check(logger *zap.Logger) *backend.Metric {
	leader, err := ckr.client.Status().Leader()
	isHealth := true
	if err != nil || len(leader) == 0 {
		isHealth = false
	}
	return _buildPacket(ckr.metricName, isHealth)
}

func runHealthCheckers(ctx context.Context, bd backend.Backend, logger *zap.Logger) error {
	var checkers []HealthChecker
	urls := []string{
		deploydURL,
		consoleURL,
		swarmURL,
	}
	names := []string{
		deploydMetric,
		consoleMetric,
		swarmMetric,
	}
	var packets []*backend.Metric

	for i := 0; i < len(names); i++ {
		url := urls[i]
		name := names[i]
		checkers = append(checkers, &urlHealthChecker{name, url})
	}
	checkers = append(checkers, &etcdHealthChecker{etcdMetric, etcdURL})

	consulCkr, err := newConsulHealthChecker(consulMetric, consulAddr)
	if err != nil {
		return err
	}
	checkers = append(checkers, consulCkr)
	go func() {
		ticker := time.NewTicker(HEALTHCHECK_INTERVAL)
		defer ticker.Stop()
		for {
			select {
			case now := <-ticker.C:
				logger.Info("health check...", zap.Time("now", now))
				for _, ckr := range checkers {
					packets = append(packets, ckr.Check(logger))
				}
				bd.Send(packets, logger)
				logger.Info("health check done.", zap.Time("now", now))
			case <-ctx.Done():
				logger.Info("health checker has been cancelled.")
				return
			}
		}
	}()
	return nil
}
