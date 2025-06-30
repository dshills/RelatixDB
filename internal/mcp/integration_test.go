package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dshills/RelatixDB/internal/graph"
)

// TestMCPToolFunctions tests all MCP tool functions end-to-end
func TestMCPToolFunctions(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()

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
	addNodeCmd := `{"cmd": "add_node", "args": {"id": "user:alice", "type": "user", "props": {"name": "Alice", "email": "alice@example.com"}}}`
	response, err := handler.ProcessSingleCommand(ctx, addNodeCmd)
	if err != nil {
		t.Fatalf("Failed to process add_node command: %v", err)
	}

	var resp Response
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Verify response structure
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map")
	}

	if result["node_id"] != "user:alice" {
		t.Fatalf("Expected node_id 'user:alice', got %v", result["node_id"])
	}

	if result["action"] != "added" {
		t.Fatalf("Expected action 'added', got %v", result["action"])
	}

	// Test node without properties
	addNodeCmd2 := `{"cmd": "add_node", "args": {"id": "user:bob", "type": "user"}}`
	response, err = handler.ProcessSingleCommand(ctx, addNodeCmd2)
	if err != nil {
		t.Fatalf("Failed to process add_node command: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Test node without type
	addNodeCmd3 := `{"cmd": "add_node", "args": {"id": "item:123", "props": {"value": "test"}}}`
	response, err = handler.ProcessSingleCommand(ctx, addNodeCmd3)
	if err != nil {
		t.Fatalf("Failed to process add_node command: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}
}

func testAddEdgeOperations(ctx context.Context, t *testing.T, handler *Handler) {
	// Add more nodes for edge testing
	addNodeCmd := `{"cmd": "add_node", "args": {"id": "file:doc1", "type": "file", "props": {"name": "document1.txt"}}}`
	handler.ProcessSingleCommand(ctx, addNodeCmd)

	addNodeCmd2 := `{"cmd": "add_node", "args": {"id": "file:doc2", "type": "file", "props": {"name": "document2.txt"}}}`
	handler.ProcessSingleCommand(ctx, addNodeCmd2)

	// Test basic edge addition
	addEdgeCmd := `{"cmd": "add_edge", "args": {"from": "user:alice", "to": "file:doc1", "label": "owns", "props": {"since": "2024-01-01"}}}`
	response, err := handler.ProcessSingleCommand(ctx, addEdgeCmd)
	if err != nil {
		t.Fatalf("Failed to process add_edge command: %v", err)
	}

	var resp Response
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Verify response structure
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map")
	}

	if result["from"] != "user:alice" {
		t.Fatalf("Expected from 'user:alice', got %v", result["from"])
	}
	if result["to"] != "file:doc1" {
		t.Fatalf("Expected to 'file:doc1', got %v", result["to"])
	}
	if result["label"] != "owns" {
		t.Fatalf("Expected label 'owns', got %v", result["label"])
	}
	if result["action"] != "added" {
		t.Fatalf("Expected action 'added', got %v", result["action"])
	}

	// Test edge without properties
	addEdgeCmd2 := `{"cmd": "add_edge", "args": {"from": "user:bob", "to": "file:doc2", "label": "reads"}}`
	response, err = handler.ProcessSingleCommand(ctx, addEdgeCmd2)
	if err != nil {
		t.Fatalf("Failed to process add_edge command: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Test multiple edges between same nodes
	addEdgeCmd3 := `{"cmd": "add_edge", "args": {"from": "user:alice", "to": "file:doc1", "label": "edits"}}`
	response, err = handler.ProcessSingleCommand(ctx, addEdgeCmd3)
	if err != nil {
		t.Fatalf("Failed to process add_edge command: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}
}

func testQueryOperations(ctx context.Context, t *testing.T, handler *Handler) {
	// Test neighbors query - outgoing
	queryCmd := `{"cmd": "query", "args": {"type": "neighbors", "node": "user:alice", "direction": "out"}}`
	response, err := handler.ProcessSingleCommand(ctx, queryCmd)
	if err != nil {
		t.Fatalf("Failed to process neighbors query: %v", err)
	}

	var resp Response
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Verify query result structure
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map")
	}

	nodes, ok := result["nodes"].([]interface{})
	if !ok {
		t.Fatalf("Expected nodes to be an array")
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 neighbor, got %d", len(nodes))
	}

	// Test neighbors query - incoming
	queryCmd2 := `{"cmd": "query", "args": {"type": "neighbors", "node": "file:doc1", "direction": "in"}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd2)
	if err != nil {
		t.Fatalf("Failed to process incoming neighbors query: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result = resp.Result.(map[string]interface{})
	nodes = result["nodes"].([]interface{})
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 incoming neighbor, got %d", len(nodes))
	}

	// Test neighbors query - both directions
	queryCmd3 := `{"cmd": "query", "args": {"type": "neighbors", "node": "file:doc1", "direction": "both"}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd3)
	if err != nil {
		t.Fatalf("Failed to process both directions neighbors query: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Test find query by type
	queryCmd4 := `{"cmd": "query", "args": {"type": "find", "filters": {"type": "user"}}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd4)
	if err != nil {
		t.Fatalf("Failed to process find query: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result = resp.Result.(map[string]interface{})
	nodes = result["nodes"].([]interface{})
	if len(nodes) != 2 {
		t.Fatalf("Expected 2 user nodes, got %d", len(nodes))
	}

	// Test find query with multiple filters
	queryCmd5 := `{"cmd": "query", "args": {"type": "find", "filters": {"type": "user", "name": "Alice"}}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd5)
	if err != nil {
		t.Fatalf("Failed to process filtered find query: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Test paths query
	queryCmd6 := `{"cmd": "query", "args": {"type": "paths", "from": "user:alice", "to": "file:doc1", "max_depth": 3}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd6)
	if err != nil {
		t.Fatalf("Failed to process paths query: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result = resp.Result.(map[string]interface{})
	paths, ok := result["paths"].([]interface{})
	if !ok {
		t.Fatalf("Expected paths to be an array")
	}

	if len(paths) == 0 {
		t.Fatalf("Expected at least 1 path, got %d", len(paths))
	}
}

func testDeleteOperations(ctx context.Context, t *testing.T, handler *Handler) {
	// Test delete edge
	deleteEdgeCmd := `{"cmd": "delete_edge", "args": {"from": "user:alice", "to": "file:doc1", "label": "edits"}}`
	response, err := handler.ProcessSingleCommand(ctx, deleteEdgeCmd)
	if err != nil {
		t.Fatalf("Failed to process delete_edge command: %v", err)
	}

	var resp Response
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Verify response structure
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map")
	}

	if result["action"] != "deleted" {
		t.Fatalf("Expected action 'deleted', got %v", result["action"])
	}

	// Verify edge is actually deleted by checking neighbors
	queryCmd := `{"cmd": "query", "args": {"type": "neighbors", "node": "user:alice", "direction": "out"}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd)
	if err != nil {
		t.Fatalf("Failed to verify edge deletion: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	// Should still have the "owns" edge, but not "edits"
	result = resp.Result.(map[string]interface{})
	nodes := result["nodes"].([]interface{})
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 neighbor after edge deletion, got %d", len(nodes))
	}

	// Test delete node
	deleteNodeCmd := `{"cmd": "delete_node", "args": {"id": "item:123"}}`
	response, err = handler.ProcessSingleCommand(ctx, deleteNodeCmd)
	if err != nil {
		t.Fatalf("Failed to process delete_node command: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result = resp.Result.(map[string]interface{})
	if result["action"] != "deleted" {
		t.Fatalf("Expected action 'deleted', got %v", result["action"])
	}
}

func testErrorHandling(ctx context.Context, t *testing.T, handler *Handler) {
	// Test invalid JSON
	invalidJSON := `{"cmd": "add_node", "args":}`
	response, err := handler.ProcessSingleCommand(ctx, invalidJSON)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}

	var resp Response
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.OK {
		t.Fatalf("Expected error response for invalid JSON")
	}

	// Test missing required fields
	missingID := `{"cmd": "add_node", "args": {"type": "test"}}`
	response, err = handler.ProcessSingleCommand(ctx, missingID)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.OK {
		t.Fatalf("Expected error response for missing node ID")
	}

	// Test duplicate node
	duplicateNode := `{"cmd": "add_node", "args": {"id": "user:alice", "type": "user"}}`
	response, err = handler.ProcessSingleCommand(ctx, duplicateNode)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.OK {
		t.Fatalf("Expected error response for duplicate node")
	}

	// Test edge with non-existent node
	badEdge := `{"cmd": "add_edge", "args": {"from": "nonexistent", "to": "user:alice", "label": "connects"}}`
	response, err = handler.ProcessSingleCommand(ctx, badEdge)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.OK {
		t.Fatalf("Expected error response for edge with non-existent node")
	}

	// Test invalid query type
	invalidQuery := `{"cmd": "query", "args": {"type": "invalid_type", "node": "user:alice"}}`
	response, err = handler.ProcessSingleCommand(ctx, invalidQuery)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.OK {
		t.Fatalf("Expected error response for invalid query type")
	}

	// Test unknown command
	unknownCmd := `{"cmd": "unknown_command", "args": {}}`
	response, err = handler.ProcessSingleCommand(ctx, unknownCmd)
	if err != nil {
		t.Fatalf("Expected no error from handler, got %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.OK {
		t.Fatalf("Expected error response for unknown command")
	}
}

func testComplexScenarios(ctx context.Context, t *testing.T, handler *Handler) {
	// Create a more complex graph for testing
	commands := []string{
		`{"cmd": "add_node", "args": {"id": "project:webapp", "type": "project", "props": {"name": "Web Application", "status": "active"}}}`,
		`{"cmd": "add_node", "args": {"id": "task:frontend", "type": "task", "props": {"name": "Frontend Development", "priority": "high"}}}`,
		`{"cmd": "add_node", "args": {"id": "task:backend", "type": "task", "props": {"name": "Backend Development", "priority": "medium"}}}`,
		`{"cmd": "add_node", "args": {"id": "task:testing", "type": "task", "props": {"name": "Testing", "priority": "high"}}}`,
		`{"cmd": "add_edge", "args": {"from": "project:webapp", "to": "task:frontend", "label": "contains"}}`,
		`{"cmd": "add_edge", "args": {"from": "project:webapp", "to": "task:backend", "label": "contains"}}`,
		`{"cmd": "add_edge", "args": {"from": "project:webapp", "to": "task:testing", "label": "contains"}}`,
		`{"cmd": "add_edge", "args": {"from": "task:frontend", "to": "task:testing", "label": "blocks"}}`,
		`{"cmd": "add_edge", "args": {"from": "task:backend", "to": "task:testing", "label": "blocks"}}`,
		`{"cmd": "add_edge", "args": {"from": "user:alice", "to": "task:frontend", "label": "assigned"}}`,
		`{"cmd": "add_edge", "args": {"from": "user:bob", "to": "task:backend", "label": "assigned"}}`,
	}

	// Execute all setup commands
	for i, cmd := range commands {
		response, err := handler.ProcessSingleCommand(ctx, cmd)
		if err != nil {
			t.Fatalf("Failed to execute setup command %d: %v", i, err)
		}

		var resp Response
		if err := json.Unmarshal([]byte(response), &resp); err != nil {
			t.Fatalf("Failed to parse setup response %d: %v", i, err)
		}

		if !resp.OK {
			t.Fatalf("Setup command %d failed: %s", i, resp.Error)
		}
	}

	// Test complex queries

	// 1. Find all tasks in the project
	queryCmd := `{"cmd": "query", "args": {"type": "neighbors", "node": "project:webapp", "direction": "out"}}`
	response, err := handler.ProcessSingleCommand(ctx, queryCmd)
	if err != nil {
		t.Fatalf("Failed to query project tasks: %v", err)
	}

	var resp Response
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result := resp.Result.(map[string]interface{})
	nodes := result["nodes"].([]interface{})
	if len(nodes) != 3 {
		t.Fatalf("Expected 3 project tasks, got %d", len(nodes))
	}

	// 2. Find all high priority tasks
	queryCmd2 := `{"cmd": "query", "args": {"type": "find", "filters": {"type": "task", "priority": "high"}}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd2)
	if err != nil {
		t.Fatalf("Failed to query high priority tasks: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result = resp.Result.(map[string]interface{})
	nodes = result["nodes"].([]interface{})
	if len(nodes) != 2 {
		t.Fatalf("Expected 2 high priority tasks, got %d", len(nodes))
	}

	// 3. Find path from Alice to testing task
	queryCmd3 := `{"cmd": "query", "args": {"type": "paths", "from": "user:alice", "to": "task:testing", "max_depth": 5}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd3)
	if err != nil {
		t.Fatalf("Failed to query paths: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result = resp.Result.(map[string]interface{})
	paths := result["paths"].([]interface{})
	if len(paths) == 0 {
		t.Fatalf("Expected at least 1 path from Alice to testing task, got %d", len(paths))
	}

	// 4. Test query with default direction (should be "both")
	queryCmd4 := `{"cmd": "query", "args": {"type": "neighbors", "node": "task:testing"}}`
	response, err = handler.ProcessSingleCommand(ctx, queryCmd4)
	if err != nil {
		t.Fatalf("Failed to query with default direction: %v", err)
	}

	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.OK {
		t.Fatalf("Expected successful response, got error: %s", resp.Error)
	}

	result = resp.Result.(map[string]interface{})
	nodes = result["nodes"].([]interface{})
	// Should have neighbors from both directions: project:webapp (in) and task:frontend, task:backend (in via "blocks")
	if len(nodes) < 2 {
		t.Fatalf("Expected at least 2 neighbors with default direction, got %d", len(nodes))
	}
}

// TestMCPCommandValidation tests command and argument validation
func TestMCPCommandValidation(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()

	// Test various validation scenarios
	validationTests := []struct {
		name        string
		command     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty command",
			command:     `{"args": {"id": "test"}}`,
			expectError: true,
			errorMsg:    "command field is required",
		},
		{
			name:        "empty node ID",
			command:     `{"cmd": "add_node", "args": {"id": "", "type": "test"}}`,
			expectError: true,
			errorMsg:    "node ID is required",
		},
		{
			name:        "empty edge from",
			command:     `{"cmd": "add_edge", "args": {"from": "", "to": "node2", "label": "connects"}}`,
			expectError: true,
			errorMsg:    "from node is required",
		},
		{
			name:        "empty edge to",
			command:     `{"cmd": "add_edge", "args": {"from": "node1", "to": "", "label": "connects"}}`,
			expectError: true,
			errorMsg:    "to node is required",
		},
		{
			name:        "empty edge label",
			command:     `{"cmd": "add_edge", "args": {"from": "node1", "to": "node2", "label": ""}}`,
			expectError: true,
			errorMsg:    "edge label is required",
		},
		{
			name:        "invalid query direction",
			command:     `{"cmd": "query", "args": {"type": "neighbors", "node": "test", "direction": "invalid"}}`,
			expectError: true,
			errorMsg:    "invalid direction",
		},
		{
			name:        "neighbors query without node",
			command:     `{"cmd": "query", "args": {"type": "neighbors", "direction": "out"}}`,
			expectError: true,
			errorMsg:    "node is required for neighbors query",
		},
		{
			name:        "paths query without from",
			command:     `{"cmd": "query", "args": {"type": "paths", "to": "node2"}}`,
			expectError: true,
			errorMsg:    "from node is required for paths query",
		},
		{
			name:        "paths query without to",
			command:     `{"cmd": "query", "args": {"type": "paths", "from": "node1"}}`,
			expectError: true,
			errorMsg:    "to node is required for paths query",
		},
		{
			name:        "find query without filters",
			command:     `{"cmd": "query", "args": {"type": "find"}}`,
			expectError: true,
			errorMsg:    "filters are required for find query",
		},
	}

	for _, test := range validationTests {
		t.Run(test.name, func(t *testing.T) {
			response, err := handler.ProcessSingleCommand(ctx, test.command)
			if err != nil {
				t.Fatalf("Expected no error from handler, got %v", err)
			}

			var resp Response
			if err := json.Unmarshal([]byte(response), &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if test.expectError {
				if resp.OK {
					t.Fatalf("Expected error response, got success")
				}
				if !strings.Contains(resp.Error, test.errorMsg) {
					t.Fatalf("Expected error containing '%s', got '%s'", test.errorMsg, resp.Error)
				}
			} else {
				if !resp.OK {
					t.Fatalf("Expected successful response, got error: %s", resp.Error)
				}
			}
		})
	}
}
