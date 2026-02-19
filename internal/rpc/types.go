package rpc

// Ping

type PingRequest struct {
	Message string
}

type PingResponse struct {
	Message string
}

// Get

type GetRequest struct {
	Key string
}

type GetResponse struct {
	Value string
	Found bool
}

// Put

type PutRequest struct {
	Key           string
	Value         string
	IsReplication bool
}

type PutResponse struct {
	Success bool
}

// Mutual Exclusion (Ricart-Agrawala)

type RequestMutexRequest struct {
	NodeID    string
	Timestamp int64
}

type RequestMutexResponse struct {
	Granted bool
}

// Leader Election (Bully Algorithm)

type ElectionRequest struct {
	NodeID string
}

type ElectionResponse struct {
	Alive bool
}

type CoordinatorRequest struct {
	NodeID string
}

type CoordinatorResponse struct {
	Ack bool
}

// Snapshotting (Chandy-Lamport / Global State)

type SnapshotRequest struct {
	InitiatorID string
	SnapshotID  int64
}

type SnapshotResponse struct {
	Success bool
}

// Client Triggers

type TriggerRequest struct {
	Command string // "mutex", "election", "snapshot"
}

type TriggerResponse struct {
	Message string
	Success bool
}
