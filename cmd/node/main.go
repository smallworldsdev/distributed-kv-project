package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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

	// Channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Node %s listening on port %s\n", cfg.NodeID, cfg.Port)

	// Start background tasks (heartbeat, monitoring)
	service.Start()

	// Run server in goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Don't log error if listener is closed during shutdown
				select {
				case <-sigChan:
					return
				default:
					log.Println("Connection error:", err)
				}
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %s. Shutting down...", sig)

	// Graceful shutdown logic
	service.Shutdown()
	listener.Close()
	log.Println("Listener closed.")
	// Add store.Close() here if implemented in future
	log.Println("Shutdown complete.")
}
