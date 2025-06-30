# Local Graph Database Specification for MCP Tool Server

## Overview
This document defines the specification for a high-performance, local graph database intended for use as a Model Context Protocol (MCP) tool. It will be operated by a local LLM agent or model orchestrator, primarily via standard I/O streams following the MCP tool interface. This is not an enterprise-grade graph database—it is optimized for fast, local, contextual knowledge manipulation by LLMs.

## Primary Use Cases
- Enabling LLMs to reason about entities and their relationships
- Storing relationships between source code elements (files, functions, modules)
- Tracking tool actions and derivations (e.g., "this file was generated from X")
- Supporting lightweight planning and memory for AI agents

## Design Goals
- Blazing fast local access (single-process, in-memory or mmap-backed)
- Immutable or append-only with occasional compaction
- MCP stdio-based command interface
- Optional embedded API (Go package)
- JSON-formatted responses for ease of LLM parsing

## Non-Goals
- Distributed operation
- Multi-user concurrency
- Persistent high-availability storage
- Cypher/Gremlin support or query language support beyond custom commands

---

## Data Model

### Nodes
- Each node is identified by a unique string ID
- Nodes may have types (e.g., `file`, `function`, `module`, `prompt`, `tool`, `user`)
- Arbitrary key/value properties are allowed (string values only)

```json
{
  "id": "func:1234",
  "type": "function",
  "props": {
    "name": "handleLogin",
    "language": "go",
    "path": "auth/login.go"
  }
}
```

### Edges
- Directed, labeled edges between nodes
- Edge label is a required string (e.g., `calls`, `imports`, `generated_from`, `owned_by`)
- Optional key/value metadata on the edge

```json
{
  "from": "func:1234",
  "to": "file:5678",
  "label": "defined_in"
}
```

### Graph Characteristics
- Multigraph: multiple edges of different types between same nodes allowed
- No enforced schema, but conventions are encouraged for tool interop

---

## Storage Engine

### Backing Store
- LMDB, BoltDB, or custom mmap-based store
- Each graph is its own file/directory
- Read-optimized, with append-only or journaling for writes

### Indexes
- Node ID index (required)
- Type index (optional, for fast node-type queries)
- Out-edge index (from-node based)
- In-edge index (to-node based)

---

## MCP Tool Server Mode

### Command Format (stdin)
Each line of input is a JSON object with the following fields:

```json
{
  "cmd": "add_node" | "add_edge" | "query" | "delete_node" | "delete_edge",
  "args": { ... }
}
```

### Response Format (stdout)
Each response is a single line of JSON with:
```json
{
  "ok": true | false,
  "result": { ... },
  "error": "error message if any"
}
```

### Supported Commands
#### `add_node`
```json
{
  "cmd": "add_node",
  "args": {
    "id": "node_id",
    "type": "node_type",
    "props": { "key": "val" }
  }
}
```

#### `add_edge`
```json
{
  "cmd": "add_edge",
  "args": {
    "from": "node_id",
    "to": "node_id",
    "label": "edge_label",
    "props": { "weight": "1.0" }
  }
}
```

#### `query`
Supported subcommands:
- `neighbors`: get direct neighbors from a node
- `paths`: all paths (depth-limited) between two nodes
- `find`: by property or type

```json
{
  "cmd": "query",
  "args": {
    "type": "neighbors",
    "node": "node_id",
    "label": "calls",
    "direction": "out" | "in"
  }
}
```

---

## Go API Interface (Optional)
```go
type Graph interface {
  AddNode(ctx context.Context, node Node) error
  AddEdge(ctx context.Context, edge Edge) error
  DeleteNode(ctx context.Context, id string) error
  DeleteEdge(ctx context.Context, from, to, label string) error
  Query(ctx context.Context, q Query) (any, error)
}

type Node struct {
  ID    string
  Type  string
  Props map[string]string
}

type Edge struct {
  From  string
  To    string
  Label string
  Props map[string]string
}

type Query struct {
  Type      string // "neighbors", "paths", "find"
  Node      string
  Label     string
  Direction string
  MaxDepth  int
  Filters   map[string]string
}
```

---

## Performance Targets
- **Node insertion:** < 100µs
- **Edge insertion:** < 150µs
- **Neighborhood query (1-hop):** < 1ms
- **Path query (depth ≤ 4):** < 10ms

---

## Optional Features
- Node tagging and fulltext index (for enhanced prompt querying)
- Snapshots (read-only export of graph state)
- Compaction/GC tool
- Subgraph extraction and re-injection

---

## Example Usage (MCP Interaction)
```json
{"cmd": "add_node", "args": {"id": "file:auth", "type": "file", "props": {"path": "auth/login.go"}}}
{"cmd": "add_node", "args": {"id": "func:handleLogin", "type": "function", "props": {"name": "handleLogin"}}}
{"cmd": "add_edge", "args": {"from": "func:handleLogin", "to": "file:auth", "label": "defined_in"}}
{"cmd": "query", "args": {"type": "neighbors", "node": "func:handleLogin", "direction": "out"}}
```

---

## Future Directions
- Allow hybrid usage with vector DBs (graph references vector IDs)
- Use in MCP memory systems for toolchain-level planning
- Add provenance tracking on edges (LLM/tool responsible, timestamp, certainty)
- Plugin model for agent-defined edge types and node kinds
