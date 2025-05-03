package cmd

import (
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	rpcBalancerUpstreamLatestBlock = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rpc_balancer_upstream_latest_block",
		Help: "Latest block available on upstream RPC node",
	},
	[]string{"chainid","chainname","name","url"},
	)
)
var (
	rpcBalancerUpstreamLatestBlockTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rpc_balancer_upstream_latest_block_timestamp",
		Help: "Latest block timestamp on upstream RPC node",
	},
	[]string{"chainid","chainname","name","url"},
	)
)
var (
	rpcBalancerUpstreamUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rpc_balancer_upstream_up",
		Help: "RPC health",
	},
	[]string{"chainid","chainname","name","url"},
	)
)
var (
	rpcBalancerChainHealthyUpstreamNum = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rpc_balancer_chain_healthy_upstream_num",
		Help: "Number of healthy upstreams for chain",
	},
	[]string{"chainid","chainname"},
	)
)
var (
	rpcBalancerChainLatestBlock = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rpc_balancer_chain_latest_block",
		Help: "Latest block available for chain",
	},
	[]string{"chainid","chainname"},
	)
)
var (
	rpcBalancerChainLatestBlockTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rpc_balancer_chain_latest_block_timestamp",
		Help: "Latest block timestamp for chain",
	},
	[]string{"chainid","chainname"},
	)
)
var (
	rpcBalancerUpstreamHttpRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_balancer_upstream_http_requests_total",
		Help: "Total HTTP requests to upstream",
	},
	[]string{"chainid","chainname","name","url"},
	)
)
var (
	rpcBalancerUpstreamHttpRequest = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_balancer_upstream_http_requests",
		Help: "HTTP requests to upstream by status",
	},
	[]string{"url","status"},
	)
)
var (
	rpcBalancerUpstreamWsRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rpc_balancer_upstream_ws_requests_total",
		Help: "Total WS requests to upstream",
	},
	[]string{"chainid","chainname","name","url"},
	)
)