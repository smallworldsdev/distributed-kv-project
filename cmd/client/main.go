package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	internalrpc "github.com/smallworldsdev/distributed-kv-project/internal/rpc"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run main.go <host:port> <command> [args...]")
	}

	address := os.Args[1]
	command := os.Args[2]

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("Error connecting to node:", err)
	}

	switch command {
	case "ping":
		req := &internalrpc.PingRequest{
			Message: "Hello Node!",
		}
		var res internalrpc.PingResponse
		err = client.Call("NodeService.Ping", req, &res)
		if err != nil {
			log.Fatal("RPC call failed:", err)
		}
		fmt.Println("Response from node:", res.Message)

	case "get":
		if len(os.Args) < 4 {
			log.Fatal("Usage: get <key>")
		}
		key := os.Args[3]
		req := &internalrpc.GetRequest{
			Key: key,
		}
		var res internalrpc.GetResponse
		err = client.Call("NodeService.Get", req, &res)
		if err != nil {
			log.Fatal("RPC call failed:", err)
		}
		if res.Found {
			fmt.Printf("Value: %s\n", res.Value)
		} else {
			fmt.Println("Key not found")
		}

	case "put":
		if len(os.Args) < 5 {
			log.Fatal("Usage: put <key> <value>")
		}
		key := os.Args[3]
		value := os.Args[4]
		req := &internalrpc.PutRequest{
			Key:   key,
			Value: value,
		}
		var res internalrpc.PutResponse
		err = client.Call("NodeService.Put", req, &res)
		if err != nil {
			log.Fatal("RPC call failed:", err)
		}
		if res.Success {
			fmt.Println("Put successful")
		} else {
			fmt.Println("Put failed")
		}

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
