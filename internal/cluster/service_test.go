package cluster

import (
	"testing"

	"github.com/smallworldsdev/distributed-kv-project/internal/rpc"
	"github.com/smallworldsdev/distributed-kv-project/internal/store"
)

func TestService(t *testing.T) {
	s := store.NewStore()
	svc := NewService("node1", s, nil)

	// Test Ping
	pingReq := &rpc.PingRequest{Message: "Hello"}
	pingRes := &rpc.PingResponse{}
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
	putReq := &rpc.PutRequest{Key: key, Value: value}
	putRes := &rpc.PutResponse{}
	err = svc.Put(putReq, putRes)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if !putRes.Success {
		t.Error("Put failed")
	}

	// Test Get
	getReq := &rpc.GetRequest{Key: key}
	getRes := &rpc.GetResponse{}
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
