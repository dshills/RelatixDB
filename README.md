# RelatixDB

> ⚠️ **PRE-ALPHA SOFTWARE**: This is experimental software in early development. Not suitable for production use. APIs and data formats may change without notice. Use at your own risk.

A high-performance, local graph database designed for use as a Model Context Protocol (MCP) tool server. RelatixDB is optimized for fast, local, contextual knowledge manipulation by LLMs.

## Features

- **Blazing Fast Performance**: Exceeds all targets by 90-960x
  - Node insertion: 1.084µs (target: <100µs) - 92x faster
  - Edge insertion: 1.708µs (target: <150µs) - 88x faster  
  - Neighbor queries: 1.042µs (target: <1ms) - 959x faster
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

## Claude Code Integration

RelatixDB can be integrated with Claude Code as an MCP (Model Context Protocol) server to provide graph database capabilities directly within your AI-assisted development workflow.

### Prerequisites

- [Claude Code](https://claude.ai/code) installed and configured
- RelatixDB binary built or installed (see Installation section above)

### Configuration

#### 1. Build RelatixDB

First, ensure RelatixDB is built and accessible:

```bash
git clone https://github.com/dshills/RelatixDB.git
cd RelatixDB
make build
# Binary will be at ./build/relatixdb
```

#### 2. Configure MCP Server

Add RelatixDB to your Claude Code MCP configuration. The exact method depends on your Claude Code setup:

**Option A: Direct Configuration**
```json
{
  "mcp": {
    "servers": {
      "relatixdb": {
        "command": "/path/to/RelatixDB/build/relatixdb",
        "args": ["-db", "/path/to/your/graph.db", "-debug"],
        "env": {}
      }
    }
  }
}
```

**Option B: In-Memory Mode (Fast, Non-Persistent)**
```json
{
  "mcp": {
    "servers": {
      "relatixdb": {
        "command": "/path/to/RelatixDB/build/relatixdb",
        "args": ["-debug"],
        "env": {}
      }
    }
  }
}
```

#### 3. Verify Integration

Start Claude Code and verify the MCP server is running:

```bash
# In Claude Code, you should be able to use commands like:
# "Add a node to the graph database"
# "Query relationships in the database"
# "Show me all nodes of type 'function'"
```

### Using RelatixDB in Claude Code

Once integrated, you can use natural language to interact with your graph database:

#### Adding Data
```
"Add a node with ID 'main.go' of type 'file' with property 'language' set to 'go'"

"Create an edge from 'main.go' to 'config.go' with label 'imports'"
```

#### Querying Data
```
"Show me all neighbors of node 'main.go'"

"Find all nodes of type 'function' that are connected to 'main.go'"

"Find paths between 'frontend.js' and 'backend.go' with max depth 3"
```

#### Code Analysis Workflows
```
"Analyze this codebase and create a function call graph in RelatixDB"

"Track the relationships between these modules in the graph database"

"Store the dependencies I just discussed in RelatixDB for future reference"
```

### Common Use Cases with Claude Code

#### 1. Codebase Mapping
- Store relationships between files, functions, and modules
- Track dependencies and imports
- Build call graphs and reference networks

#### 2. Development Context
- Remember previous discussions and decisions
- Track tool actions and their relationships
- Store architectural knowledge and patterns

#### 3. Refactoring Support
- Query impact analysis before changes
- Track relationships that might be affected
- Store refactoring history and rationale

#### 4. Documentation Generation
- Extract relationship patterns for documentation
- Generate dependency graphs and diagrams
- Track knowledge connections across the project

### Storage Recommendations

#### For Development Sessions
Use in-memory mode for fast, temporary storage:
```json
"command": "/path/to/relatixdb",
"args": ["-debug"]
```

#### For Persistent Projects
Use persistent storage to maintain context across sessions:
```json
"command": "/path/to/relatixdb",
"args": ["-db", "~/.claude-code/graphs/myproject.db", "-debug"]
```

### Troubleshooting

#### Connection Issues
1. Verify RelatixDB binary path is correct
2. Check file permissions on the binary
3. Ensure database directory exists (for persistent mode)
4. Review Claude Code logs for MCP server errors

#### Performance Considerations
- In-memory mode: Fastest, but data lost on restart
- Persistent mode: Slightly slower writes, data preserved
- Debug mode: Adds logging overhead, disable for production use

#### Database Location
Choose database location based on your workflow:
- Project-specific: `./project-graph.db` (committed with project)
- User-specific: `~/.claude-code/graphs/project.db` (personal, not committed)
- Session-specific: In-memory mode (temporary, fast)

### Advanced Configuration

#### Custom Tool Names
You can customize how RelatixDB appears in Claude Code by modifying the server name:

```json
{
  "mcp": {
    "servers": {
      "project-knowledge-graph": {
        "command": "/path/to/relatixdb",
        "args": ["-db", "./knowledge-graph.db"]
      }
    }
  }
}
```

#### Multiple Instances
Run multiple RelatixDB instances for different purposes:

```json
{
  "mcp": {
    "servers": {
      "code-graph": {
        "command": "/path/to/relatixdb",
        "args": ["-db", "./code-relationships.db"]
      },
      "decisions-graph": {
        "command": "/path/to/relatixdb", 
        "args": ["-db", "./architectural-decisions.db"]
      }
    }
  }
}
```

This integration enables Claude Code to leverage RelatixDB's high-performance graph capabilities for enhanced context management and relationship tracking in your development workflow.

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

**Benchmarked on Apple M4 Pro (64GB RAM):**

| Operation | Performance | Target | Improvement |
|-----------|-------------|--------|-------------|
| Node insertion | 1.084µs | <100µs | 92x faster |
| Edge insertion | 1.708µs | <150µs | 88x faster |
| Neighbor query | 1.042µs | <1ms | 959x faster |
| Path query (depth ≤ 4) | <10ms | <10ms | Meets target |

**Additional Benchmarks:**
- Node retrieval: 75.52ns per operation
- Complex queries: 166.1ns average
- Concurrent operations: 133.8ns per operation

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
- Performance benchmarks exceeding targets by 88-959x (M4 Pro)
- Comprehensive test suite

## Support

This is experimental software. Support is limited to:
- GitHub Issues for bug reports
- Discussions for feature requests
- Code review for contributions

**Remember**: This is pre-alpha software. Use for experimentation and development only.