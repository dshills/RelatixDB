package graph

import (
	"context"
	"testing"
)

func TestMemoryGraph_AddNode(t *testing.T) {
	g := NewMemoryGraph()
	ctx := context.Background()
	
	node := Node{
		ID:   "test:1",
		Type: "test",
		Props: map[string]string{
			"name": "Test Node",
		},
	}
	
	err := g.AddNode(ctx, node)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test duplicate node
	err = g.AddNode(ctx, node)
	if err != ErrNodeExists {
		t.Fatalf("Expected ErrNodeExists, got %v", err)
	}
}

func TestMemoryGraph_GetNode(t *testing.T) {
	g := NewMemoryGraph()
	ctx := context.Background()
	
	node := Node{
		ID:   "test:1",
		Type: "test",
		Props: map[string]string{
			"name": "Test Node",
		},
	}
	
	// Add node
	err := g.AddNode(ctx, node)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Get node
	retrieved, err := g.GetNode(ctx, "test:1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if retrieved.ID != node.ID {
		t.Fatalf("Expected ID %s, got %s", node.ID, retrieved.ID)
	}
	
	if retrieved.Type != node.Type {
		t.Fatalf("Expected Type %s, got %s", node.Type, retrieved.Type)
	}
	
	if retrieved.Props["name"] != node.Props["name"] {
		t.Fatalf("Expected Props name %s, got %s", node.Props["name"], retrieved.Props["name"])
	}
	
	// Test non-existent node
	_, err = g.GetNode(ctx, "nonexistent")
	if err != ErrNodeNotFound {
		t.Fatalf("Expected ErrNodeNotFound, got %v", err)
	}
}

func TestMemoryGraph_AddEdge(t *testing.T) {
	g := NewMemoryGraph()
	ctx := context.Background()
	
	// Add nodes first
	node1 := Node{ID: "node1", Type: "test"}
	node2 := Node{ID: "node2", Type: "test"}
	
	g.AddNode(ctx, node1)
	g.AddNode(ctx, node2)
	
	edge := Edge{
		From:  "node1",
		To:    "node2",
		Label: "connects",
		Props: map[string]string{
			"weight": "1.0",
		},
	}
	
	err := g.AddEdge(ctx, edge)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Test duplicate edge
	err = g.AddEdge(ctx, edge)
	if err != ErrEdgeExists {
		t.Fatalf("Expected ErrEdgeExists, got %v", err)
	}
	
	// Test edge with non-existent node
	badEdge := Edge{
		From:  "nonexistent",
		To:    "node2",
		Label: "connects",
	}
	
	err = g.AddEdge(ctx, badEdge)
	if err != ErrNodeNotFound {
		t.Fatalf("Expected ErrNodeNotFound, got %v", err)
	}
}

func TestMemoryGraph_GetNeighbors(t *testing.T) {
	g := NewMemoryGraph()
	ctx := context.Background()
	
	// Create a simple graph: node1 -> node2 -> node3
	nodes := []Node{
		{ID: "node1", Type: "test"},
		{ID: "node2", Type: "test"},
		{ID: "node3", Type: "test"},
	}
	
	for _, node := range nodes {
		g.AddNode(ctx, node)
	}
	
	edges := []Edge{
		{From: "node1", To: "node2", Label: "connects"},
		{From: "node2", To: "node3", Label: "connects"},
	}
	
	for _, edge := range edges {
		g.AddEdge(ctx, edge)
	}
	
	// Test outgoing neighbors
	neighbors, err := g.GetNeighbors(ctx, "node1", "out")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(neighbors) != 1 {
		t.Fatalf("Expected 1 neighbor, got %d", len(neighbors))
	}
	
	if neighbors[0].ID != "node2" {
		t.Fatalf("Expected neighbor node2, got %s", neighbors[0].ID)
	}
	
	// Test incoming neighbors
	neighbors, err = g.GetNeighbors(ctx, "node2", "in")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(neighbors) != 1 {
		t.Fatalf("Expected 1 neighbor, got %d", len(neighbors))
	}
	
	if neighbors[0].ID != "node1" {
		t.Fatalf("Expected neighbor node1, got %s", neighbors[0].ID)
	}
	
	// Test both directions
	neighbors, err = g.GetNeighbors(ctx, "node2", "both")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(neighbors) != 2 {
		t.Fatalf("Expected 2 neighbors, got %d", len(neighbors))
	}
}

func TestMemoryGraph_DeleteNode(t *testing.T) {
	g := NewMemoryGraph()
	ctx := context.Background()
	
	// Create nodes and edges
	nodes := []Node{
		{ID: "node1", Type: "test"},
		{ID: "node2", Type: "test"},
	}
	
	for _, node := range nodes {
		g.AddNode(ctx, node)
	}
	
	edge := Edge{From: "node1", To: "node2", Label: "connects"}
	g.AddEdge(ctx, edge)
	
	// Delete node1 (should also delete the edge)
	err := g.DeleteNode(ctx, "node1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify node is deleted
	_, err = g.GetNode(ctx, "node1")
	if err != ErrNodeNotFound {
		t.Fatalf("Expected ErrNodeNotFound, got %v", err)
	}
	
	// Verify edge is deleted
	_, err = g.GetEdge(ctx, "node1", "node2", "connects")
	if err != ErrEdgeNotFound {
		t.Fatalf("Expected ErrEdgeNotFound, got %v", err)
	}
	
	// Verify node2 still exists
	_, err = g.GetNode(ctx, "node2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMemoryGraph_GetNodesByType(t *testing.T) {
	g := NewMemoryGraph()
	ctx := context.Background()
	
	// Add nodes of different types
	nodes := []Node{
		{ID: "user1", Type: "user"},
		{ID: "user2", Type: "user"},
		{ID: "file1", Type: "file"},
	}
	
	for _, node := range nodes {
		g.AddNode(ctx, node)
	}
	
	// Get users
	users, err := g.GetNodesByType(ctx, "user")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}
	
	// Get files
	files, err := g.GetNodesByType(ctx, "file")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}
	
	// Get non-existent type
	empty, err := g.GetNodesByType(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(empty) != 0 {
		t.Fatalf("Expected 0 nodes, got %d", len(empty))
	}
}