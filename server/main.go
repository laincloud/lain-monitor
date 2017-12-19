package main

import (
	"context"
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
	port = 8080
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("zap.NewProduction() failed, error: %v.", err)
	}
	defer logger.Sync()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)

	mux := http.NewServeMux()
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
