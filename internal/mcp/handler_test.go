package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/dshills/RelatixDB/internal/graph"
)

func TestHandler_ProcessSingleCommand(t *testing.T) {
	g := graph.NewMemoryGraph()
	handler := NewHandler(g, nil, nil, false)
	ctx := context.Background()
	
	// Test add_node command
	addNodeCmd := `{"cmd": "add_node", "args": {"id": "test:1", "type": "test", "props": {"name": "Test Node"}}}`
	
	response, err := handler.ProcessSingleCommand(ctx, addNodeCmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !strings.Contains(response, `"ok":true`) {
		t.Fatalf("Expected successful response, got %s", response)
	}
	
	// Test add_edge command (should fail - no target node)
	addEdgeCmd := `{"cmd": "add_edge", "args": {"from": "test:1", "to": "test:2", "label": "connects"}}`
	
	response, err = handler.ProcessSingleCommand(ctx, addEdgeCmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !strings.Contains(response, `"ok":false`) {
		t.Fatalf("Expected error response, got %s", response)
	}
	
	// Add target node and try again
	addNode2Cmd := `{"cmd": "add_node", "args": {"id": "test:2", "type": "test"}}`
	
	response, err = handler.ProcessSingleCommand(ctx, addNode2Cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	response, err = handler.ProcessSingleCommand(ctx, addEdgeCmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !strings.Contains(response, `"ok":true`) {
		t.Fatalf("Expected successful response, got %s", response)
	}
	
	// Test query command
	queryCmd := `{"cmd": "query", "args": {"type": "neighbors", "node": "test:1", "direction": "out"}}`
	
	response, err = handler.ProcessSingleCommand(ctx, queryCmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !strings.Contains(response, `"ok":true`) {
		t.Fatalf("Expected successful response, got %s", response)
	}
}

func TestParseCommand(t *testing.T) {
	// Test valid command
	cmdJSON := `{"cmd": "add_node", "args": {"id": "test:1"}}`
	
	cmd, err := ParseCommand([]byte(cmdJSON))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if cmd.Cmd != "add_node" {
		t.Fatalf("Expected cmd 'add_node', got %s", cmd.Cmd)
	}
	
	// Test invalid JSON
	invalidJSON := `{"cmd": "add_node", "args":}`
	
	_, err = ParseCommand([]byte(invalidJSON))
	if err == nil {
		t.Fatalf("Expected error for invalid JSON")
	}
	
	// Test missing cmd field
	missingCmd := `{"args": {"id": "test:1"}}`
	
	_, err = ParseCommand([]byte(missingCmd))
	if err == nil {
		t.Fatalf("Expected error for missing cmd field")
	}
}

func TestCommandArgsValidation(t *testing.T) {
	// Test AddNodeArgs validation
	validArgs := AddNodeArgs{ID: "test:1", Type: "test"}
	if err := validArgs.Validate(); err != nil {
		t.Fatalf("Expected no error for valid args, got %v", err)
	}
	
	invalidArgs := AddNodeArgs{ID: "", Type: "test"}
	if err := invalidArgs.Validate(); err == nil {
		t.Fatalf("Expected error for empty ID")
	}
	
	// Test AddEdgeArgs validation
	validEdgeArgs := AddEdgeArgs{From: "node1", To: "node2", Label: "connects"}
	if err := validEdgeArgs.Validate(); err != nil {
		t.Fatalf("Expected no error for valid edge args, got %v", err)
	}
	
	invalidEdgeArgs := AddEdgeArgs{From: "", To: "node2", Label: "connects"}
	if err := invalidEdgeArgs.Validate(); err == nil {
		t.Fatalf("Expected error for empty From field")
	}
	
	// Test QueryArgs validation
	validQueryArgs := QueryArgs{Type: "neighbors", Node: "test:1"}
	if err := validQueryArgs.Validate(); err != nil {
		t.Fatalf("Expected no error for valid query args, got %v", err)
	}
	
	invalidQueryArgs := QueryArgs{Type: "neighbors", Node: ""}
	if err := invalidQueryArgs.Validate(); err == nil {
		t.Fatalf("Expected error for empty Node in neighbors query")
	}
	
	pathQueryArgs := QueryArgs{Type: "paths", From: "node1", To: "node2"}
	if err := pathQueryArgs.Validate(); err != nil {
		t.Fatalf("Expected no error for valid path query args, got %v", err)
	}
	
	invalidPathArgs := QueryArgs{Type: "paths", From: "node1", To: ""}
	if err := invalidPathArgs.Validate(); err == nil {
		t.Fatalf("Expected error for empty To in path query")
	}
}