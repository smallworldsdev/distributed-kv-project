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
	Key   string
	Value string
}

type PutResponse struct {
	Success bool
}
