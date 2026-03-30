# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o rpc-balancer .

# Run (requires config.yaml in current dir)
./rpc-balancer
./rpc-balancer --config config.yaml --port 8080 --metrics-port 6060

# Dependencies
go mod tidy
go mod download

# Lint (uses golangci-lint v2)
golangci-lint run

# Docker build
docker build -t rpc-balancer .
```

There are no tests in this codebase currently.

## Architecture

All application logic lives in the `cmd/` package. `main.go` simply calls `cmd.Run()`.

**Request flow:**
1. `cmd/main.go` (`Run()`) — parses config, builds `network` map keyed by URL path, registers HTTP handlers, starts metrics server goroutine, then starts the main HTTP server.
2. Incoming requests are dispatched by `r.URL.Path` to the matching `network`. If the request has `Upgrade: websocket`, it is routed to a WebSocket upstream; otherwise to an HTTP upstream.
3. Upstream selection (`getNextUpstream` / `getNextWsUpstream`) uses random selection from the `HealthyUpstreams` / `WsUpstreams` slices.

**Health checking** (`cmd/upstreams.go` — `setHealthyUpstreams`):
- Runs in a background goroutine per network, polling every `upstreamCheckInterval` (15s).
- Calls `eth_blockNumber` then `eth_getBlockByNumber` on each upstream concurrently.
- An upstream is healthy if its block is within `blockHealthyDiff` (5 blocks) or its timestamp is within `timestampHealthyDiff` (3 seconds) of the maximum observed across all upstreams.
- Unhealthy upstreams are removed from `HealthyUpstreams`/`WsUpstreams` slices; healthy ones are added back.

**Key types:**
- `Configuration` (`cmd/config.go`) — YAML config struct; global `config` var populated at startup.
- `network` (`cmd/network.go`) — pairs a chain's metadata with its `*upstreams`.
- `upstreams` (`cmd/upstreams.go`) — holds all upstreams + healthy subset; performs health checks.
- `upstream` — wraps an `httputil.ReverseProxy`, a `*WebsocketProxy`, and the `rpcEndpoint`.
- `rpcEndpoint` (`cmd/rpc.go`) — holds HTTP and WS URLs (both parsed and raw) and the health-check RPC call helpers.
- `WebsocketProxy` (`cmd/websockets.go`) — full-duplex WebSocket reverse proxy using `gorilla/websocket`.

**Metrics** (`cmd/prometheus.go`): All Prometheus gauges/counters are package-level vars registered at init via `promauto`. Labels use `chainid`, `chainname`, `name`, `url`. The metrics server runs on a separate port (default 6060).

**Config file** (`config.yaml`):
```yaml
networks:
  - chainId: "100500"
    name: MyNet
    path: /my-net          # URL path the balancer listens on
    upstreams:
      - name: node-1
        url: "http://host:8545"
        wsUrl: "ws://host:8546"
```

**Hardcoded tunables** in `cmd/config.go`:
- `connectTimeout = 5` (seconds)
- `upstreamCheckInterval = 15` (seconds)
- `blockHealthyDiff = 5` (blocks)
- `timestampHealthyDiff = 3` (seconds)

**CI/CD:**
- Lint: `golangci-lint` runs on push/PR to `main`.
- Build & release: GoReleaser publishes binaries on `v*` tags; Docker image (`ghcr.io/ai-ia-org/rpc-balancer`) is built for `linux/amd64` and `linux/arm64` on pushes to `main` and on `v*` tags.
