package cluster

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	rpcRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kv_rpc_requests_total",
			Help: "Total number of RPC requests",
		},
		[]string{"method", "status"},
	)

	kvOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kv_operations_total",
			Help: "Total number of KV store operations",
		},
		[]string{"operation", "result"},
	)
)
