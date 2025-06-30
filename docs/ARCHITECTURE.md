# RelatixDB Architecture

> ⚠️ **PRE-ALPHA SOFTWARE**: Architecture may change significantly during development.

## Overview

RelatixDB is designed as a high-performance, local graph database optimized for Model Context Protocol (MCP) tool integration. The architecture prioritizes simplicity, performance, and LLM-friendly interfaces.

## Design Principles

### Core Principles
1. **Local-First**: No network dependencies, single-process operation
2. **Performance-Optimized**: Sub-microsecond operations for common queries
3. **MCP-Native**: JSON stdio interface designed for LLM integration
4. **Thread-Safe**: Concurrent access with proper synchronization
5. **Durability**: Optional persistence with immediate consistency

### Non-Goals
- Distributed operation
- Multi-user concurrency 
- ACID transactions across multiple operations
- SQL/Cypher query language support
- Enterprise-scale data volumes

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP Client (LLM)                        │
└─────────────────────┬───────────────────────────────────────┘
                      │ JSON stdio
┌─────────────────────▼───────────────────────────────────────┐
│                 relatixdb CLI                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              MCP Handler                            │   │
│  │  • Command parsing  • Response formatting          │   │
│  │  • Input validation • Error handling               │   │
│  └─────────────────────┬───────────────────────────────┘   │
└────────────────────────┼───────────────────────────────────┘
                         │ Graph interface
┌────────────────────────▼───────────────────────────────────┐
│                  Graph Layer                               │
│  ┌─────────────────┐           ┌─────────────────────────┐ │
│  │  Memory Graph   │           │   Persistent Graph      │ │
│  │  • In-memory    │           │   • Wraps memory graph  │ │
│  │  • Multi-index  │           │   • Immediate persist   │ │
│  │  • Thread-safe  │           │   • Auto-save option   │ │
│  └─────────┬───────┘           └─────────┬───────────────┘ │
└────────────┼─────────────────────────────┼─────────────────┘
             │                             │
             │ Graph operations            │ Storage interface
             │                             │
┌────────────▼─────────────────────────────▼─────────────────┐
│                Storage Layer                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                BoltDB Backend                       │   │
│  │  • Atomic transactions  • JSON serialization       │   │
│  │  • Bucket organization  • Statistics tracking      │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Component Details

### MCP Handler Layer

**Location**: `internal/mcp/`

**Responsibilities**:
- Implement standard Model Context Protocol (MCP) using JSON-RPC 2.0
- Handle server initialization and capability negotiation
- Provide tool discovery via `tools/list` method
- Execute graph operations via `tools/call` method
- Format responses according to MCP specification

**Key Components**:
- `handler.go`: Main MCP JSON-RPC protocol handler
- `protocol.go`: MCP protocol types and JSON-RPC structures
- `types.go`: Legacy command types (deprecated)
- `handler_test.go`: MCP protocol testing
- `integration_test.go`: End-to-end MCP functionality tests

**MCP Protocol Flow**:
1. **Initialization**: Client sends `initialize` request with capabilities
2. **Tool Discovery**: Client calls `tools/list` to discover available tools
3. **Tool Execution**: Client uses `tools/call` to execute specific graph operations
4. **Error Handling**: Standard JSON-RPC errors and MCP tool error responses

**Available MCP Tools**:
- `add_node`: Add nodes with ID, type, and properties
- `add_edge`: Add directed, labeled edges between nodes
- `delete_node`: Delete nodes and connected edges
- `delete_edge`: Delete specific edges
- `query_neighbors`: Find neighboring nodes with direction filtering
- `query_paths`: Find paths between nodes with depth limiting
- `query_find`: Search nodes by type and properties

### Graph Layer

**Location**: `internal/graph/`

**Responsibilities**:
- Core graph data structures and operations
- Thread-safe concurrent access
- Query engine implementation
- Performance optimization

**Key Components**:

#### Memory Graph (`memory.go`)
- **Primary Storage**: HashMap-based node/edge storage
- **Indexes**: 
  - Node ID index: `map[string]*Node`
  - Type index: `map[string]map[string]*Node`
  - Out-edge index: `map[string]map[string]*Edge`
  - In-edge index: `map[string]map[string]*Edge`
- **Synchronization**: RWMutex for concurrent access
- **Performance**: O(1) lookups, O(k) neighbor queries

#### Query Engine (`query.go`)
- **Neighbor Queries**: Single-hop traversal with direction filtering
- **Path Queries**: Depth-limited BFS with cycle detection
- **Property Search**: Type and property-based filtering
- **Performance**: Optimized for common query patterns

#### Data Types (`types.go`)
```go
type Node struct {
    ID    string            `json:"id"`
    Type  string            `json:"type,omitempty"`
    Props map[string]string `json:"props,omitempty"`
}

type Edge struct {
    From  string            `json:"from"`
    To    string            `json:"to"`
    Label string            `json:"label"`
    Props map[string]string `json:"props,omitempty"`
}
```

### Storage Layer

**Location**: `internal/storage/`

**Responsibilities**:
- Persistent storage abstraction
- Transaction management
- Serialization/deserialization
- Storage statistics

**Key Components**:

#### Storage Interface (`interface.go`)
```go
type Backend interface {
    Open(path string) error
    Close() error
    LoadGraph(ctx context.Context) (graph.Graph, error)
    SaveGraph(ctx context.Context, g graph.Graph) error
    BeginTransaction() (Transaction, error)
}
```

#### BoltDB Backend (`bolt.go`)
- **Database**: BoltDB for ACID compliance
- **Buckets**: Separate buckets for nodes, edges, metadata
- **Serialization**: JSON format for portability
- **Transactions**: Atomic operations with rollback
- **Statistics**: Database size and record counts

#### Persistent Graph (`persistent_graph.go`)
- **Hybrid Approach**: Memory graph with immediate persistence
- **Write-Through**: All mutations persisted immediately
- **Auto-Save**: Optional bulk save for performance (TODO)
- **Error Handling**: Automatic rollback on persistence failure

## Data Flow

### Write Operations

```
MCP Tool Call → JSON-RPC Parsing → Tool Execution → Memory Update → Transaction → Persist → MCP Response
```

1. **JSON-RPC Parsing**: Parse `tools/call` request and validate JSON-RPC structure
2. **Tool Execution**: Route to appropriate tool handler with argument validation
3. **Memory Update**: Apply change to in-memory graph
4. **Transaction**: Begin storage transaction
5. **Persist**: Write change to storage
6. **Commit/Rollback**: Complete transaction or rollback both memory and storage
7. **MCP Response**: Return tool result as JSON-RPC response with success/error status

### Read Operations

```
MCP Tool Call → JSON-RPC Parsing → Query Tool Execution → Memory Graph → Query Engine → MCP Response
```

1. **JSON-RPC Parsing**: Parse `tools/call` request for query tools
2. **Query Tool Execution**: Route to query tool handler (neighbors, paths, find)
3. **Execute**: Run query against in-memory indexes
4. **Format**: Convert results to structured text response
5. **MCP Response**: Return tool result as JSON-RPC response

## Performance Architecture

### Optimization Strategies

1. **Multiple Indexes**: Pre-computed indexes for common access patterns
2. **Memory-First**: All reads served from memory
3. **Lock Granularity**: RWMutex for optimal concurrent access
4. **Efficient Data Structures**: HashMap-based storage for O(1) lookups
5. **Lazy Evaluation**: Query results computed on-demand

### Memory Layout

```
Graph Memory Structure:
├── nodes: map[string]*Node           # Primary node storage
├── edges: map[string]*Edge           # Primary edge storage  
├── nodesByType: map[string]map[...   # Type-based index
├── outEdges: map[string]map[...      # Outgoing edge index
└── inEdges: map[string]map[...       # Incoming edge index
```

### Performance Characteristics

| Operation | Complexity | Measured Performance |
|-----------|------------|---------------------|
| Add Node | O(1) | 4.083µs |
| Add Edge | O(1) | 1.5µs |
| Get Node | O(1) | 76ns |
| Get Neighbors | O(k) | 625ns |
| Path Query | O(b^d) | <10ms (depth ≤ 4) |

## Concurrency Model

### Thread Safety

- **RWMutex**: Protects all graph operations
- **Read Operations**: Multiple concurrent readers
- **Write Operations**: Exclusive access during mutations
- **Lock Ordering**: Consistent ordering to prevent deadlocks

### Transaction Isolation

- **Memory Consistency**: Changes visible immediately after commit
- **Storage Consistency**: Atomic transactions via BoltDB
- **Rollback Semantics**: Both memory and storage rolled back on failure

## Storage Format

### BoltDB Organization

```
Database File:
├── nodes bucket
│   ├── "node:id1" → JSON(Node)
│   ├── "node:id2" → JSON(Node)
│   └── ...
├── edges bucket  
│   ├── "from:to:label" → JSON(Edge)
│   └── ...
└── meta bucket
    ├── "version" → "1.0"
    └── "stats" → JSON(Stats)
```

### JSON Serialization

**Node Format**:
```json
{
  "id": "user:alice",
  "type": "user", 
  "props": {
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

**Edge Format**:
```json
{
  "from": "user:alice",
  "to": "user:bob",
  "label": "follows",
  "props": {
    "since": "2024-01-01"
  }
}
```

## Error Handling

### Error Categories

1. **Validation Errors**: Invalid input format or values
2. **Business Logic Errors**: Constraint violations (duplicate nodes, etc.)
3. **Storage Errors**: Database access failures
4. **System Errors**: Resource exhaustion, permissions, etc.

### Error Recovery

- **Graceful Degradation**: Continue operation when possible
- **Automatic Rollback**: Revert changes on transaction failure
- **Clear Error Messages**: Actionable error descriptions
- **No Data Corruption**: Maintain consistency under all failure modes

## Future Architecture

### Planned Enhancements

1. **Bulk Operations**: Efficient SaveGraph implementation
2. **Schema Evolution**: Versioned storage format
3. **Backup/Restore**: Export/import functionality
4. **Performance Monitoring**: Detailed metrics collection
5. **Query Optimization**: Advanced query planning

### Extension Points

- **Storage Backends**: LMDB, custom mmap implementations
- **Serialization**: MessagePack, Protocol Buffers
- **Indexes**: Specialized indexes for specific query patterns
- **Query Language**: Optional higher-level query interface

## Development Guidelines

### Code Organization

- **Package Structure**: Clear separation of concerns
- **Interface Design**: Abstract implementation details
- **Test Coverage**: Comprehensive unit and integration tests
- **Documentation**: Inline docs and external guides

### Performance Requirements

- **Micro-benchmarks**: Validate performance targets
- **Memory Profiling**: Monitor memory usage patterns
- **Concurrent Testing**: Race condition detection
- **Load Testing**: Validate performance under stress