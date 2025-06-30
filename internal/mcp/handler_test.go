package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/dshills/RelatixDB/internal/graph"
)

func TestHandler_ProcessSingleRequest(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()

	// Initialize the server first
	initReq := `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}`

	response, err := handler.ProcessSingleRequest(ctx, initReq)
	if err != nil {
		t.Fatalf("Expected no error for initialization, got %v", err)
	}

	if !strings.Contains(response, `"result"`) {
		t.Fatalf("Expected successful initialization response, got %s", response)
	}

	// Test add_node tool call
	addNodeReq := `{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "test:1", "type": "test", "props": {"name": "Test Node"}}}}`

	response, err = handler.ProcessSingleRequest(ctx, addNodeReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"result"`) || strings.Contains(response, `"error"`) {
		t.Fatalf("Expected successful response, got %s", response)
	}

	// Test add_edge tool call (should succeed after adding target node)
	addNode2Req := `{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "test:2", "type": "test"}}}`

	_, err = handler.ProcessSingleRequest(ctx, addNode2Req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	addEdgeReq := `{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "test:1", "to": "test:2", "label": "connects"}}}`

	response, err = handler.ProcessSingleRequest(ctx, addEdgeReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"result"`) || strings.Contains(response, `"error"`) {
		t.Fatalf("Expected successful response, got %s", response)
	}

	// Test query_neighbors tool call
	queryReq := `{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "query_neighbors", "arguments": {"node": "test:1", "direction": "out"}}}`

	response, err = handler.ProcessSingleRequest(ctx, queryReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"result"`) || strings.Contains(response, `"error"`) {
		t.Fatalf("Expected successful response, got %s", response)
	}
}

func TestHandler_ToolsList(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()

	// Initialize the server first
	initReq := `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}`

	_, err := handler.ProcessSingleRequest(ctx, initReq)
	if err != nil {
		t.Fatalf("Expected no error for initialization, got %v", err)
	}

	// Test tools/list request
	listReq := `{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}`

	response, err := handler.ProcessSingleRequest(ctx, listReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"tools"`) {
		t.Fatalf("Expected tools list in response, got %s", response)
	}

	// Check for expected tools
	expectedTools := []string{"add_node", "add_edge", "delete_node", "delete_edge", "query_neighbors", "query_paths", "query_find"}
	for _, tool := range expectedTools {
		if !strings.Contains(response, tool) {
			t.Fatalf("Expected tool '%s' in response, got %s", tool, response)
		}
	}
}

func TestParseJSONRPCRequest(t *testing.T) {
	// Test valid request
	reqJSON := `{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}`

	req, err := ParseJSONRPCRequest([]byte(reqJSON))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if req.Method != "tools/list" {
		t.Fatalf("Expected method 'tools/list', got %s", req.Method)
	}

	if req.JSONRpc != "2.0" {
		t.Fatalf("Expected jsonrpc '2.0', got %s", req.JSONRpc)
	}

	// Test invalid JSON
	invalidJSON := `{"jsonrpc": "2.0", "method":}`

	_, err = ParseJSONRPCRequest([]byte(invalidJSON))
	if err == nil {
		t.Fatalf("Expected error for invalid JSON")
	}

	// Test missing method field
	missingMethod := `{"jsonrpc": "2.0", "id": 1}`

	_, err = ParseJSONRPCRequest([]byte(missingMethod))
	if err == nil {
		t.Fatalf("Expected error for missing method field")
	}

	// Test invalid JSON-RPC version
	invalidVersion := `{"jsonrpc": "1.0", "id": 1, "method": "test"}`

	_, err = ParseJSONRPCRequest([]byte(invalidVersion))
	if err == nil {
		t.Fatalf("Expected error for invalid JSON-RPC version")
	}
}

func TestHandler_ErrorHandling(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()

	// Test uninitialized server
	listReq := `{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}`

	response, err := handler.ProcessSingleRequest(ctx, listReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"error"`) {
		t.Fatalf("Expected error response for uninitialized server, got %s", response)
	}

	// Initialize the server
	initReq := `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}`

	_, err = handler.ProcessSingleRequest(ctx, initReq)
	if err != nil {
		t.Fatalf("Expected no error for initialization, got %v", err)
	}

	// Test unknown method
	unknownReq := `{"jsonrpc": "2.0", "id": 2, "method": "unknown/method"}`

	response, err = handler.ProcessSingleRequest(ctx, unknownReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"error"`) {
		t.Fatalf("Expected error response for unknown method, got %s", response)
	}

	// Test invalid tool call
	invalidToolReq := `{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "unknown_tool", "arguments": {}}}`

	response, err = handler.ProcessSingleRequest(ctx, invalidToolReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"isError":true`) {
		t.Fatalf("Expected error response for unknown tool, got %s", response)
	}

	// Test tool call with invalid arguments
	invalidArgsReq := `{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "add_node", "arguments": {}}}`

	response, err = handler.ProcessSingleRequest(ctx, invalidArgsReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.Contains(response, `"isError":true`) {
		t.Fatalf("Expected error response for invalid arguments, got %s", response)
	}
}
