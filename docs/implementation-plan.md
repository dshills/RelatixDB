# RelatixDB Implementation Plan

## Project Overview
RelatixDB is a high-performance, local graph database designed as an MCP (Model Context Protocol) tool server. It provides fast, local, contextual knowledge manipulation for LLMs through a JSON-based stdio interface.

## Architecture Overview

### Core Components
```
├── cmd/relatixdb/          # Main executable
├── internal/
│   ├── graph/             # Core graph data structures and operations
│   ├── mcp/               # MCP protocol handling
│   ├── storage/           # Persistent storage backends
│   └── query/             # Query engine implementation
├── pkg/                   # Public API (optional)
└── docs/                  # Documentation
```

## Implementation Phases

### Phase 1: Core Foundation (High Priority)

#### 1.1 Core Data Structures (`internal/graph/types.go`)
- **Node struct**: ID, Type, Properties (map[string]string)
- **Edge struct**: From, To, Label, Properties (map[string]string)
- **Graph interface**: AddNode, AddEdge, DeleteNode, DeleteEdge, Query methods
- JSON marshaling/unmarshaling support for MCP compatibility

#### 1.2 In-Memory Graph Storage (`internal/graph/memory.go`)
- **Primary storage**: HashMap-based node storage
- **Indexes**:
  - Node ID index (map[string]*Node)
  - Type index (map[string][]*Node)
  - Out-edge index (map[string][]*Edge)
  - In-edge index (map[string][]*Edge)
- **Multigraph support**: Allow multiple edges between same nodes

#### 1.3 MCP Command Parser (`internal/mcp/parser.go`)
- **Command struct**: cmd, args fields
- **Response struct**: ok, result, error fields
- JSON parsing for commands: add_node, add_edge, query, delete_node, delete_edge
- Input validation and error handling

#### 1.4 Basic CRUD Operations (`internal/graph/operations.go`)
- AddNode: Insert node with ID uniqueness check
- AddEdge: Insert edge with node existence validation
- DeleteNode: Remove node and all connected edges
- DeleteEdge: Remove specific edge by from/to/label

### Phase 2: Query Engine & Interface (Medium Priority)

#### 2.1 Query Engine (`internal/query/engine.go`)
- **Neighbors query**: 
  - Direction support (in/out/both)
  - Label filtering
  - Single-hop traversal
- **Path query**:
  - Depth-limited BFS/DFS
  - All paths between nodes
  - Cycle detection
- **Find query**:
  - Property-based search
  - Type-based search
  - Combined filters

#### 2.2 MCP Protocol Handler (`internal/mcp/handler.go`)
- **Stdio interface**: Read JSON from stdin, write JSON to stdout
- **Request processing**: Parse command, execute operation, format response
- **Error handling**: Structured error responses
- **Logging**: Optional debug logging to stderr

#### 2.3 Main CLI Executable (`cmd/relatixdb/main.go`)
- **Command-line arguments**:
  - Database file path
  - MCP mode vs API mode
  - Debug/verbose flags
- **Initialization**: Load/create database
- **Main loop**: Process MCP commands until EOF

### Phase 3: Persistence & Performance (Lower Priority)

#### 3.1 Persistent Storage (`internal/storage/`)
- **Backend interface**: Load, Save, Close methods
- **BoltDB implementation**:
  - Separate buckets for nodes, edges, indexes
  - Efficient serialization (JSON or msgpack)
  - Atomic transactions
- **Recovery**: Load existing database on startup

#### 3.2 Testing & Benchmarking
- **Unit tests**: Each component with mock dependencies
- **Integration tests**: Full MCP protocol scenarios
- **Benchmarks**: 
  - Node insertion < 100µs
  - Edge insertion < 150µs
  - Neighborhood query < 1ms
  - Path query (depth ≤ 4) < 10ms
- **Property-based testing**: Random graph generation

#### 3.3 Performance Optimization
- **Memory pooling**: Reuse objects to reduce GC pressure
- **Batch operations**: Bulk inserts/updates
- **Index optimization**: Efficient data structures
- **Lazy loading**: On-demand index building
- **Profiling**: CPU and memory profiling integration

## Performance Targets

| Operation | Target |
|-----------|--------|
| Node insertion | < 100µs |
| Edge insertion | < 150µs |
| Neighborhood query (1-hop) | < 1ms |
| Path query (depth ≤ 4) | < 10ms |

## Data Model Examples

### Node Example
```json
{
  "id": "func:handleLogin",
  "type": "function",
  "props": {
    "name": "handleLogin",
    "language": "go",
    "path": "auth/login.go"
  }
}
```

### Edge Example
```json
{
  "from": "func:handleLogin",
  "to": "file:auth",
  "label": "defined_in",
  "props": {
    "line": "42"
  }
}
```

### MCP Command Examples
```json
{"cmd": "add_node", "args": {"id": "file:auth", "type": "file", "props": {"path": "auth/login.go"}}}
{"cmd": "add_edge", "args": {"from": "func:handleLogin", "to": "file:auth", "label": "defined_in"}}
{"cmd": "query", "args": {"type": "neighbors", "node": "func:handleLogin", "direction": "out"}}
```

## Development Workflow

1. **Start with Phase 1**: Implement core data structures and basic operations
2. **Test incrementally**: Unit tests for each component
3. **MCP integration**: Early stdio interface testing
4. **Performance validation**: Benchmark against targets
5. **Persistence layer**: Add durability once core is stable
6. **Optimization**: Profile and optimize based on real usage

## Dependencies

- **Required**: Go 1.24+
- **Storage**: BoltDB (`go.etcd.io/bbolt`)
- **Testing**: Standard library + benchmarking
- **Optional**: Profiling tools (pprof)

## Validation Strategy

- **Unit tests**: 90%+ coverage
- **Integration tests**: Full MCP protocol scenarios
- **Performance tests**: Automated benchmarking
- **Manual testing**: Real-world graph scenarios
- **Property-based testing**: Fuzz testing with random graphs