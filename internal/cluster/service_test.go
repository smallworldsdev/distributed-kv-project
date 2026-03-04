package cluster

import (
	"net"
	"net/rpc"
	"testing"

	internalrpc "github.com/smallworldsdev/distributed-kv-project/internal/rpc"
	"github.com/smallworldsdev/distributed-kv-project/internal/store"
)

func TestService(t *testing.T) {
	s := store.NewStore()
	svc := NewService("node1", s, nil)
	svc.leaderID = "node1"

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

func TestPutRejectsWhenLeaderUnknown(t *testing.T) {
	s := store.NewStore()
	svc := NewService("node1", s, nil)

	req := &internalrpc.PutRequest{Key: "k1", Value: "v1"}
	res := &internalrpc.PutResponse{}
	err := svc.Put(req, res)
	if err == nil {
		t.Fatal("expected error when leader is unknown")
	}
	if res.Success {
		t.Fatal("expected put response to be unsuccessful")
	}
	if _, ok := s.Get("k1"); ok {
		t.Fatal("follower must not apply write when leader is unknown")
	}
}

func TestPutForwardsToLeader(t *testing.T) {
	leaderStore := store.NewStore()
	leaderSvc := NewService("node3", leaderStore, nil)
	leaderSvc.leaderID = "node3"

	server := rpc.NewServer()
	if err := server.RegisterName("NodeService", leaderSvc); err != nil {
		t.Fatalf("failed to register leader RPC: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go server.ServeConn(conn)
		}
	}()

	followerStore := store.NewStore()
	followerSvc := NewService("node1", followerStore, []string{ln.Addr().String()})
	followerSvc.leaderID = "node3"

	req := &internalrpc.PutRequest{Key: "forwarded", Value: "value"}
	res := &internalrpc.PutResponse{}
	if err := followerSvc.Put(req, res); err != nil {
		t.Fatalf("unexpected put error: %v", err)
	}
	if !res.Success {
		t.Fatal("expected put to succeed via leader forwarding")
	}
	if _, ok := followerStore.Get("forwarded"); ok {
		t.Fatal("follower should not apply non-replication writes locally")
	}
	val, ok := leaderStore.Get("forwarded")
	if !ok || val != "value" {
		t.Fatal("leader did not apply forwarded write")
	}
}
