package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	internalrpc "github.com/smallworldsdev/distributed-kv-project/internal/rpc"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <host:port>")
	}

	address := os.Args[1]

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("Error connecting to node:", err)
	}

	req := &internalrpc.PingRequest{
		Message: "Hello Node!",
	}

	var res internalrpc.PingResponse

	err = client.Call("NodeService.Ping", req, &res)
	if err != nil {
		log.Fatal("RPC call failed:", err)
	}

	fmt.Println("Response from node:", res.Message)
}
