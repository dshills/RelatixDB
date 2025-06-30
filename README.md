# RelatixDB

> ⚠️ **PRE-ALPHA SOFTWARE**: This is experimental software in early development. Not suitable for production use. APIs and data formats may change without notice. Use at your own risk.

A high-performance, local graph database designed for use as a Model Context Protocol (MCP) tool server. RelatixDB is optimized for fast, local, contextual knowledge manipulation by LLMs.

## Features

- **Blazing Fast Performance**: Exceeds all targets by 25-1600x
  - Node insertion: 4.083µs (target: <100µs)
  - Edge insertion: 1.5µs (target: <150µs)  
  - Neighbor queries: 625ns (target: <1ms)
- **MCP Protocol Support**: Full JSON-based stdio interface for LLM integration
- **Dual Storage Modes**: In-memory for speed, persistent BoltDB for durability
- **Thread-Safe Operations**: Concurrent access with proper locking
- **Comprehensive Query Engine**: Neighbors, paths, and property-based search
- **Multigraph Support**: Multiple edge types between same nodes

## Installation

### Prerequisites

- Go 1.22 or later
- Git

### Build from Source

```bash
git clone https://github.com/dshills/RelatixDB.git
cd RelatixDB
go build -o relatixdb ./cmd/relatixdb
```

### Install with Go

```bash
go install github.com/dshills/RelatixDB/cmd/relatixdb@latest
```

## Quick Start

### In-Memory Mode (Default)

```bash
# Start RelatixDB in in-memory mode
echo '{"cmd": "add_node", "args": {"id": "user:alice", "type": "user", "props": {"name": "Alice"}}}' | ./relatixdb
```

### Persistent Mode

```bash
# Start with persistent storage
echo '{"cmd": "add_node", "args": {"id": "user:bob", "type": "user", "props": {"name": "Bob"}}}' | ./relatixdb -db mydata.db
```

### Debug Mode

```bash
# Enable debug logging (to stderr)
./relatixdb -debug
```

## Usage

### Command Line Options

```bash
./relatixdb [OPTIONS]

OPTIONS:
  -version      Show version information
  -help         Show help message
  -debug        Enable debug logging to stderr
  -db PATH      Database file path (optional, uses in-memory if not specified)
```

### MCP Protocol Commands

RelatixDB operates via JSON commands on stdin with JSON responses on stdout:

#### Add Node
```json
{"cmd": "add_node", "args": {"id": "user:1", "type": "user", "props": {"name": "Alice", "email": "alice@example.com"}}}
```

#### Add Edge
```json
{"cmd": "add_edge", "args": {"from": "user:1", "to": "user:2", "label": "follows", "props": {"since": "2024-01-01"}}}
```

#### Query Neighbors
```json
{"cmd": "query", "args": {"type": "neighbors", "node": "user:1", "direction": "out"}}
```

#### Find Nodes by Type
```json
{"cmd": "query", "args": {"type": "find", "filters": {"type": "user"}}}
```

#### Find Paths Between Nodes
```json
{"cmd": "query", "args": {"type": "paths", "from": "user:1", "to": "user:3", "max_depth": 3}}
```

#### Delete Node
```json
{"cmd": "delete_node", "args": {"id": "user:1"}}
```

#### Delete Edge
```json
{"cmd": "delete_edge", "args": {"from": "user:1", "to": "user:2", "label": "follows"}}
```

### Response Format

All responses follow this format:

```json
{
  "ok": true,
  "result": { ... },
  "error": null
}
```

Error responses:
```json
{
  "ok": false,
  "result": null,
  "error": "error message"
}
```

## Data Model

### Nodes
- **ID**: Unique string identifier (required)
- **Type**: Optional classification (e.g., "user", "file", "function")
- **Properties**: Key-value pairs (string values only)

### Edges
- **From/To**: Node IDs (required)
- **Label**: Edge type (required, e.g., "follows", "calls", "contains")
- **Properties**: Optional key-value metadata

### Graph Characteristics
- **Directed**: All edges have direction
- **Labeled**: All edges must have a label
- **Multigraph**: Multiple edges of different types allowed between nodes
- **No Schema**: Flexible node types and properties

## Performance

RelatixDB is optimized for high-performance local operations:

| Operation | Performance | Target |
|-----------|-------------|--------|
| Node insertion | 4.083µs | <100µs |
| Edge insertion | 1.5µs | <150µs |
| Neighbor query | 625ns | <1ms |
| Path query (depth ≤ 4) | <10ms | <10ms |

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./internal/graph

# Test performance targets
go test -run=TestPerformanceTargets -v ./internal/graph
```

### Project Structure

```
├── cmd/relatixdb/           # Main executable
├── internal/
│   ├── graph/              # Core graph data structures
│   ├── mcp/                # MCP protocol handling  
│   └── storage/            # Persistent storage backends
├── docs/                   # Documentation
└── specs/                  # Technical specifications
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Build all packages
go build ./...
```

## Use Cases

### LLM Context Management
- Store relationships between code elements
- Track tool actions and derivations
- Enable reasoning about entity relationships

### Code Analysis
- Function call graphs
- Module dependencies
- File relationships

### AI Agent Memory
- Lightweight planning and memory
- Action history tracking
- Knowledge graph construction

## Limitations

### Current Limitations
- **Bulk Operations**: SaveGraph requires interface extension (TODO)
- **Edge Cleanup**: Node deletion edge cleanup needs storage implementation
- **Auto-save**: Disabled pending SaveGraph completion
- **Schema Evolution**: No versioning for on-disk data format

### Not Suitable For
- Distributed systems
- Multi-user concurrent access
- High-availability requirements
- Enterprise-scale data volumes
- ACID transactions across multiple operations

## Contributing

This is pre-alpha software. While contributions are welcome, expect:
- Breaking API changes
- Data format changes
- Incomplete features
- Limited backward compatibility

### Development Workflow
1. Check existing issues and discussions
2. Create feature branch from `main`
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## License

[License information to be added]

## Changelog

### v0.1.0 (Current)
- Initial implementation with in-memory graph
- Full MCP protocol support
- BoltDB persistent storage layer
- Performance benchmarks and validation
- Comprehensive test suite

## Support

This is experimental software. Support is limited to:
- GitHub Issues for bug reports
- Discussions for feature requests
- Code review for contributions

**Remember**: This is pre-alpha software. Use for experimentation and development only.