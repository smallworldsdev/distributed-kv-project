package main

import (
	"log"
	"net"
	"net/rpc"
	"os"

	internalrpc "github.com/smallworldsdev/distributed-kv-project/internal/rpc"
)

// ---- RPC Service ----

type NodeService struct {
	NodeID string
}

func (n *NodeService) Ping(req *internalrpc.PingRequest, res *internalrpc.PingResponse) error {
	log.Printf("Received ping: %s", req.Message)

	res.Message = "Pong from node " + n.NodeID
	return nil
}

// ---- Main ----

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = port
	}

	service := &NodeService{
		NodeID: nodeID,
	}

	// Register RPC service
	err := rpc.Register(service)
	if err != nil {
		log.Fatal("Error registering RPC service:", err)
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Error starting listener:", err)
	}

	log.Printf("Node %s listening on port %s\n", nodeID, port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		go rpc.ServeConn(conn)
	}
}
