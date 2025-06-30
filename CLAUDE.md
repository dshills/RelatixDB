# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
RelatixDB is a high-performance, local graph database designed for use as a Model Context Protocol (MCP) tool server. It's optimized for fast, local, contextual knowledge manipulation by LLMs, not as an enterprise-grade database.

## Key Architecture Concepts

### Data Model
- **Nodes**: Uniquely identified by string IDs with optional types (e.g., `file`, `function`, `module`) and key/value properties
- **Edges**: Directed, labeled connections between nodes with optional metadata
- **Multigraph**: Multiple edges of different types allowed between same nodes

### Storage Engine
- Backed by BoltDB for persistent storage or in-memory for speed
- Read-optimized with append-only or journaling for writes
- Indexed by node ID, type, and edge relationships

### Interface Mode
**MCP Tool Server**: Standard Model Context Protocol implementation using JSON-RPC 2.0 over stdio for LLM interaction

## MCP Protocol Interface
RelatixDB implements the standard Model Context Protocol (MCP) using JSON-RPC 2.0. All communication follows the official MCP specification.

### Server Initialization
```json
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "client", "version": "1.0"}}}
```

### Tool Discovery
```json
{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}
```

### Available MCP Tools
1. **add_node** - Add nodes with ID, type, and properties
2. **add_edge** - Add directed, labeled edges between nodes  
3. **delete_node** - Delete nodes (and connected edges)
4. **delete_edge** - Delete specific edges
5. **query_neighbors** - Find neighboring nodes
6. **query_paths** - Find paths between nodes
7. **query_find** - Search nodes by type/properties

### Tool Execution Example
```json
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "node_id", "type": "type", "props": {"key": "value"}}}}
```

## Development Commands
This project includes a comprehensive Makefile for development tasks:

```bash
# Essential commands
make help              # Show all available targets
make build             # Build the RelatixDB binary
make test              # Run all tests
make test-mcp          # Run comprehensive MCP protocol tests
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
- Store builds in build/ directory
- Use second-opinion to check code before commit
- Use docs/ for documentation and planning documents
- Use rekatixdb to understand project structure

## Project Metadata Tracking
- Store the import relationships from this Go project in RelatixDB, with modules as nodes and "imports" as edge labels

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.