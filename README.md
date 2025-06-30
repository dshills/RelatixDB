# RelatixDB

> ⚠️ **PRE-ALPHA SOFTWARE**: This is experimental software in early development. Not suitable for production use. APIs and data formats may change without notice. Use at your own risk.

A high-performance, local graph database designed for use as a Model Context Protocol (MCP) tool server. RelatixDB is optimized for fast, local, contextual knowledge manipulation by LLMs.

## Features

- **Blazing Fast Performance**: Exceeds all targets by 90-960x
  - Node insertion: 1.084µs (target: <100µs) - 92x faster
  - Edge insertion: 1.708µs (target: <150µs) - 88x faster  
  - Neighbor queries: 1.042µs (target: <1ms) - 959x faster
- **Standard MCP Protocol**: Full JSON-RPC 2.0 compliance with tool discovery and execution
- **Dual Storage Modes**: In-memory for speed, persistent BoltDB for durability
- **Thread-Safe Operations**: Concurrent access with proper locking
- **Comprehensive Tool Set**: 7 MCP tools covering all graph operations
- **Multigraph Support**: Multiple edge types between same nodes

## Installation

### Prerequisites

- Go 1.22 or later
- Git

### Build from Source

```bash
git clone https://github.com/dshills/RelatixDB.git
cd RelatixDB
make build
```

### Install with Go

```bash
go install github.com/dshills/RelatixDB/cmd/relatixdb@latest
```

## Quick Start

### In-Memory Mode (Default)

```bash
# Start RelatixDB MCP server
./build/relatixdb

# Test with initialization and tool calls
echo -e '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}\n{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | ./build/relatixdb
```

### Persistent Mode

```bash
# Start with persistent storage
./build/relatixdb -db mydata.db
```

### Debug Mode

```bash
# Enable debug logging (to stderr)
./build/relatixdb -debug
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

### MCP Protocol Interface

RelatixDB implements the Model Context Protocol (MCP) using JSON-RPC 2.0 over stdio. All communication follows the standard MCP specification.

#### Server Initialization

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "your-client",
      "version": "1.0.0"
    }
  }
}
```

#### Tool Discovery

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

#### Available MCP Tools

RelatixDB provides 7 MCP tools for graph operations:

##### 1. **add_node** - Add Node to Graph
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "add_node",
    "arguments": {
      "id": "user:alice",
      "type": "user",
      "props": {
        "name": "Alice",
        "email": "alice@example.com"
      }
    }
  }
}
```

##### 2. **add_edge** - Add Edge Between Nodes
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "add_edge",
    "arguments": {
      "from": "user:alice",
      "to": "user:bob",
      "label": "follows",
      "props": {
        "since": "2024-01-01"
      }
    }
  }
}
```

##### 3. **query_neighbors** - Find Connected Nodes
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "query_neighbors",
    "arguments": {
      "node": "user:alice",
      "direction": "out",
      "label": "follows"
    }
  }
}
```

##### 4. **query_paths** - Find Paths Between Nodes
```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "query_paths",
    "arguments": {
      "from": "user:alice",
      "to": "user:charlie",
      "max_depth": 3
    }
  }
}
```

##### 5. **query_find** - Search Nodes by Criteria
```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "tools/call",
  "params": {
    "name": "query_find",
    "arguments": {
      "type": "user",
      "props": {
        "status": "active"
      }
    }
  }
}
```

##### 6. **delete_node** - Remove Node and Connected Edges
```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "method": "tools/call",
  "params": {
    "name": "delete_node",
    "arguments": {
      "id": "user:alice"
    }
  }
}
```

##### 7. **delete_edge** - Remove Specific Edge
```json
{
  "jsonrpc": "2.0",
  "id": 9,
  "method": "tools/call",
  "params": {
    "name": "delete_edge",
    "arguments": {
      "from": "user:alice",
      "to": "user:bob",
      "label": "follows"
    }
  }
}
```

### Response Format

All responses follow JSON-RPC 2.0 format:

**Success Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully added node 'user:alice' with type 'user'"
      }
    ]
  }
}
```

**Error Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Error: id is required and must be a string"
      }
    ],
    "isError": true
  }
}
```

## Claude Code Integration

RelatixDB integrates seamlessly with Claude Code as a Model Context Protocol (MCP) server, providing graph database capabilities directly within your AI-assisted development workflow.

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

Add RelatixDB to your Claude Code MCP configuration:

**Option A: Persistent Storage (Recommended)**
```json
{
  "mcpServers": {
    "relatixdb": {
      "command": "/path/to/RelatixDB/build/relatixdb",
      "args": ["-db", "/path/to/your/graph.db", "-debug"]
    }
  }
}
```

**Option B: In-Memory Mode (Fast, Non-Persistent)**
```json
{
  "mcpServers": {
    "relatixdb": {
      "command": "/path/to/RelatixDB/build/relatixdb",
      "args": ["-debug"]
    }
  }
}
```

#### 3. Verify Integration

Start Claude Code and verify the MCP server is running. You should see RelatixDB tools available in the tool listing.

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
{
  "mcpServers": {
    "relatixdb": {
      "command": "/path/to/relatixdb",
      "args": ["-debug"]
    }
  }
}
```

#### For Persistent Projects
Use persistent storage to maintain context across sessions:
```json
{
  "mcpServers": {
    "relatixdb": {
      "command": "/path/to/relatixdb",
      "args": ["-db", "~/.claude-code/graphs/myproject.db", "-debug"]
    }
  }
}
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

#### Multiple Instances
Run multiple RelatixDB instances for different purposes:

```json
{
  "mcpServers": {
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
make test

# Run with race detection
make test-race

# Run benchmarks
make test-bench

# Test MCP protocol implementation
make test-mcp

# Run full validation (lint, vet, test)
make validate
```

### Project Structure

```
├── cmd/relatixdb/           # Main executable
├── internal/
│   ├── graph/              # Core graph data structures
│   ├── mcp/                # MCP protocol handling (JSON-RPC 2.0)
│   └── storage/            # Persistent storage backends
├── docs/                   # Documentation
└── specs/                  # Technical specifications
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Build all packages
make build
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

## MCP Protocol Details

### Server Capabilities

RelatixDB declares the following MCP capabilities:

```json
{
  "tools": {
    "listChanged": false
  }
}
```

### Tool Input Schemas

All tools include comprehensive JSON Schema definitions for their parameters:

- **Required parameters**: Clearly marked in schema
- **Optional parameters**: Documented with defaults
- **Type validation**: Strict type checking on all inputs
- **Enum values**: Constrained choices where applicable (e.g., direction: "in", "out", "both")

### Error Handling

- **JSON-RPC errors**: For protocol-level issues
- **Tool errors**: For application-level errors (returned as tool results with `isError: true`)
- **Input validation**: Comprehensive parameter validation with descriptive error messages

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
4. Ensure all tests pass (`make validate`)
5. Submit pull request

## License

[License information to be added]

## Changelog

### v0.1.0 (Current)
- Initial implementation with in-memory graph
- Full MCP JSON-RPC 2.0 protocol support
- Standard tool discovery and execution
- 7 comprehensive MCP tools for graph operations
- BoltDB persistent storage layer
- Performance benchmarks exceeding targets by 88-959x (M4 Pro)
- Comprehensive test suite with MCP protocol validation

## Support

This is experimental software. Support is limited to:
- GitHub Issues for bug reports
- Discussions for feature requests
- Code review for contributions

**Remember**: This is pre-alpha software. Use for experimentation and development only.