package common

import (
	"net/http"

	"go.uber.org/zap"
)

// Ping is for health check
func Ping(w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.Error("w.Write() failed.", zap.Error(err))
	}
}
