# RelatixDB Usage Guide

> ⚠️ **PRE-ALPHA SOFTWARE**: This is experimental software. APIs may change without notice.

## Overview

RelatixDB provides a standard Model Context Protocol (MCP) interface using JSON-RPC 2.0 over stdio for graph database operations. This document covers detailed usage patterns and examples.

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

## MCP Protocol Interface

### Server Initialization

Before using any tools, you must initialize the MCP server:

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

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "RelatixDB",
      "version": "1.0.0"
    }
  }
}
```

### Tool Discovery

Discover available tools:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "add_node",
        "description": "Add a node to the graph with ID, optional type, and properties",
        "inputSchema": {
          "type": "object",
          "properties": {
            "id": {"type": "string", "description": "Unique identifier for the node"},
            "type": {"type": "string", "description": "Optional type of the node"},
            "props": {"type": "object", "description": "Optional key/value properties"}
          },
          "required": ["id"]
        }
      }
      // ... more tools
    ]
  }
}
```

## MCP Tools Reference

### 1. add_node - Add Node to Graph

**Basic Node:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "add_node",
    "arguments": {
      "id": "user:alice",
      "type": "user"
    }
  }
}
```

**Node with Properties:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "add_node",
    "arguments": {
      "id": "file:auth.go",
      "type": "file",
      "props": {
        "path": "/src/auth.go",
        "language": "go",
        "lines": "150"
      }
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully added node 'file:auth.go' with type 'file'"
      }
    ]
  }
}
```

### 2. add_edge - Add Edge Between Nodes

**Basic Edge:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "add_edge",
    "arguments": {
      "from": "user:alice",
      "to": "user:bob",
      "label": "follows"
    }
  }
}
```

**Edge with Properties:**
```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "add_edge",
    "arguments": {
      "from": "function:login",
      "to": "file:auth.go",
      "label": "defined_in",
      "props": {
        "line": "42",
        "visibility": "public"
      }
    }
  }
}
```

### 3. query_neighbors - Find Connected Nodes

**Outgoing Neighbors:**
```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "tools/call",
  "params": {
    "name": "query_neighbors",
    "arguments": {
      "node": "user:alice",
      "direction": "out"
    }
  }
}
```

**All Neighbors with Label Filter:**
```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "method": "tools/call",
  "params": {
    "name": "query_neighbors",
    "arguments": {
      "node": "user:alice",
      "direction": "both",
      "label": "follows"
    }
  }
}
```

**Response Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Found 2 neighbors for node 'user:alice':\n- user:bob (type: user)\n- user:charlie (type: user)\n\nConnecting edges (2):\n- user:alice -> user:bob (follows)\n- user:alice -> user:charlie (follows)"
      }
    ]
  }
}
```

### 4. query_paths - Find Paths Between Nodes

```json
{
  "jsonrpc": "2.0",
  "id": 9,
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

**Response Example:**
```json
{
  "jsonrpc": "2.0",
  "id": 9,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Found 2 paths from 'user:alice' to 'user:charlie':\nPath 1: user:alice -> user:bob -> user:charlie\nPath 2: user:alice -> user:charlie"
      }
    ]
  }
}
```

### 5. query_find - Search Nodes by Criteria

**Find by Node Type:**
```json
{
  "jsonrpc": "2.0",
  "id": 10,
  "method": "tools/call",
  "params": {
    "name": "query_find",
    "arguments": {
      "type": "user"
    }
  }
}
```

**Find by Properties:**
```json
{
  "jsonrpc": "2.0",
  "id": 11,
  "method": "tools/call",
  "params": {
    "name": "query_find",
    "arguments": {
      "type": "file",
      "props": {
        "language": "go"
      }
    }
  }
}
```

### 6. delete_node - Remove Node and Connected Edges

```json
{
  "jsonrpc": "2.0",
  "id": 12,
  "method": "tools/call",
  "params": {
    "name": "delete_node",
    "arguments": {
      "id": "user:alice"
    }
  }
}
```

⚠️ This also deletes all connected edges.

### 7. delete_edge - Remove Specific Edge

```json
{
  "jsonrpc": "2.0",
  "id": 13,
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

## Complete Examples

### Social Network Example

```bash
#!/bin/bash

# Start RelatixDB and create a complete social network
{
  # Initialize server
  echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "example", "version": "1.0"}}}'
  
  # Create users
  echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "user:alice", "type": "user", "props": {"name": "Alice", "age": "25"}}}}'
  echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "user:bob", "type": "user", "props": {"name": "Bob", "age": "30"}}}}'
  echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "user:charlie", "type": "user", "props": {"name": "Charlie", "age": "28"}}}}'
  
  # Create relationships
  echo '{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "user:alice", "to": "user:bob", "label": "follows"}}}'
  echo '{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "user:bob", "to": "user:charlie", "label": "follows"}}}'
  
  # Query Alice's followers
  echo '{"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "query_neighbors", "arguments": {"node": "user:alice", "direction": "out"}}}'
  
  # Find path from Alice to Charlie
  echo '{"jsonrpc": "2.0", "id": 8, "method": "tools/call", "params": {"name": "query_paths", "arguments": {"from": "user:alice", "to": "user:charlie", "max_depth": 3}}}'
  
} | ./relatixdb -debug
```

### Code Dependency Example

```bash
#!/bin/bash

# Build a code dependency graph
{
  # Initialize server
  echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "code-analyzer", "version": "1.0"}}}'
  
  # Create code elements
  echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "file:main.go", "type": "file", "props": {"path": "main.go"}}}}'
  echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "function:main", "type": "function", "props": {"name": "main"}}}}'
  echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "function:login", "type": "function", "props": {"name": "login"}}}}'
  
  # Create relationships
  echo '{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "function:main", "to": "file:main.go", "label": "defined_in"}}}'
  echo '{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "function:main", "to": "function:login", "label": "calls"}}}'
  
  # Find all functions in main.go
  echo '{"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "query_neighbors", "arguments": {"node": "file:main.go", "direction": "in", "label": "defined_in"}}}'
  
  # Find all Go files
  echo '{"jsonrpc": "2.0", "id": 8, "method": "tools/call", "params": {"name": "query_find", "arguments": {"type": "file", "props": {"path": "*.go"}}}}'
  
} | ./relatixdb -debug
```

## Interactive Usage

### Using with JSON Lines Tools

```bash
# Install jq for JSON processing
brew install jq  # macOS
sudo apt-get install jq  # Ubuntu

# Pretty print responses
./relatixdb | jq '.'
```

### Pipe from File

```bash
# Create command file
cat > mcp_commands.jsonl << 'EOF'
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "test:1", "type": "test"}}}
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "test:2", "type": "test"}}}
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "test:1", "to": "test:2", "label": "connects"}}}
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "query_neighbors", "arguments": {"node": "test:1", "direction": "out"}}}
EOF

# Execute commands
cat mcp_commands.jsonl | ./relatixdb -debug
```

## Error Handling

### JSON-RPC Errors

#### Server Not Initialized
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {"name": "add_node", "arguments": {"id": "test"}}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "error": {
    "code": -32600,
    "message": "Server not initialized"
  }
}
```

#### Invalid Method
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "unknown/method"
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "error": {
    "code": -32601,
    "message": "Method not found"
  }
}
```

### Tool Errors

#### Missing Required Arguments
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "add_node",
    "arguments": {}
  }
}
```

**Response:**
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

#### Unknown Tool
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "unknown_tool",
    "arguments": {}
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Error: unknown tool: unknown_tool"
      }
    ],
    "isError": true
  }
}
```

## Performance Considerations

### Best Practices

1. **Initialize Once**: Send the initialize request only once per session
2. **Batch Tool Calls**: Use sequential IDs to send multiple tool calls efficiently
3. **Use Persistent Storage**: For datasets larger than a few thousand nodes
4. **Index-Friendly IDs**: Use consistent ID patterns for better lookup performance
5. **Limit Path Depth**: Keep path queries under depth 4 for optimal performance

### Performance Characteristics

- **Node/Edge Operations**: O(1) average case
- **Neighbor Queries**: O(k) where k is number of neighbors
- **Path Queries**: O(b^d) where b is branching factor, d is depth
- **Find Queries**: O(n) where n is nodes matching criteria

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

1. **Server Not Initialized**: Must send `initialize` request before using tools
2. **Invalid JSON-RPC**: Ensure `jsonrpc: "2.0"` and proper structure
3. **Database Locked**: Only one process can access a BoltDB file at a time
4. **Permission Errors**: Ensure write permissions for database file location
5. **Invalid Tool Arguments**: Check tool schemas with `tools/list`

### Logs
Debug information is written to stderr, leaving stdout clean for MCP protocol.

### Testing Tool Schemas

Use `tools/list` to see the exact input schema for each tool:
```bash
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0"}}}' | ./relatixdb
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | ./relatixdb | jq '.result.tools[0].inputSchema'
```