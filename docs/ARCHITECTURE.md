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
- Parse JSON commands from stdin
- Validate command structure and arguments
- Route commands to appropriate graph operations
- Format responses as JSON to stdout
- Handle errors gracefully

**Key Components**:
- `handler.go`: Main MCP protocol handler
- `types.go`: Command/response type definitions
- `handler_test.go`: Protocol testing

**Command Flow**:
1. Read JSON line from stdin
2. Parse into Command struct
3. Validate command and arguments
4. Execute via Graph interface
5. Format result as Response
6. Write JSON line to stdout

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
MCP Command → Validation → Memory Update → Transaction → Persist → Response
```

1. **Validation**: Check command structure and business rules
2. **Memory Update**: Apply change to in-memory graph
3. **Transaction**: Begin storage transaction
4. **Persist**: Write change to storage
5. **Commit/Rollback**: Complete transaction or rollback both memory and storage
6. **Response**: Return success/error to client

### Read Operations

```
MCP Query → Memory Graph → Query Engine → Response
```

1. **Parse Query**: Extract query type and parameters
2. **Execute**: Run query against in-memory indexes
3. **Format**: Convert results to JSON response
4. **Return**: Send response to client

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