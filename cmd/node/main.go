package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/smallworldsdev/distributed-kv-project/internal/cluster"
	"github.com/smallworldsdev/distributed-kv-project/internal/config"
	"github.com/smallworldsdev/distributed-kv-project/internal/store"
)

// ---- Main ----

func main() {
	cfg := config.LoadConfig()

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	storePath := filepath.Join(cfg.DataDir, cfg.NodeID+".json")
	kvStore, err := store.NewPersistentStore(storePath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	service := cluster.NewService(cfg.NodeID, kvStore, cfg.Peers)

	// Register RPC service
	err = rpc.RegisterName("NodeService", service)
	if err != nil {
		log.Fatal("Error registering RPC service:", err)
	}

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatal("Error starting listener:", err)
	}

	log.Printf("Node %s listening on port %s\n", cfg.NodeID, cfg.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		go rpc.ServeConn(conn)
	}
}
