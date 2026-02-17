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
