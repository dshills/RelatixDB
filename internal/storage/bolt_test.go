package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/RelatixDB/internal/graph"
)

func TestBoltBackend_OpenClose(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	backend := NewBoltBackend()

	// Test open
	err := backend.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Test close
	err = backend.Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("Database file was not created")
	}
}

func TestBoltBackend_Transaction(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	backend := NewBoltBackend()
	err := backend.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer backend.Close()

	// Test transaction operations
	tx, err := backend.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Test node operations
	node := graph.Node{
		ID:   "test:1",
		Type: "test",
		Props: map[string]string{
			"name": "Test Node",
		},
	}

	err = tx.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}

	// Test edge operations
	edge := graph.Edge{
		From:  "test:1",
		To:    "test:2",
		Label: "connects",
		Props: map[string]string{
			"weight": "1.0",
		},
	}

	err = tx.SaveEdge(edge)
	if err != nil {
		t.Fatalf("Failed to save edge: %v", err)
	}

	// Test commit
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
}

func TestBoltBackend_LoadGraph(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	backend := NewBoltBackend()
	err := backend.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer backend.Close()

	// Save some data first
	tx, err := backend.BeginTransaction()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	node := graph.Node{
		ID:   "test:1",
		Type: "test",
		Props: map[string]string{
			"name": "Test Node",
		},
	}

	err = tx.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Now test loading
	ctx := context.Background()
	loadedGraph, err := backend.LoadGraph(ctx)
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Verify the node was loaded
	retrievedNode, err := loadedGraph.GetNode(ctx, "test:1")
	if err != nil {
		t.Fatalf("Failed to get loaded node: %v", err)
	}

	if retrievedNode.ID != node.ID {
		t.Fatalf("Expected node ID %s, got %s", node.ID, retrievedNode.ID)
	}

	if retrievedNode.Type != node.Type {
		t.Fatalf("Expected node type %s, got %s", node.Type, retrievedNode.Type)
	}

	if retrievedNode.Props["name"] != node.Props["name"] {
		t.Fatalf("Expected node name %s, got %s", node.Props["name"], retrievedNode.Props["name"])
	}
}

func TestJSONSerializer(t *testing.T) {
	serializer := &JSONSerializer{}

	// Test node serialization
	node := graph.Node{
		ID:   "test:1",
		Type: "test",
		Props: map[string]string{
			"name": "Test Node",
		},
	}

	data, err := serializer.SerializeNode(node)
	if err != nil {
		t.Fatalf("Failed to serialize node: %v", err)
	}

	deserializedNode, err := serializer.DeserializeNode(data)
	if err != nil {
		t.Fatalf("Failed to deserialize node: %v", err)
	}

	if deserializedNode.ID != node.ID {
		t.Fatalf("Expected ID %s, got %s", node.ID, deserializedNode.ID)
	}

	// Test edge serialization
	edge := graph.Edge{
		From:  "test:1",
		To:    "test:2",
		Label: "connects",
		Props: map[string]string{
			"weight": "1.0",
		},
	}

	edgeData, err := serializer.SerializeEdge(edge)
	if err != nil {
		t.Fatalf("Failed to serialize edge: %v", err)
	}

	deserializedEdge, err := serializer.DeserializeEdge(edgeData)
	if err != nil {
		t.Fatalf("Failed to deserialize edge: %v", err)
	}

	if deserializedEdge.From != edge.From {
		t.Fatalf("Expected From %s, got %s", edge.From, deserializedEdge.From)
	}
}
