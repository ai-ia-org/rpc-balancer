# RPC Balancer

A Golang load balancer for Ethereum JSON RPC servers.

[![Go Report Card](https://goreportcard.com/badge/github.com/ai-ia-org/rpc-balancer)](https://goreportcard.com/report/github.com/ai-ia-org/rpc-balancer)

## Features

- Load balancing across multiple RPC endpoints
- Health monitoring and automatic failover
- Metrics collection and exportation (Prometheus compatible) (TBD)
- Support for HTTP and WS protocols

## Installation

### From Source

```bash
git clone https://github.com/ai-ia-org/rpc-balancer.git
cd rpc-balancer
go build -o rpc-balancer .
```

### Pre-built Binaries (Will be available soon)

Download the pre-built binaries for your platform from the [releases page](https://github.com/ai-ia-org/rpc-balancer/releases).

### Docker Installation

Run docker

```bash
docker run --name rpc-balancer -d -p 8080:8080 ghcr.io/ai-ia-org/rpc-balancer:latest
```

## Quick Start

### Using Binary

1. Create a configuration file:

```bash
cat > config.yml << EOF
networks:
  - chainId: 100500
    name: Devnet1
    path: /devnet-1
    upstreams:
    - name: devnet-1
      url: "http:/localhost:8545"
      wsUrl: "ws://localhost:8546"
  - chainId: 100501
    name: Devnet2
    path: /devnet-2
    upstreams:
    - name: devnet-2
      url: "http://localhost:18545"
      wsUrl: "ws://localhost:18546"
EOF
```

2. Run the balancer:

```bash
./rpc-balancer
```

### Using Docker

1. Create a configuration file as shown above.

2. Run with Docker:

```bash
docker run -d \
  --name rpc-balancer \
  -p 8080:8080 \
  -v $(pwd)/config.yml:/app/config.yml \
  ghcr.io/ai-ia-org/rpc-balancer:latest
```

## Command Line Options

```
Usage: rpc-balancer [options]

Options:
  --config string      Path to config file (default "config.yml")
  --port int           Server port (default 8080)
  --metrics-port int   Metrics port (default 6060)
```

## Health Checks

RPC Balancer performs regular health checks on backend services. If a backend fails its health check, it is temporarily removed from the load balancing pool until it becomes healthy again.

## Monitoring

### Prometheus Metrics

RPC Balancer exposes the following Prometheus metrics on the metrics port:

- `rpc_balancer_upstream_latest_block` - Latest block available on upstream
- `rpc_balancer_upstream_latest_block_timestamp` - Timestamp of latest block available on upstream
- `rpc_balancer_upstream_up` - Upstream health status (1 = up, 0 = down)
- `rpc_balancer_chain_latest_block` - Latest block available for whole chain (max block from all upstreams)
- `rpc_balancer_chain_latest_block_timestamp` - Timestamp of latest block available for whole chain
- `rpc_balancer_chain_healthy_upstream_num` - Number of healhy upstreams for chain

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the GPL-3.0 License - see the LICENSE file for details.

## Acknowledgments

- Thanks to all contributors who have helped with the development of RPC Balancer
