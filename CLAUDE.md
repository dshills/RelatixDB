# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
RelatixDB is a high-performance, local graph database designed for use as a Model Context Protocol (MCP) tool. It's optimized for fast, local, contextual knowledge manipulation by LLMs, not as an enterprise-grade database.

## Key Architecture Concepts

### Data Model
- **Nodes**: Uniquely identified by string IDs with optional types (e.g., `file`, `function`, `module`) and key/value properties
- **Edges**: Directed, labeled connections between nodes with optional metadata
- **Multigraph**: Multiple edges of different types allowed between same nodes

### Storage Engine
- Backed by LMDB, BoltDB, or custom mmap-based store
- Read-optimized with append-only or journaling for writes
- Indexed by node ID, type, and edge relationships

### Interface Modes
1. **MCP Tool Server**: JSON-based stdio interface for LLM interaction
2. **Go API**: Optional embedded Go package interface

## Command Interface
The database operates via JSON commands on stdin with JSON responses on stdout:

```json
{"cmd": "add_node", "args": {"id": "node_id", "type": "type", "props": {}}}
{"cmd": "add_edge", "args": {"from": "id1", "to": "id2", "label": "relationship"}}
{"cmd": "query", "args": {"type": "neighbors", "node": "id", "direction": "out"}}
```

## Development Commands
This project includes a comprehensive Makefile for development tasks:

```bash
# Essential commands
make help              # Show all available targets
make build             # Build the RelatixDB binary
make test              # Run all tests
make test-mcp          # Run comprehensive MCP tool function tests
make validate          # Run all validation checks (lint, vet, test)

# Testing commands
make test-unit         # Run unit tests only
make test-integration  # Run integration tests
make test-bench        # Run benchmark tests
make test-coverage     # Run tests with coverage report
make test-race         # Run tests with race detector

# Development workflow
make fmt              # Format Go code
make lint             # Run golangci-lint
make vet              # Run go vet
make clean            # Clean build artifacts
make pre-commit       # Run checks before committing

# Running RelatixDB
make run              # Run in memory mode with debug
make run-persistent   # Run with persistent storage
make dev              # Run in development mode

# Direct Go commands (if needed)
go mod tidy           # Tidy modules
go test ./...         # Run all tests
go test -race ./...   # Run with race detection
golangci-lint run     # Lint code
```

## Performance Targets
- Node insertion: < 100µs
- Edge insertion: < 150µs  
- Neighborhood query (1-hop): < 1ms
- Path query (depth ≤ 4): < 10ms

## Primary Use Cases
- Enable LLM reasoning about entities and relationships
- Store source code element relationships
- Track tool actions and derivations
- Support lightweight AI agent planning and memory

## Tool Memories
- Use embeddix tools to save and pull relevant project information
- store builds in build/ directory
- Use second-opinion to check code before commit
- Use docs/ for documentation and planning documents