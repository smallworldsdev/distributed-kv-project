package cluster

import (
	"log"
	"net/rpc"

	internalrpc "github.com/smallworldsdev/distributed-kv-project/internal/rpc"
	"github.com/smallworldsdev/distributed-kv-project/internal/store"
)

type Service struct {
	NodeID string
	Store  *store.Store
	Peers  []string
}

func NewService(nodeID string, store *store.Store, peers []string) *Service {
	return &Service{
		NodeID: nodeID,
		Store:  store,
		Peers:  peers,
	}
}

func (s *Service) Ping(req *internalrpc.PingRequest, res *internalrpc.PingResponse) error {
	rpcRequests.WithLabelValues("Ping", "success").Inc()
	log.Printf("Received ping: %s", req.Message)
	res.Message = "Pong from node " + s.NodeID
	return nil
}

func (s *Service) Get(req *internalrpc.GetRequest, res *internalrpc.GetResponse) error {
	rpcRequests.WithLabelValues("Get", "received").Inc()
	log.Printf("Received Get request for key: %s", req.Key)
	val, found := s.Store.Get(req.Key)
	res.Value = val
	res.Found = found
	if found {
		kvOperations.WithLabelValues("Get", "found").Inc()
	} else {
		kvOperations.WithLabelValues("Get", "not_found").Inc()
	}
	return nil
}

func (s *Service) Put(req *internalrpc.PutRequest, res *internalrpc.PutResponse) error {
	rpcRequests.WithLabelValues("Put", "received").Inc()
	log.Printf("Received Put request for key: %s, value: %s, isReplication: %v", req.Key, req.Value, req.IsReplication)
	err := s.Store.Put(req.Key, req.Value)
	if err != nil {
		kvOperations.WithLabelValues("Put", "error").Inc()
		log.Printf("Error saving to store: %v", err)
		res.Success = false
		return err
	}
	kvOperations.WithLabelValues("Put", "success").Inc()

	// Replicate to peers if this is not a replication request
	if !req.IsReplication {
		for _, peer := range s.Peers {
			go s.replicateToPeer(peer, req)
		}
	}

	res.Success = true
	return nil
}

func (s *Service) replicateToPeer(peer string, originalReq *internalrpc.PutRequest) {
	client, err := rpc.Dial("tcp", peer)
	if err != nil {
		log.Printf("Failed to dial peer %s: %v", peer, err)
		return
	}
	defer client.Close()

	replReq := *originalReq
	replReq.IsReplication = true

	var replRes internalrpc.PutResponse
	if err := client.Call("NodeService.Put", &replReq, &replRes); err != nil {
		log.Printf("Failed to replicate to peer %s: %v", peer, err)
	}
}
