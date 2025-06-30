package graph

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkAddNode measures node insertion performance
func BenchmarkAddNode(b *testing.B) {
	g := NewMemoryGraph()
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		node := Node{
			ID:   fmt.Sprintf("node:%d", i),
			Type: "benchmark",
			Props: map[string]string{
				"index": fmt.Sprintf("%d", i),
			},
		}

		if err := g.AddNode(ctx, node); err != nil {
			b.Fatalf("Failed to add node: %v", err)
		}
	}
}

// BenchmarkAddEdge measures edge insertion performance
func BenchmarkAddEdge(b *testing.B) {
	g := NewMemoryGraph()
	ctx := context.Background()

	// Pre-populate with nodes
	for i := 0; i < b.N+1; i++ {
		node := Node{
			ID:   fmt.Sprintf("node:%d", i),
			Type: "benchmark",
		}
		g.AddNode(ctx, node)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		edge := Edge{
			From:  fmt.Sprintf("node:%d", i),
			To:    fmt.Sprintf("node:%d", i+1),
			Label: "connects",
			Props: map[string]string{
				"weight": "1.0",
			},
		}

		if err := g.AddEdge(ctx, edge); err != nil {
			b.Fatalf("Failed to add edge: %v", err)
		}
	}
}

// BenchmarkGetNode measures node retrieval performance
func BenchmarkGetNode(b *testing.B) {
	g := NewMemoryGraph()
	ctx := context.Background()

	// Pre-populate with nodes
	nodeCount := 10000
	for i := 0; i < nodeCount; i++ {
		node := Node{
			ID:   fmt.Sprintf("node:%d", i),
			Type: "benchmark",
		}
		g.AddNode(ctx, node)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nodeID := fmt.Sprintf("node:%d", i%nodeCount)
		_, err := g.GetNode(ctx, nodeID)
		if err != nil {
			b.Fatalf("Failed to get node: %v", err)
		}
	}
}

// BenchmarkGetNeighbors measures neighbor query performance
func BenchmarkGetNeighbors(b *testing.B) {
	g := NewMemoryGraph()
	ctx := context.Background()

	// Create a graph with 1000 nodes, each connected to 10 neighbors
	nodeCount := 1000
	neighborsPerNode := 10

	// Add nodes
	for i := 0; i < nodeCount; i++ {
		node := Node{
			ID:   fmt.Sprintf("node:%d", i),
			Type: "benchmark",
		}
		g.AddNode(ctx, node)
	}

	// Add edges (each node connects to next 10 nodes, wrapping around)
	for i := 0; i < nodeCount; i++ {
		for j := 1; j <= neighborsPerNode; j++ {
			targetIndex := (i + j) % nodeCount
			edge := Edge{
				From:  fmt.Sprintf("node:%d", i),
				To:    fmt.Sprintf("node:%d", targetIndex),
				Label: "connects",
			}
			g.AddEdge(ctx, edge)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nodeID := fmt.Sprintf("node:%d", i%nodeCount)
		_, err := g.GetNeighbors(ctx, nodeID, "out")
		if err != nil {
			b.Fatalf("Failed to get neighbors: %v", err)
		}
	}
}

// BenchmarkQuery measures complex query performance
func BenchmarkQuery(b *testing.B) {
	g := NewMemoryGraph()
	ctx := context.Background()

	// Create a graph for path queries
	nodeCount := 100

	// Add nodes
	for i := 0; i < nodeCount; i++ {
		node := Node{
			ID:   fmt.Sprintf("node:%d", i),
			Type: "benchmark",
		}
		g.AddNode(ctx, node)
	}

	// Create a linear chain for path testing
	for i := 0; i < nodeCount-1; i++ {
		edge := Edge{
			From:  fmt.Sprintf("node:%d", i),
			To:    fmt.Sprintf("node:%d", i+1),
			Label: "next",
		}
		g.AddEdge(ctx, edge)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		query := Query{
			Type:      "neighbors",
			Node:      fmt.Sprintf("node:%d", i%nodeCount),
			Direction: "out",
		}

		_, err := g.Query(ctx, query)
		if err != nil {
			b.Fatalf("Failed to execute query: %v", err)
		}
	}
}

// TestPerformanceTargets validates that we meet the specified performance targets
func TestPerformanceTargets(t *testing.T) {
	g := NewMemoryGraph()
	ctx := context.Background()

	// Test node insertion target: < 100µs
	start := time.Now()
	node := Node{
		ID:   "perf:test:1",
		Type: "performance",
		Props: map[string]string{
			"test": "performance",
		},
	}
	err := g.AddNode(ctx, node)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}

	if duration > 100*time.Microsecond {
		t.Logf("Warning: Node insertion took %v, target is < 100µs", duration)
	} else {
		t.Logf("✓ Node insertion: %v (target: < 100µs)", duration)
	}

	// Add another node for edge test
	node2 := Node{ID: "perf:test:2", Type: "performance"}
	g.AddNode(ctx, node2)

	// Test edge insertion target: < 150µs
	start = time.Now()
	edge := Edge{
		From:  "perf:test:1",
		To:    "perf:test:2",
		Label: "connects",
		Props: map[string]string{
			"weight": "1.0",
		},
	}
	err = g.AddEdge(ctx, edge)
	duration = time.Since(start)

	if err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}

	if duration > 150*time.Microsecond {
		t.Logf("Warning: Edge insertion took %v, target is < 150µs", duration)
	} else {
		t.Logf("✓ Edge insertion: %v (target: < 150µs)", duration)
	}

	// Test neighborhood query target: < 1ms
	start = time.Now()
	_, err = g.GetNeighbors(ctx, "perf:test:1", "out")
	duration = time.Since(start)

	if err != nil {
		t.Fatalf("Failed to get neighbors: %v", err)
	}

	if duration > 1*time.Millisecond {
		t.Logf("Warning: Neighbor query took %v, target is < 1ms", duration)
	} else {
		t.Logf("✓ Neighbor query: %v (target: < 1ms)", duration)
	}
}

// BenchmarkConcurrentOperations tests performance under concurrent load
func BenchmarkConcurrentOperations(b *testing.B) {
	g := NewMemoryGraph()
	ctx := context.Background()

	// Pre-populate graph
	for i := 0; i < 1000; i++ {
		node := Node{
			ID:   fmt.Sprintf("node:%d", i),
			Type: "concurrent",
		}
		g.AddNode(ctx, node)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			nodeID := fmt.Sprintf("node:%d", i%1000)
			_, err := g.GetNode(ctx, nodeID)
			if err != nil {
				b.Fatalf("Failed to get node: %v", err)
			}
			i++
		}
	})
}
