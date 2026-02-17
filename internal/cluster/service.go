package cluster

import (
	"log"

	"github.com/smallworldsdev/distributed-kv-project/internal/rpc"
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

func (s *Service) Ping(req *rpc.PingRequest, res *rpc.PingResponse) error {
	log.Printf("Received ping: %s", req.Message)
	res.Message = "Pong from node " + s.NodeID
	return nil
}

func (s *Service) Get(req *rpc.GetRequest, res *rpc.GetResponse) error {
	log.Printf("Received Get request for key: %s", req.Key)
	val, found := s.Store.Get(req.Key)
	res.Value = val
	res.Found = found
	return nil
}

func (s *Service) Put(req *rpc.PutRequest, res *rpc.PutResponse) error {
	log.Printf("Received Put request for key: %s, value: %s", req.Key, req.Value)
	err := s.Store.Put(req.Key, req.Value)
	if err != nil {
		log.Printf("Error saving to store: %v", err)
		res.Success = false
		return err
	}
	res.Success = true
	return nil
}
