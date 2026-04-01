package config

import (
	"os"
	"strings"
)

type Config struct {
	NodeID      string
	Port        string
	Peers       []string
	DataDir     string
	MetricsPort string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "node-" + port
	}

	peersEnv := os.Getenv("PEERS")
	var peers []string
	if peersEnv != "" {
		peers = strings.Split(peersEnv, ",")
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	return &Config{
		NodeID:      nodeID,
		Port:        port,
		Peers:       peers,
		DataDir:     dataDir,
		MetricsPort: metricsPort,
	}
}
