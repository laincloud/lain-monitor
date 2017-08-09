package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

const (
	serverAddr  = "server-1"
	tinydnsOK   = "OK"
	tinydnsDown = "Down"
)

var (
	port = flag.Int("port", 8080, "The port to listen")
)

func main() {
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("zap.NewProduction() failed, error: %v.", err)
	}
	defer logger.Sync()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: handle(ping, logger),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("server.ListenAndServe() failed.", zap.Error(err))
		}
	}()
	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error("server.Shutdown() failed.", zap.Error(err))
		}
	}()
	logger.Info("server.ListenAndServe()...", zap.String("Addr", ":8080"))

	<-quit
	logger.Info("Shutting down...")
}

func handle(f func(w http.ResponseWriter, r *http.Request, logger *zap.Logger), logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newLogger := logger.With(zap.String("RequestID", newRequestID()))
		newLogger.Info("Receive a new request",
			zap.String("RemoteAddr", r.RemoteAddr),
			zap.String("Method", r.Method),
			zap.String("URL", r.URL.String()),
			zap.Any("Header", r.Header),
		)
		f(w, r, newLogger)
		newLogger.Info("Response is sent.",
			zap.String("RemoteAddr", r.RemoteAddr),
			zap.String("Method", r.Method),
			zap.String("URL", r.URL.String()),
			zap.Any("Header", r.Header),
		)
	}
}

func ping(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	err := exec.Command("ping", "-c", "1", serverAddr).Run()
	if err != nil {
		if _, err := w.Write([]byte(tinydnsDown)); err != nil {
			logger.Error("w.Write() failed.",
				zap.String("Message", tinydnsDown),
				zap.Error(err),
			)
		}
	}

	if _, err := w.Write([]byte(tinydnsOK)); err != nil {
		logger.Error("w.Write() failed.",
			zap.String("Message", tinydnsOK),
			zap.Error(err),
		)
	}
}

func newRequestID() string {
	bs := make([]byte, 16)
	if _, err := rand.Read(bs); err != nil {
		return "0"
	}

	return hex.EncodeToString(bs)
}
