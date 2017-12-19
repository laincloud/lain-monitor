package main

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

type config struct {
	GraphiteAddr string `json:"graphite_addr"`
}

func newConfig(filename string, logger *zap.Logger) (*config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err1 := file.Close(); err1 != nil {
			logger.Error("file.Close() failed.", zap.Error(err1))
		}
	}()

	var c config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
