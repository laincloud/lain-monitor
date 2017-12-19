package main

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

var (
	url = fmt.Sprintf("http://%s/ping", serverAddr)
)

func getTinyDNSStatus(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	resp, err := http.Get(url)
	if err != nil {
		logger.Error("http.Get() failed.",
			zap.String("url", url),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		if err1 := resp.Body.Close(); err1 != nil {
			logger.Error("resp.Body.Close() failed.", zap.Error(err1))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error("http.Get() failed.",
			zap.String("url", url),
			zap.Int("resp.StatusCode", resp.StatusCode),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
