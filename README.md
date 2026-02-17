
# Distributed KV

A distributed key–value store written in Go using the standard `net/rpc` package.

This project is designed to:

* Demonstrate core distributed systems concepts
* Run across multiple laptops on a LAN
* Run locally using Docker containers
* Be easy to develop and extend collaboratively

---

# Features (Planned / In Progress)

* RPC-based inter-node communication (`net/rpc`)
* Multi-node cluster support
* Docker-based local multi-node simulation
* Makefile-based developer workflow

---

# Project Structure

```
distributed-kv-project/
│
├── cmd/
│   ├── node/           # Node binary entrypoint
│   └── client/         # CLI client
│
├── internal/
│   ├── rpc/            # RPC server + client wrappers
│   ├── store/          # Storage engine
│   ├── cluster/        # Node + replication logic
│   └── config/         # Configuration parsing
│
├── deployments/
│   └── docker/         # Dockerfile + docker-compose
│
├── test/
│   └── integration/    # Distributed integration tests
│
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

# Architecture Overview

## Node Responsibilities

Each node:

* Hosts a key-value store
* Exposes RPC endpoints
* Communicates with peer nodes

## Logical Separation

We explicitly separate:

* RPC transport layer (`internal/rpc`)
* Storage engine (`internal/store`)
* Cluster coordination (`internal/cluster`)
* Configuration parsing (`internal/config`)

This keeps business logic independent of network transport.

---

# Getting Started

## 1. Clone Repository

```bash
git clone git@github.com/smallworldsdev/distributed-kv-project.git
cd distributed-kv
```

---

## 2. Build

```bash
make build
```

Binaries are placed in `bin/`.

---

## 3. Run Single Node (Dev Mode)

```bash
make run
```

Or manually:

```bash
PORT=8001 go run ./cmd/node
```

---

## 4. Run Client

```bash
make client
```

---

# Docker Setup (Local Multi-Node Testing)

This simulates multiple machines locally.

## Start 3-Node Cluster

```bash
make up
```

This uses:

```
deployments/docker/docker-compose.yml
```
## View Logs
```
make logs
```

## Stop Cluster

```bash
make down
```

---

# Running Across Multiple Laptops

To demo across real machines:

1. Ensure all laptops are on the same network.
2. Determine each machine’s LAN IP.
3. Example command to start node1 with sample ip addresses:

```bash
PORT=8001 PEERS=192.168.1.10:8002,192.168.1.11:8003 go run ./cmd/node
```

Important:

* Do NOT use `localhost`
* Ensure firewall allows chosen ports

---

# Makefile Commands

| Command       | Description                  |
| ------------- | ---------------------------- |
| `make build`  | Build node + client binaries |
| `make run`    | Run single node locally      |
| `make client` | Run CLI client               |
| `make test`   | Run all tests                |
| `make up`     | Start docker cluster         |
| `make down`   | Stop docker cluster          |
| `make logs`   | View docker logs             |
| `make clean`  | Remove build artifacts       |

---

# Development Guide

## Code Organization Rules

1. `cmd/` only contains main entrypoints.
2. All core logic goes inside `internal/`.
3. RPC handlers should call service logic — not contain business logic.
4. Storage layer must not depend on RPC layer.
5. Cluster logic must not depend on CLI client.

---

## Writing Tests

### Unit Tests

```
internal/store/store_test.go
```

Run:

```bash
make test
```

### Integration Tests

Use:

* Multiple in-memory nodes
* Or docker-based environment

Goal: simulate real distributed behavior.

---

# Configuration

Nodes use environment variables:

| Variable  | Description               |
| --------- | ------------------------- |
| `PORT`    | Listening port            |
| `PEERS`   | Comma-separated peer list |
| `NODE_ID` | Optional logical node ID  |

Example(for local):

```bash
PORT=8001 PEERS=localhost:8002,localhost:8003 go run ./cmd/node
```

---

# RPC Design

We use Go’s standard:

```go
net/rpc
```

RPC methods follow:

```
func (s *Service) Method(args *Args, reply *Reply) error
```

---

# Development Workflow (Team)

## Branch Strategy

* `main` → stable demo branch
* `dev` → integration branch
* `feature/<name>` → feature work

Example:

```
feature/replication
feature/leader-election
feature/persistence
```

---

## Commit Guidelines

* Small, atomic commits
* Meaningful messages
* No broken `main`

---

# Demo Strategy

For presentations:

### Local Demo

```bash
make up
```

Then:

* Kill one container
* Show cluster behavior
* Restart container

### Multi-Laptop Demo

* 3 laptops
* Same WiFi
* Start nodes manually
* Demonstrate replication and failure

---

# Contributing

1. Create feature branch
2. Submit PR

All PRs must:

* Pass tests
* Not break Docker setup
* Preserve architecture separation

