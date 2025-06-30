# RelatixDB Usage Guide

> ⚠️ **PRE-ALPHA SOFTWARE**: This is experimental software. APIs may change without notice.

## Overview

RelatixDB provides a JSON-based stdio interface following the Model Context Protocol (MCP) for graph database operations. This document covers detailed usage patterns and examples.

## Basic Operations

### Starting RelatixDB

#### In-Memory Mode (Default)
```bash
./relatixdb
```
Data is stored in memory only and lost when the process exits.

#### Persistent Mode
```bash
./relatixdb -db mydata.db
```
Data is stored in a BoltDB file and persists across restarts.

#### Debug Mode
```bash
./relatixdb -debug -db mydata.db
```
Enables verbose logging to stderr (does not interfere with MCP protocol on stdout).

### Command Format

All commands follow this JSON format:
```json
{
  "cmd": "command_name",
  "args": {
    "parameter": "value"
  }
}
```

### Response Format

Successful responses:
```json
{
  "ok": true,
  "result": { ... }
}
```

Error responses:
```json
{
  "ok": false,
  "error": "error description"
}
```

## Node Operations

### Adding Nodes

#### Basic Node
```json
{"cmd": "add_node", "args": {"id": "user:alice", "type": "user"}}
```

#### Node with Properties
```json
{
  "cmd": "add_node",
  "args": {
    "id": "file:auth.go",
    "type": "file",
    "props": {
      "path": "/src/auth.go",
      "language": "go",
      "lines": "150"
    }
  }
}
```

### Getting Nodes
```json
{"cmd": "get_node", "args": {"id": "user:alice"}}
```

### Deleting Nodes
```json
{"cmd": "delete_node", "args": {"id": "user:alice"}}
```
⚠️ This also deletes all connected edges.

## Edge Operations

### Adding Edges

#### Basic Edge
```json
{
  "cmd": "add_edge",
  "args": {
    "from": "user:alice",
    "to": "user:bob",
    "label": "follows"
  }
}
```

#### Edge with Properties
```json
{
  "cmd": "add_edge",
  "args": {
    "from": "function:login",
    "to": "file:auth.go",
    "label": "defined_in",
    "props": {
      "line": "42",
      "visibility": "public"
    }
  }
}
```

### Deleting Edges
```json
{
  "cmd": "delete_edge",
  "args": {
    "from": "user:alice",
    "to": "user:bob",
    "label": "follows"
  }
}
```

## Query Operations

### Neighbor Queries

#### Outgoing Neighbors
```json
{
  "cmd": "query",
  "args": {
    "type": "neighbors",
    "node": "user:alice",
    "direction": "out"
  }
}
```

#### Incoming Neighbors
```json
{
  "cmd": "query",
  "args": {
    "type": "neighbors",
    "node": "user:alice",
    "direction": "in"
  }
}
```

#### All Neighbors
```json
{
  "cmd": "query",
  "args": {
    "type": "neighbors",
    "node": "user:alice",
    "direction": "both"
  }
}
```

#### Filtered by Edge Label
```json
{
  "cmd": "query",
  "args": {
    "type": "neighbors",
    "node": "user:alice",
    "direction": "out",
    "label": "follows"
  }
}
```

### Path Queries

#### Find Paths Between Nodes
```json
{
  "cmd": "query",
  "args": {
    "type": "paths",
    "from": "user:alice",
    "to": "user:charlie",
    "max_depth": 3
  }
}
```

### Property-Based Search

#### Find by Node Type
```json
{
  "cmd": "query",
  "args": {
    "type": "find",
    "filters": {
      "type": "user"
    }
  }
}
```

#### Find by Property
```json
{
  "cmd": "query",
  "args": {
    "type": "find",
    "filters": {
      "type": "file",
      "language": "go"
    }
  }
}
```

## Complete Examples

### Social Network Example

```bash
# Create users
echo '{"cmd": "add_node", "args": {"id": "user:alice", "type": "user", "props": {"name": "Alice", "age": "25"}}}' | ./relatixdb
echo '{"cmd": "add_node", "args": {"id": "user:bob", "type": "user", "props": {"name": "Bob", "age": "30"}}}' | ./relatixdb
echo '{"cmd": "add_node", "args": {"id": "user:charlie", "type": "user", "props": {"name": "Charlie", "age": "28"}}}' | ./relatixdb

# Create relationships
echo '{"cmd": "add_edge", "args": {"from": "user:alice", "to": "user:bob", "label": "follows"}}' | ./relatixdb
echo '{"cmd": "add_edge", "args": {"from": "user:bob", "to": "user:charlie", "label": "follows"}}' | ./relatixdb

# Query Alice's followers
echo '{"cmd": "query", "args": {"type": "neighbors", "node": "user:alice", "direction": "out"}}' | ./relatixdb

# Find path from Alice to Charlie
echo '{"cmd": "query", "args": {"type": "paths", "from": "user:alice", "to": "user:charlie", "max_depth": 3}}' | ./relatixdb
```

### Code Dependency Example

```bash
# Create code elements
echo '{"cmd": "add_node", "args": {"id": "file:main.go", "type": "file", "props": {"path": "main.go"}}}' | ./relatixdb
echo '{"cmd": "add_node", "args": {"id": "function:main", "type": "function", "props": {"name": "main"}}}' | ./relatixdb
echo '{"cmd": "add_node", "args": {"id": "function:login", "type": "function", "props": {"name": "login"}}}' | ./relatixdb

# Create relationships
echo '{"cmd": "add_edge", "args": {"from": "function:main", "to": "file:main.go", "label": "defined_in"}}' | ./relatixdb
echo '{"cmd": "add_edge", "args": {"from": "function:main", "to": "function:login", "label": "calls"}}' | ./relatixdb

# Find all functions in main.go
echo '{"cmd": "query", "args": {"type": "neighbors", "node": "file:main.go", "direction": "in", "label": "defined_in"}}' | ./relatixdb
```

## Interactive Usage

### Using with readline/rlwrap
```bash
# Install rlwrap for better interactive experience
brew install rlwrap  # macOS
sudo apt-get install rlwrap  # Ubuntu

# Run with readline support
rlwrap -a ./relatixdb -debug
```

### Pipe from File
```bash
# Create command file
cat > commands.json << 'EOF'
{"cmd": "add_node", "args": {"id": "test:1", "type": "test"}}
{"cmd": "add_node", "args": {"id": "test:2", "type": "test"}}
{"cmd": "add_edge", "args": {"from": "test:1", "to": "test:2", "label": "connects"}}
{"cmd": "query", "args": {"type": "neighbors", "node": "test:1", "direction": "out"}}
EOF

# Execute commands
cat commands.json | ./relatixdb -debug
```

## Error Handling

### Common Errors

#### Node Already Exists
```json
{"cmd": "add_node", "args": {"id": "duplicate", "type": "test"}}
{"cmd": "add_node", "args": {"id": "duplicate", "type": "test"}}
```
Response:
```json
{"ok": false, "error": "node already exists"}
```

#### Node Not Found
```json
{"cmd": "delete_node", "args": {"id": "nonexistent"}}
```
Response:
```json
{"ok": false, "error": "node not found"}
```

#### Invalid Edge
```json
{"cmd": "add_edge", "args": {"from": "missing", "to": "node:1", "label": "connects"}}
```
Response:
```json
{"ok": false, "error": "node not found"}
```

### Validation Errors

#### Empty Node ID
```json
{"cmd": "add_node", "args": {"id": "", "type": "test"}}
```
Response:
```json
{"ok": false, "error": "node ID cannot be empty"}
```

#### Invalid Query Direction
```json
{"cmd": "query", "args": {"type": "neighbors", "node": "test:1", "direction": "invalid"}}
```
Response:
```json
{"ok": false, "error": "invalid direction: must be 'in', 'out', or 'both'"}
```

## Performance Considerations

### Best Practices

1. **Batch Operations**: Send multiple commands in sequence rather than starting/stopping the process
2. **Use Persistent Storage**: For datasets larger than a few thousand nodes
3. **Index-Friendly IDs**: Use consistent ID patterns for better lookup performance
4. **Limit Path Depth**: Keep path queries under depth 4 for optimal performance

### Performance Characteristics

- **Node/Edge Lookups**: O(1) average case
- **Neighbor Queries**: O(k) where k is number of neighbors
- **Path Queries**: O(b^d) where b is branching factor, d is depth
- **Type Queries**: O(n) where n is nodes of that type

### Memory Usage

- **In-Memory Mode**: ~200 bytes per node, ~150 bytes per edge
- **Persistent Mode**: Additional overhead for transaction logging
- **Indexes**: Multiple indexes maintained for fast access

## Troubleshooting

### Debug Mode
Always use `-debug` flag when troubleshooting:
```bash
./relatixdb -debug -db mydata.db
```

### Common Issues

1. **Database Locked**: Only one process can access a BoltDB file at a time
2. **Permission Errors**: Ensure write permissions for database file location
3. **Invalid JSON**: Use a JSON validator to check command format
4. **Large Datasets**: Consider chunking large import operations

### Logs
Debug information is written to stderr, leaving stdout clean for MCP protocol.