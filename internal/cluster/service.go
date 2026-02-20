package cluster

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"

	internalrpc "github.com/smallworldsdev/distributed-kv-project/internal/rpc"
	"github.com/smallworldsdev/distributed-kv-project/internal/store"
)

const (
	HeartbeatInterval = 1 * time.Second
	ElectionTimeout   = 3 * time.Second
)

type Service struct {
	NodeID string
	Store  *store.Store
	Peers  []string
	
	mu             sync.Mutex
	cond           *sync.Cond
	lamportClock   int64
	myTimestamp    int64
	requestingCS   bool
	replyCount     int
	deferredReplies []string
	
	leaderID       string
	isElectionInProgress bool
	lastHeartbeat  time.Time
	done           chan struct{}

	snapshotID     int64
	snapshots      map[int64]map[string]string
}

func NewService(nodeID string, store *store.Store, peers []string) *Service {
	s := &Service{
		NodeID:    nodeID,
		Store:     store,
		Peers:     peers,
		deferredReplies: make([]string, 0),
		snapshots: make(map[int64]map[string]string),
		lastHeartbeat: time.Now(),
		done:          make(chan struct{}),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *Service) Start() {
	go s.runHeartbeatSender()
	go s.runElectionMonitor()
}

func (s *Service) Shutdown() {
	close(s.done)
}

func (s *Service) Ping(req *internalrpc.PingRequest, res *internalrpc.PingResponse) error {
	log.Printf("Received ping: %s", req.Message)
	res.Message = "Pong from node " + s.NodeID
	return nil
}

func (s *Service) Get(req *internalrpc.GetRequest, res *internalrpc.GetResponse) error {
	log.Printf("Received Get request for key: %s", req.Key)
	val, found := s.Store.Get(req.Key)
	res.Value = val
	res.Found = found
	return nil
}

func (s *Service) Put(req *internalrpc.PutRequest, res *internalrpc.PutResponse) error {
	log.Printf("Received Put request for key: %s, value: %s, isReplication: %v", req.Key, req.Value, req.IsReplication)
	err := s.Store.Put(req.Key, req.Value)
	if err != nil {
		log.Printf("Error saving to store: %v", err)
		res.Success = false
		return err
	}

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

// ---- Mutual Exclusion (Ricart-Agrawala) ----

func (s *Service) EnterCS() {
	s.mu.Lock()
	s.requestingCS = true
	s.lamportClock++
	s.myTimestamp = s.lamportClock
	timestamp := s.myTimestamp
	s.mu.Unlock()

	log.Printf("Node %s requesting CS with timestamp %d", s.NodeID, timestamp)

	// Send RequestMutex to all peers
	// In a real implementation, we should handle errors and retries.
	// Here we just wait for all to return.
	var wg sync.WaitGroup
	for _, peer := range s.Peers {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			client, err := rpc.Dial("tcp", p)
			if err != nil {
				log.Printf("Failed to dial peer %s for mutex: %v", p, err)
				return // Proceed even if peer is down? In RA, we need all permissions. Assuming reliability here.
			}
			defer client.Close()

			req := internalrpc.RequestMutexRequest{
				NodeID:    s.NodeID,
				Timestamp: timestamp,
			}
			var res internalrpc.RequestMutexResponse
			if err := client.Call("NodeService.RequestMutex", &req, &res); err != nil {
				log.Printf("Failed to request mutex from %s: %v", p, err)
			}
		}(peer)
	}
	wg.Wait()
	log.Printf("Node %s entered CS", s.NodeID)
}

func (s *Service) ExitCS() {
	s.mu.Lock()
	s.requestingCS = false
	s.cond.Broadcast() // Wake up waiting requests
	s.mu.Unlock()
	log.Printf("Node %s exited CS", s.NodeID)
}

func (s *Service) RequestMutex(req *internalrpc.RequestMutexRequest, res *internalrpc.RequestMutexResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update clock
	if req.Timestamp > s.lamportClock {
		s.lamportClock = req.Timestamp
	}
	s.lamportClock++

	// Decide whether to defer
	// Defer if we are requesting AND (our timestamp < req timestamp OR (timestamps equal AND our ID < req ID))
	// Note: We use string comparison for NodeID as tie-breaker.
	shouldDefer := s.requestingCS && 
		(s.myTimestamp < req.Timestamp || (s.myTimestamp == req.Timestamp && s.NodeID < req.NodeID))

	if shouldDefer {
		log.Printf("Node %s deferring request from %s (myTS: %d, reqTS: %d)", s.NodeID, req.NodeID, s.myTimestamp, req.Timestamp)
		// Wait until we are done with CS
		// We loop because cond.Wait() can spuriously wake up, or we might be woken up but then someone else gets in?
		// Actually in RA, once we exit CS, we reply to ALL deferred.
		// So when we wake up, 'requestingCS' should be false.
		for s.requestingCS && (s.myTimestamp < req.Timestamp || (s.myTimestamp == req.Timestamp && s.NodeID < req.NodeID)) {
			s.cond.Wait()
		}
	}

	res.Granted = true
	return nil
}

// ---- Leader Election (Bully Algorithm) ----

func (s *Service) StartElection() {
	s.mu.Lock()
	if s.isElectionInProgress {
		s.mu.Unlock()
		return
	}
	s.isElectionInProgress = true
	s.mu.Unlock()

	log.Printf("Node %s starting election", s.NodeID)

	higherNodeFound := false
	var wg sync.WaitGroup

	// Assume Peers contains *other* nodes. If self is in Peers, we need to skip.
	// But since we don't know our own address, we just ignore self-dial issues or handle errors.
	for _, peer := range s.Peers {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			client, err := rpc.Dial("tcp", p)
			if err != nil {
				// Peer down, ignore
				return
			}
			defer client.Close()

			req := internalrpc.ElectionRequest{NodeID: s.NodeID}
			var res internalrpc.ElectionResponse
			if err := client.Call("NodeService.Election", &req, &res); err == nil {
				if res.Alive {
					s.mu.Lock()
					higherNodeFound = true
					s.mu.Unlock()
				}
			}
		}(p)
	}
	wg.Wait()

	s.mu.Lock()
	if higherNodeFound {
		log.Printf("Node %s found higher node, yielding", s.NodeID)
		s.isElectionInProgress = false
		s.mu.Unlock()
		// Wait for coordinator message... (handled by Coordinator RPC)
	} else {
		log.Printf("Node %s is the new leader", s.NodeID)
		s.leaderID = s.NodeID
		s.isElectionInProgress = false
		
		// Broadcast Coordinator
		for _, peer := range s.Peers {
			go func(p string) {
				client, err := rpc.Dial("tcp", p)
				if err != nil {
					return
				}
				defer client.Close()
				
				req := internalrpc.CoordinatorRequest{NodeID: s.NodeID}
				var res internalrpc.CoordinatorResponse
				if err := client.Call("NodeService.Coordinator", &req, &res); err != nil {
					log.Printf("Failed to notify coordinator to %s: %v", p, err)
				}
			}(peer)
		}
		s.mu.Unlock()
	}
}

func (s *Service) Election(req *internalrpc.ElectionRequest, res *internalrpc.ElectionResponse) error {
	s.mu.Lock()
	myID := s.NodeID
	s.mu.Unlock()

	// If my ID is higher, I stop the sender by replying Alive=true
	// And I start my own election if not already started.
	if myID > req.NodeID {
		res.Alive = true
		go s.StartElection()
	} else {
		// If my ID is lower, I yield. Alive=false.
		res.Alive = false
	}
	return nil
}

func (s *Service) Coordinator(req *internalrpc.CoordinatorRequest, res *internalrpc.CoordinatorResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leaderID = req.NodeID
	s.isElectionInProgress = false
	s.lastHeartbeat = time.Now() // Reset timeout on new leader
	log.Printf("Node %s acknowledged leader: %s", s.NodeID, req.NodeID)
	res.Ack = true
	return nil
}

// ---- Snapshotting (Chandy-Lamport Simplified) ----

func (s *Service) StartSnapshot() {
	snapshotID := time.Now().UnixNano()
	log.Printf("Node %s starting snapshot %d", s.NodeID, snapshotID)

	// Record local state
	state := s.Store.GetCopy()
	
	s.mu.Lock()
	if s.snapshots == nil {
		s.snapshots = make(map[int64]map[string]string)
	}
	s.snapshots[snapshotID] = state
	s.mu.Unlock()

	// Send Marker to all peers
	for _, peer := range s.Peers {
		go func(p string) {
			client, err := rpc.Dial("tcp", p)
			if err != nil {
				return
			}
			defer client.Close()

			req := internalrpc.SnapshotRequest{
				InitiatorID: s.NodeID,
				SnapshotID:  snapshotID,
			}
			var res internalrpc.SnapshotResponse
			if err := client.Call("NodeService.Snapshot", &req, &res); err != nil {
				log.Printf("Failed to send snapshot marker to %s: %v", p, err)
			}
		}(peer)
	}
}

func (s *Service) Snapshot(req *internalrpc.SnapshotRequest, res *internalrpc.SnapshotResponse) error {
	s.mu.Lock()
	if s.snapshots == nil {
		s.snapshots = make(map[int64]map[string]string)
	}
	// Check if we already recorded state for this snapshot
	if _, exists := s.snapshots[req.SnapshotID]; exists {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	log.Printf("Node %s recording state for snapshot %d initiated by %s", s.NodeID, req.SnapshotID, req.InitiatorID)

	// Record state
	state := s.Store.GetCopy()
	
	s.mu.Lock()
	// Double check
	if _, exists := s.snapshots[req.SnapshotID]; !exists {
		s.snapshots[req.SnapshotID] = state
	}
	s.mu.Unlock()

	res.Success = true
	return nil
}

// ---- Client Trigger ----

func (s *Service) Trigger(req *internalrpc.TriggerRequest, res *internalrpc.TriggerResponse) error {
	log.Printf("Received trigger command: %s", req.Command)
	switch req.Command {
	case "mutex":
		go func() {
			log.Println("Initiating Mutex Entry...")
			s.EnterCS()
			log.Println("Entered CS. Doing critical work (simulated)...")
			time.Sleep(2 * time.Second)
			s.ExitCS()
			log.Println("Exited CS.")
		}()
		res.Message = "Mutex algorithm triggered"
		res.Success = true
	case "election":
		go s.StartElection()
		res.Message = "Leader election triggered"
		res.Success = true
	case "snapshot":
		go s.StartSnapshot()
		res.Message = "Snapshot algorithm triggered"
		res.Success = true
	default:
		res.Message = "Unknown command"
		res.Success = false
	}
	return nil
}

// ---- Heartbeat & Monitoring ----

func (s *Service) Heartbeat(req *internalrpc.HeartbeatRequest, res *internalrpc.HeartbeatResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If we receive a heartbeat, update our lastHeartbeat timestamp
	s.lastHeartbeat = time.Now()

	// If the heartbeat comes from a valid leader, ensure we recognize them
	if req.LeaderID != s.leaderID {
		log.Printf("Node %s received heartbeat from new leader: %s", s.NodeID, req.LeaderID)
		s.leaderID = req.LeaderID
		s.isElectionInProgress = false
	}

	res.Success = true
	return nil
}

func (s *Service) runHeartbeatSender() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			isLeader := s.leaderID == s.NodeID
			s.mu.Unlock()

			if isLeader {
				for _, peer := range s.Peers {
					go func(p string) {
						client, err := rpc.Dial("tcp", p)
						if err != nil {
							// Peer might be down
							return
						}
						defer client.Close()

						req := internalrpc.HeartbeatRequest{LeaderID: s.NodeID}
						var res internalrpc.HeartbeatResponse
						if err := client.Call("NodeService.Heartbeat", &req, &res); err != nil {
							// Log failure sparingly?
						}
					}(peer)
				}
			}
		}
	}
}

func (s *Service) runElectionMonitor() {
	ticker := time.NewTicker(500 * time.Millisecond) // Check frequently
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			elapsed := time.Since(s.lastHeartbeat)
			isLeader := s.leaderID == s.NodeID
			electionInProgress := s.isElectionInProgress
			s.mu.Unlock()

			if !isLeader && !electionInProgress && elapsed > ElectionTimeout {
				log.Printf("Node %s detected leader failure (last heartbeat %v ago). Starting election.", s.NodeID, elapsed)
				// Reset lastHeartbeat to prevent immediate re-triggering while election starts?
				// StartElection will set isElectionInProgress=true, which blocks this check.
				go s.StartElection()
			}
		}
	}
}
