package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dshills/RelatixDB/internal/graph"
)

// PersistentGraph wraps a memory graph with persistent storage
type PersistentGraph struct {
	memory  graph.Graph
	backend Backend
	
	// Auto-save configuration
	autoSave     bool
	saveInterval time.Duration
	stopChan     chan struct{}
	mu           sync.RWMutex
}

// NewPersistentGraph creates a new persistent graph
func NewPersistentGraph(backend Backend, autoSave bool, saveInterval time.Duration) *PersistentGraph {
	return &PersistentGraph{
		memory:       graph.NewMemoryGraph(),
		backend:      backend,
		autoSave:     autoSave,
		saveInterval: saveInterval,
		stopChan:     make(chan struct{}),
	}
}

// Load loads the graph from persistent storage
func (pg *PersistentGraph) Load(ctx context.Context) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	
	memGraph, err := pg.backend.LoadGraph(ctx)
	if err != nil {
		return fmt.Errorf("failed to load from storage: %w", err)
	}
	
	pg.memory = memGraph
	
	// Start auto-save if enabled
	if pg.autoSave {
		go pg.autoSaveLoop()
	}
	
	return nil
}

// Save saves the graph to persistent storage
func (pg *PersistentGraph) Save(ctx context.Context) error {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	
	return pg.backend.SaveGraph(ctx, pg.memory)
}

// Close closes the persistent graph and saves if needed
func (pg *PersistentGraph) Close() error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	
	// Stop auto-save
	if pg.autoSave {
		close(pg.stopChan)
	}
	
	// Skip final save for now since SaveGraph is not implemented
	// Individual operations are already persisted via transactions
	// TODO: Implement final save when SaveGraph is complete
	
	return pg.backend.Close()
}

// Auto-save loop runs in background
func (pg *PersistentGraph) autoSaveLoop() {
	ticker := time.NewTicker(pg.saveInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := pg.Save(ctx); err != nil {
				fmt.Printf("Auto-save failed: %v\n", err)
			}
		case <-pg.stopChan:
			return
		}
	}
}

// Graph interface implementation - delegate to memory graph

// AddNode adds a node to the graph
func (pg *PersistentGraph) AddNode(ctx context.Context, node graph.Node) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	
	// Add to memory first
	if err := pg.memory.AddNode(ctx, node); err != nil {
		return err
	}
	
	// Persist immediately for critical operations
	tx, err := pg.backend.BeginTransaction()
	if err != nil {
		// Rollback memory change
		pg.memory.DeleteNode(ctx, node.ID)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	if err := tx.SaveNode(node); err != nil {
		// Rollback memory change
		pg.memory.DeleteNode(ctx, node.ID)
		return fmt.Errorf("failed to persist node: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		// Rollback memory change
		pg.memory.DeleteNode(ctx, node.ID)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetNode retrieves a node by ID
func (pg *PersistentGraph) GetNode(ctx context.Context, id string) (*graph.Node, error) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	
	return pg.memory.GetNode(ctx, id)
}

// DeleteNode removes a node and all its connected edges
func (pg *PersistentGraph) DeleteNode(ctx context.Context, id string) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	
	// Get node first to check if it exists
	node, err := pg.memory.GetNode(ctx, id)
	if err != nil {
		return err
	}
	
	// Begin transaction
	tx, err := pg.backend.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Delete from memory first
	if err := pg.memory.DeleteNode(ctx, id); err != nil {
		return err
	}
	
	// Persist deletion
	if err := tx.DeleteNode(id); err != nil {
		// Rollback memory change
		pg.memory.AddNode(ctx, *node)
		return fmt.Errorf("failed to persist node deletion: %w", err)
	}
	
	// TODO: Also delete connected edges from storage
	// This requires extending the interface to get all edges for a node
	
	if err := tx.Commit(); err != nil {
		// Rollback memory change
		pg.memory.AddNode(ctx, *node)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// AddEdge adds an edge to the graph
func (pg *PersistentGraph) AddEdge(ctx context.Context, edge graph.Edge) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	
	// Add to memory first
	if err := pg.memory.AddEdge(ctx, edge); err != nil {
		return err
	}
	
	// Persist immediately
	tx, err := pg.backend.BeginTransaction()
	if err != nil {
		// Rollback memory change
		pg.memory.DeleteEdge(ctx, edge.From, edge.To, edge.Label)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	if err := tx.SaveEdge(edge); err != nil {
		// Rollback memory change
		pg.memory.DeleteEdge(ctx, edge.From, edge.To, edge.Label)
		return fmt.Errorf("failed to persist edge: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		// Rollback memory change
		pg.memory.DeleteEdge(ctx, edge.From, edge.To, edge.Label)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetEdge retrieves an edge by from, to, and label
func (pg *PersistentGraph) GetEdge(ctx context.Context, from, to, label string) (*graph.Edge, error) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	
	return pg.memory.GetEdge(ctx, from, to, label)
}

// DeleteEdge removes an edge from the graph
func (pg *PersistentGraph) DeleteEdge(ctx context.Context, from, to, label string) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	
	// Get edge first to check if it exists
	edge, err := pg.memory.GetEdge(ctx, from, to, label)
	if err != nil {
		return err
	}
	
	// Begin transaction
	tx, err := pg.backend.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Delete from memory first
	if err := pg.memory.DeleteEdge(ctx, from, to, label); err != nil {
		return err
	}
	
	// Persist deletion
	if err := tx.DeleteEdge(from, to, label); err != nil {
		// Rollback memory change
		pg.memory.AddEdge(ctx, *edge)
		return fmt.Errorf("failed to persist edge deletion: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		// Rollback memory change
		pg.memory.AddEdge(ctx, *edge)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// Query executes a graph query
func (pg *PersistentGraph) Query(ctx context.Context, query graph.Query) (*graph.QueryResult, error) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	
	return pg.memory.Query(ctx, query)
}

// NodeExists checks if a node exists in the graph
func (pg *PersistentGraph) NodeExists(ctx context.Context, id string) bool {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	
	return pg.memory.NodeExists(ctx, id)
}

// GetNodesByType returns all nodes of a specific type
func (pg *PersistentGraph) GetNodesByType(ctx context.Context, nodeType string) ([]graph.Node, error) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	
	return pg.memory.GetNodesByType(ctx, nodeType)
}

// GetNeighbors returns neighboring nodes in the specified direction
func (pg *PersistentGraph) GetNeighbors(ctx context.Context, nodeID, direction string) ([]graph.Node, error) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	
	return pg.memory.GetNeighbors(ctx, nodeID, direction)
}