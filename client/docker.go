package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const (
	dockerReservedMemoryMetric = "lain.cluster.node.docker_reserved_memory"
	swarmInfoURL               = "http://swarm.lain:2376/info"

	// KiB = 1024 Byte
	KiB = 1024
	// MiB = 1024 KiB
	MiB = 1024 * KiB
	// GiB = 1024 MiB
	GiB = 1024 * MiB
	// TiB = 1024 GiB
	TiB = 1024 * GiB
	// PiB = 1024 TiB
	PiB = 1024 * TiB
)

var (
	binaryMap = map[string]int64{
		"B":   1,
		"KiB": KiB,
		"MiB": MiB,
		"GiB": GiB,
		"TiB": TiB,
		"PiB": PiB,
	}
)

type swarmInfoResponse struct {
	SystemStatus [][2]string `json:"SystemStatus"`
}

type swarmInfo struct {
	Nodes []swarmNodeInfo
}

type swarmNodeInfo struct {
	Name           string
	ReservedMemory int64 // Unit: B
	TotalMemory    int64 // Unit: B
}

func collectDockerReservedMemory(graphite *Graphite, logger *zap.Logger) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warn("collectDockerReservedMemory recovered.", zap.Any("r", r))
		}
	}()

	resp, err := http.Get(swarmInfoURL)
	if err != nil {
		logger.Error("http.Get() failed.",
			zap.String("url", url),
			zap.Error(err),
		)
		return
	}

	var data swarmInfoResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.Error("json.NewDecoder().Decode failed.",
			zap.Error(err),
		)
		return
	}

	swarmInfo, err := parseSwarmInfo(data)
	if err != nil {
		logger.Error("parseSwarmInfo() failed.",
			zap.Error(err),
		)
		return
	}

	for _, node := range swarmInfo.Nodes {
		graphite.Send(node.Name, dockerReservedMemoryMetric, node.ReservedMemory, logger)
	}
}

func parseSwarmInfo(resp swarmInfoResponse) (*swarmInfo, error) {
	offset := 0
	for ; offset < len(resp.SystemStatus); offset++ {
		if resp.SystemStatus[offset][0] == "Nodes" {
			break
		}
	}

	nodeCount, err := strconv.Atoi(resp.SystemStatus[offset][1])
	if err != nil {
		return nil, err
	}

	nodes := make([]swarmNodeInfo, nodeCount)
	offset++
	for i := 0; i < nodeCount; i++ {
		memories := strings.Split(resp.SystemStatus[offset+5][1], "/")
		reservedMemory, err := parseBytesSize(memories[0])
		if err != nil {
			return nil, err
		}
		totalMemory, err := parseBytesSize(memories[1])
		if err != nil {
			return nil, err
		}
		nodes[i] = swarmNodeInfo{
			Name:           strings.TrimSpace(resp.SystemStatus[offset][0]),
			ReservedMemory: reservedMemory,
			TotalMemory:    totalMemory,
		}
		offset += 9
	}
	return &swarmInfo{
		Nodes: nodes,
	}, nil
}

func parseBytesSize(size string) (int64, error) {
	parts := strings.Split(strings.TrimSpace(size), " ")
	memory, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	factor, ok := binaryMap[parts[1]]
	if !ok {
		return 0, fmt.Errorf("invalid size unit %q in %q", parts[1], size)
	}
	return int64(memory * float64(factor)), nil
}
