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
	graphite, err := newGraphite(c.GraphiteAddr)
	if err != nil {
		logger.Fatal("newGraphite() failed.", zap.String("addr", c.GraphiteAddr), zap.Error(err))
	}
	defer func() {
		if err1 := graphite.close(); err1 != nil {
			logger.Error("graphite.close() failed.", zap.Error(err1))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runDaemon(ctx, graphite, logger)

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
	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error("server.Shutdown() failed.", zap.Error(err))
		}
	}()
	logger.Info("server.ListenAndServe()...", zap.String("Addr", server.Addr))

	<-quit
	logger.Info("Shutting down...")
}
