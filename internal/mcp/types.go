package mcp

// This file is kept for backward compatibility but the main MCP protocol types
// are now defined in protocol.go. The old custom command types have been replaced
// with standard MCP JSON-RPC protocol implementation.

// Legacy command constants - deprecated in favor of standard MCP methods
const (
	// Deprecated: Use standard MCP tools instead
	CmdAddNode    = "add_node"
	CmdAddEdge    = "add_edge"
	CmdDeleteNode = "delete_node"
	CmdDeleteEdge = "delete_edge"
	CmdQuery      = "query"
)
