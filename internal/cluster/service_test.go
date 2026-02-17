package cluster

import (
	"testing"

	internalrpc "github.com/smallworldsdev/distributed-kv-project/internal/rpc"
	"github.com/smallworldsdev/distributed-kv-project/internal/store"
)

func TestService(t *testing.T) {
	s := store.NewStore()
	svc := NewService("node1", s, nil)

	// Test Ping
	pingReq := &internalrpc.PingRequest{Message: "Hello"}
	pingRes := &internalrpc.PingResponse{}
	err := svc.Ping(pingReq, pingRes)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if pingRes.Message != "Pong from node node1" {
		t.Errorf("Unexpected Ping response: %s", pingRes.Message)
	}

	// Test Put
	key := "foo"
	value := "bar"
	putReq := &internalrpc.PutRequest{Key: key, Value: value}
	putRes := &internalrpc.PutResponse{}
	err = svc.Put(putReq, putRes)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if !putRes.Success {
		t.Error("Put failed")
	}

	// Test Get
	getReq := &internalrpc.GetRequest{Key: key}
	getRes := &internalrpc.GetResponse{}
	err = svc.Get(getReq, getRes)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !getRes.Found {
		t.Error("Expected key to be found")
	}
	if getRes.Value != value {
		t.Errorf("Expected value %s, got %s", value, getRes.Value)
	}

	// Test Get non-existent
	getReq.Key = "baz"
	err = svc.Get(getReq, getRes)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if getRes.Found {
		t.Error("Expected key baz to not be found")
	}
}
