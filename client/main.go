package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/laincloud/lain-monitor/client/backend"
	"github.com/laincloud/lain-monitor/common"
	"go.uber.org/zap"
)

const (
	port       = 8080
	serverAddr = "server-1:8080"
)

var (
	configFile = flag.String("config", "", "configuration file")
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("zap.NewProduction() failed, error: %v.", err)
	}
	defer logger.Sync()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)

	flag.Parse()
	if *configFile == "" {
		logger.Fatal("config is required")
	}

	c, err := newConfig(*configFile, logger)
	if err != nil {
		logger.Fatal("newConfig() failed.", zap.String("filename", *configFile), zap.Error(err))
	}
	var bd backend.Backend

	if c.BackendType == "open-falcon" {
		bd, err = backend.NewOpenFalconBackend(c.OpenFalconAddr)
	} else if c.BackendType == "graphite" {
		bd, err = backend.NewGraphite(c.GraphiteAddr)
	} else {
		logger.Fatal("unknow backend type")
	}

	if err != nil {
		logger.Fatal("NewGraphite() failed.", zap.Error(err))
	}
	defer bd.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond)
		logger.Info("Context has been cancelled.")
	}()

	go runDaemon(ctx, bd, logger)
	if err := runHealthCheckers(ctx, bd, logger); err != nil {
		logger.Fatal("Create health checkers failed", zap.Error(err))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tinydns_status", common.Handle(getTinyDNSStatus, logger))
	mux.HandleFunc("/ping", common.Handle(common.Ping, logger))
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("server.ListenAndServe() failed.", zap.String("Addr", server.Addr), zap.Error(err))
		}
	}()
	logger.Info("server.ListenAndServe()...", zap.String("Addr", server.Addr))

	<-quit
	logger.Info("Shutting down...")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error("server.Shutdown() failed.", zap.Error(err))
	} else {
		logger.Info("Server has been shutdown.")
	}
}
