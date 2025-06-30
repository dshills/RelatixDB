package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/dshills/RelatixDB/internal/graph"
)

// TestMCPToolFunctions tests all MCP tool functions end-to-end using JSON-RPC protocol
func TestMCPToolFunctions(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()

	// Initialize the server first
	initReq := `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}`
	_, err := handler.ProcessSingleRequest(ctx, initReq)
	if err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	t.Run("add_node operations", func(t *testing.T) {
		testAddNodeOperations(ctx, t, handler)
	})

	t.Run("add_edge operations", func(t *testing.T) {
		testAddEdgeOperations(ctx, t, handler)
	})

	t.Run("query operations", func(t *testing.T) {
		testQueryOperations(ctx, t, handler)
	})

	t.Run("delete operations", func(t *testing.T) {
		testDeleteOperations(ctx, t, handler)
	})

	t.Run("error handling", func(t *testing.T) {
		testErrorHandling(ctx, t, handler)
	})

	t.Run("complex scenarios", func(t *testing.T) {
		testComplexScenarios(ctx, t, handler)
	})
}

func testAddNodeOperations(ctx context.Context, t *testing.T, handler *Handler) {
	// Test basic node addition
	addNodeReq := `{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "user:alice", "type": "user", "props": {"name": "Alice", "email": "alice@example.com"}}}}`
	response, err := handler.ProcessSingleRequest(ctx, addNodeReq)
	if err != nil {
		t.Fatalf("Failed to process add_node request: %v", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}

	// Test node without properties
	addNodeReq2 := `{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "user:bob", "type": "user"}}}`
	response, err = handler.ProcessSingleRequest(ctx, addNodeReq2)
	if err != nil {
		t.Fatalf("Failed to process add_node request: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}

	// Test node with minimal data (just ID)
	addNodeReq3 := `{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "product:123"}}}`
	response, err = handler.ProcessSingleRequest(ctx, addNodeReq3)
	if err != nil {
		t.Fatalf("Failed to process add_node request: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}
}

func testAddEdgeOperations(ctx context.Context, t *testing.T, handler *Handler) {
	// Test edge addition between existing nodes
	addEdgeReq := `{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "user:alice", "to": "user:bob", "label": "friends", "props": {"since": "2023"}}}}`
	response, err := handler.ProcessSingleRequest(ctx, addEdgeReq)
	if err != nil {
		t.Fatalf("Failed to process add_edge request: %v", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}

	// Test edge without properties
	addEdgeReq2 := `{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "user:alice", "to": "product:123", "label": "purchased"}}}`
	response, err = handler.ProcessSingleRequest(ctx, addEdgeReq2)
	if err != nil {
		t.Fatalf("Failed to process add_edge request: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}
}

func testQueryOperations(ctx context.Context, t *testing.T, handler *Handler) {
	// Test neighbors query
	queryReq := `{"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "query_neighbors", "arguments": {"node": "user:alice", "direction": "out"}}}`
	response, err := handler.ProcessSingleRequest(ctx, queryReq)
	if err != nil {
		t.Fatalf("Failed to process query_neighbors request: %v", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}

	// Verify the response contains expected neighbor information
	responseStr := response
	if !strings.Contains(responseStr, "user:bob") || !strings.Contains(responseStr, "product:123") {
		t.Fatalf("Expected neighbors user:bob and product:123 in response: %s", responseStr)
	}

	// Test find query
	findReq := `{"jsonrpc": "2.0", "id": 8, "method": "tools/call", "params": {"name": "query_find", "arguments": {"type": "user"}}}`
	response, err = handler.ProcessSingleRequest(ctx, findReq)
	if err != nil {
		t.Fatalf("Failed to process query_find request: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}

	// Verify the response contains both users
	responseStr = response
	if !strings.Contains(responseStr, "user:alice") || !strings.Contains(responseStr, "user:bob") {
		t.Fatalf("Expected both users in find response: %s", responseStr)
	}

	// Test paths query
	pathsReq := `{"jsonrpc": "2.0", "id": 9, "method": "tools/call", "params": {"name": "query_paths", "arguments": {"from": "user:alice", "to": "product:123"}}}`
	response, err = handler.ProcessSingleRequest(ctx, pathsReq)
	if err != nil {
		t.Fatalf("Failed to process query_paths request: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}
}

func testDeleteOperations(ctx context.Context, t *testing.T, handler *Handler) {
	// Add a test edge to delete
	addEdgeReq := `{"jsonrpc": "2.0", "id": 10, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "user:bob", "to": "product:123", "label": "viewed"}}}`
	_, err := handler.ProcessSingleRequest(ctx, addEdgeReq)
	if err != nil {
		t.Fatalf("Failed to add test edge: %v", err)
	}

	// Test edge deletion
	deleteEdgeReq := `{"jsonrpc": "2.0", "id": 11, "method": "tools/call", "params": {"name": "delete_edge", "arguments": {"from": "user:bob", "to": "product:123", "label": "viewed"}}}`
	response, err := handler.ProcessSingleRequest(ctx, deleteEdgeReq)
	if err != nil {
		t.Fatalf("Failed to process delete_edge request: %v", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}

	// Test node deletion
	deleteNodeReq := `{"jsonrpc": "2.0", "id": 12, "method": "tools/call", "params": {"name": "delete_node", "arguments": {"id": "product:123"}}}`
	response, err = handler.ProcessSingleRequest(ctx, deleteNodeReq)
	if err != nil {
		t.Fatalf("Failed to process delete_node request: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}
}

func testErrorHandling(ctx context.Context, t *testing.T, handler *Handler) {
	// Test invalid tool name
	invalidToolReq := `{"jsonrpc": "2.0", "id": 13, "method": "tools/call", "params": {"name": "invalid_tool", "arguments": {}}}`
	response, err := handler.ProcessSingleRequest(ctx, invalidToolReq)
	if err != nil {
		t.Fatalf("Failed to process invalid tool request: %v", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check for tool result with error
	if result, ok := jsonResp.Result.(map[string]interface{}); ok {
		if isError, exists := result["isError"]; !exists || !isError.(bool) {
			t.Fatalf("Expected error response for invalid tool")
		}
	} else {
		t.Fatalf("Expected tool result in response")
	}

	// Test missing required arguments
	invalidArgsReq := `{"jsonrpc": "2.0", "id": 14, "method": "tools/call", "params": {"name": "add_node", "arguments": {}}}`
	response, err = handler.ProcessSingleRequest(ctx, invalidArgsReq)
	if err != nil {
		t.Fatalf("Failed to process invalid args request: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Check for tool result with error
	if result, ok := jsonResp.Result.(map[string]interface{}); ok {
		if isError, exists := result["isError"]; !exists || !isError.(bool) {
			t.Fatalf("Expected error response for missing arguments")
		}
	} else {
		t.Fatalf("Expected tool result in response")
	}
}

func testComplexScenarios(ctx context.Context, t *testing.T, handler *Handler) {
	// Create a more complex graph structure
	nodes := []string{"company:acme", "dept:engineering", "dept:sales", "user:charlie", "user:diana"}
	for i, nodeID := range nodes {
		reqID := fmt.Sprintf("%d", 15+i)
		addNodeReq := `{"jsonrpc": "2.0", "id": ` + reqID + `, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "` + nodeID + `", "type": "` + strings.Split(nodeID, ":")[0] + `"}}}`
		response, err := handler.ProcessSingleRequest(ctx, addNodeReq)
		if err != nil {
			t.Fatalf("Failed to add node %s: %v", nodeID, err)
		}

		var jsonResp JSONRPCResponse
		if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if jsonResp.Error != nil {
			t.Fatalf("Failed to add node %s: %s", nodeID, jsonResp.Error.Message)
		}
	}

	// Create relationships
	edges := []struct {
		from, to, label string
	}{
		{"company:acme", "dept:engineering", "has_department"},
		{"company:acme", "dept:sales", "has_department"},
		{"dept:engineering", "user:charlie", "employs"},
		{"dept:sales", "user:diana", "employs"},
		{"user:charlie", "user:diana", "colleagues"},
	}

	for i, edge := range edges {
		reqID := fmt.Sprintf("%d", 20+i)
		addEdgeReq := `{"jsonrpc": "2.0", "id": ` + reqID + `, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "` + edge.from + `", "to": "` + edge.to + `", "label": "` + edge.label + `"}}}`
		response, err := handler.ProcessSingleRequest(ctx, addEdgeReq)
		if err != nil {
			t.Fatalf("Failed to add edge %s -> %s: %v", edge.from, edge.to, err)
		}

		var jsonResp JSONRPCResponse
		if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if jsonResp.Error != nil {
			t.Fatalf("Failed to add edge %s -> %s: %s", edge.from, edge.to, jsonResp.Error.Message)
		}
	}

	// Test complex query: find path from company to user
	pathsReq := `{"jsonrpc": "2.0", "id": 25, "method": "tools/call", "params": {"name": "query_paths", "arguments": {"from": "company:acme", "to": "user:charlie", "max_depth": 3}}}`
	response, err := handler.ProcessSingleRequest(ctx, pathsReq)
	if err != nil {
		t.Fatalf("Failed to process complex paths query: %v", err)
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if jsonResp.Error != nil {
		t.Fatalf("Expected successful response, got error: %s", jsonResp.Error.Message)
	}

	// Verify path exists (company -> dept -> user)
	responseStr := response
	if !strings.Contains(responseStr, "company:acme") || !strings.Contains(responseStr, "user:charlie") {
		t.Fatalf("Expected path from company to user in response: %s", responseStr)
	}
}

// TestMCPCommandValidation tests command validation and error handling
func TestMCPCommandValidation(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()

	// Initialize server
	initReq := `{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}`
	_, err := handler.ProcessSingleRequest(ctx, initReq)
	if err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	testCases := []struct {
		name        string
		request     string
		expectError bool
		errorType   string
	}{
		{
			name:        "valid add_node",
			request:     `{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "add_node", "arguments": {"id": "test:valid"}}}`,
			expectError: false,
		},
		{
			name:        "add_node missing id",
			request:     `{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "add_node", "arguments": {}}}`,
			expectError: true,
			errorType:   "tool_error",
		},
		{
			name:        "add_edge missing required fields",
			request:     `{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "add_edge", "arguments": {"from": "test:valid"}}}`,
			expectError: true,
			errorType:   "tool_error",
		},
		{
			name:        "query_neighbors missing node",
			request:     `{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "query_neighbors", "arguments": {}}}`,
			expectError: true,
			errorType:   "tool_error",
		},
		{
			name:        "query_find no criteria",
			request:     `{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "query_find", "arguments": {}}}`,
			expectError: true,
			errorType:   "tool_error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := handler.ProcessSingleRequest(ctx, tc.request)
			if err != nil {
				t.Fatalf("Unexpected error processing request: %v", err)
			}

			var jsonResp JSONRPCResponse
			if err := json.Unmarshal([]byte(response), &jsonResp); err != nil {
				t.Fatalf("Failed to parse JSON response: %v", err)
			}

			if tc.expectError {
				// Check for tool result with error or JSON-RPC error
				if jsonResp.Error == nil {
					if result, ok := jsonResp.Result.(map[string]interface{}); ok {
						if isError, exists := result["isError"]; !exists || !isError.(bool) {
							t.Fatalf("Expected error response for test case: %s", tc.name)
						}
					} else {
						t.Fatalf("Expected error response for test case: %s", tc.name)
					}
				}
			} else {
				if jsonResp.Error != nil {
					t.Fatalf("Unexpected error for test case %s: %s", tc.name, jsonResp.Error.Message)
				}
			}
		})
	}
}
