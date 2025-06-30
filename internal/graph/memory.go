package graph

import (
	"context"
	"sync"
)

// MemoryGraph implements an in-memory graph with multiple indexes for fast access
type MemoryGraph struct {
	mu sync.RWMutex

	// Primary storage
	nodes map[string]*Node // node_id -> Node
	edges map[string]*Edge // edge_key -> Edge (key: from:to:label)

	// Indexes for fast lookups
	nodesByType map[string]map[string]*Node // type -> node_id -> Node
	outEdges    map[string]map[string]*Edge // from_node -> edge_key -> Edge
	inEdges     map[string]map[string]*Edge // to_node -> edge_key -> Edge

	closed bool
}

// NewMemoryGraph creates a new in-memory graph
func NewMemoryGraph() *MemoryGraph {
	return &MemoryGraph{
		nodes:       make(map[string]*Node),
		edges:       make(map[string]*Edge),
		nodesByType: make(map[string]map[string]*Node),
		outEdges:    make(map[string]map[string]*Edge),
		inEdges:     make(map[string]map[string]*Edge),
	}
}

// AddNode adds a node to the graph
func (g *MemoryGraph) AddNode(ctx context.Context, node Node) error {
	if err := node.Validate(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed {
		return ErrGraphClosed
	}

	// Check if node already exists
	if _, exists := g.nodes[node.ID]; exists {
		return ErrNodeExists
	}

	// Add to primary storage
	nodeCopy := node
	if nodeCopy.Props == nil {
		nodeCopy.Props = make(map[string]string)
	}
	g.nodes[node.ID] = &nodeCopy

	// Update type index
	if node.Type != "" {
		if g.nodesByType[node.Type] == nil {
			g.nodesByType[node.Type] = make(map[string]*Node)
		}
		g.nodesByType[node.Type][node.ID] = &nodeCopy
	}

	return nil
}

// GetNode retrieves a node by ID
func (g *MemoryGraph) GetNode(ctx context.Context, id string) (*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.closed {
		return nil, ErrGraphClosed
	}

	node, exists := g.nodes[id]
	if !exists {
		return nil, ErrNodeNotFound
	}

	// Return a copy to prevent external modification
	nodeCopy := *node
	return &nodeCopy, nil
}

// DeleteNode removes a node and all its connected edges
func (g *MemoryGraph) DeleteNode(ctx context.Context, id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed {
		return ErrGraphClosed
	}

	node, exists := g.nodes[id]
	if !exists {
		return ErrNodeNotFound
	}

	// Remove all outgoing edges
	if outgoing, exists := g.outEdges[id]; exists {
		for edgeKey := range outgoing {
			delete(g.edges, edgeKey)
		}
		delete(g.outEdges, id)
	}

	// Remove all incoming edges
	if incoming, exists := g.inEdges[id]; exists {
		for edgeKey := range incoming {
			edge := incoming[edgeKey]
			delete(g.edges, edgeKey)
			// Remove from outEdges of the source node
			if outgoing, exists := g.outEdges[edge.From]; exists {
				delete(outgoing, edgeKey)
			}
		}
		delete(g.inEdges, id)
	}

	// Remove from type index
	if node.Type != "" {
		if typeNodes, exists := g.nodesByType[node.Type]; exists {
			delete(typeNodes, id)
			if len(typeNodes) == 0 {
				delete(g.nodesByType, node.Type)
			}
		}
	}

	// Remove from primary storage
	delete(g.nodes, id)

	return nil
}

// AddEdge adds an edge to the graph
func (g *MemoryGraph) AddEdge(ctx context.Context, edge Edge) error {
	if err := edge.Validate(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed {
		return ErrGraphClosed
	}

	// Check that both nodes exist
	if _, exists := g.nodes[edge.From]; !exists {
		return ErrNodeNotFound
	}
	if _, exists := g.nodes[edge.To]; !exists {
		return ErrNodeNotFound
	}

	edgeKey := g.makeEdgeKey(edge.From, edge.To, edge.Label)

	// Check if edge already exists
	if _, exists := g.edges[edgeKey]; exists {
		return ErrEdgeExists
	}

	// Add to primary storage
	edgeCopy := edge
	if edgeCopy.Props == nil {
		edgeCopy.Props = make(map[string]string)
	}
	g.edges[edgeKey] = &edgeCopy

	// Update outEdges index
	if g.outEdges[edge.From] == nil {
		g.outEdges[edge.From] = make(map[string]*Edge)
	}
	g.outEdges[edge.From][edgeKey] = &edgeCopy

	// Update inEdges index
	if g.inEdges[edge.To] == nil {
		g.inEdges[edge.To] = make(map[string]*Edge)
	}
	g.inEdges[edge.To][edgeKey] = &edgeCopy

	return nil
}

// GetEdge retrieves an edge by from, to, and label
func (g *MemoryGraph) GetEdge(ctx context.Context, from, to, label string) (*Edge, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.closed {
		return nil, ErrGraphClosed
	}

	edgeKey := g.makeEdgeKey(from, to, label)
	edge, exists := g.edges[edgeKey]
	if !exists {
		return nil, ErrEdgeNotFound
	}

	// Return a copy to prevent external modification
	edgeCopy := *edge
	return &edgeCopy, nil
}

// DeleteEdge removes an edge from the graph
func (g *MemoryGraph) DeleteEdge(ctx context.Context, from, to, label string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed {
		return ErrGraphClosed
	}

	edgeKey := g.makeEdgeKey(from, to, label)

	// Check if edge exists
	if _, exists := g.edges[edgeKey]; !exists {
		return ErrEdgeNotFound
	}

	// Remove from primary storage
	delete(g.edges, edgeKey)

	// Remove from outEdges index
	if outgoing, exists := g.outEdges[from]; exists {
		delete(outgoing, edgeKey)
		if len(outgoing) == 0 {
			delete(g.outEdges, from)
		}
	}

	// Remove from inEdges index
	if incoming, exists := g.inEdges[to]; exists {
		delete(incoming, edgeKey)
		if len(incoming) == 0 {
			delete(g.inEdges, to)
		}
	}

	return nil
}

// NodeExists checks if a node exists in the graph
func (g *MemoryGraph) NodeExists(ctx context.Context, id string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.closed {
		return false
	}

	_, exists := g.nodes[id]
	return exists
}

// GetNodesByType returns all nodes of a specific type
func (g *MemoryGraph) GetNodesByType(ctx context.Context, nodeType string) ([]Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.closed {
		return nil, ErrGraphClosed
	}

	typeNodes, exists := g.nodesByType[nodeType]
	if !exists {
		return []Node{}, nil
	}

	nodes := make([]Node, 0, len(typeNodes))
	for _, node := range typeNodes {
		nodes = append(nodes, *node)
	}

	return nodes, nil
}

// GetNeighbors returns neighboring nodes in the specified direction
func (g *MemoryGraph) GetNeighbors(ctx context.Context, nodeID, direction string) ([]Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.closed {
		return nil, ErrGraphClosed
	}

	if !g.NodeExists(ctx, nodeID) {
		return nil, ErrNodeNotFound
	}

	neighbors := make(map[string]*Node)

	switch direction {
	case "out":
		if outgoing, exists := g.outEdges[nodeID]; exists {
			for _, edge := range outgoing {
				if node, exists := g.nodes[edge.To]; exists {
					neighbors[edge.To] = node
				}
			}
		}
	case "in":
		if incoming, exists := g.inEdges[nodeID]; exists {
			for _, edge := range incoming {
				if node, exists := g.nodes[edge.From]; exists {
					neighbors[edge.From] = node
				}
			}
		}
	case "both":
		// Get outgoing neighbors
		if outgoing, exists := g.outEdges[nodeID]; exists {
			for _, edge := range outgoing {
				if node, exists := g.nodes[edge.To]; exists {
					neighbors[edge.To] = node
				}
			}
		}
		// Get incoming neighbors
		if incoming, exists := g.inEdges[nodeID]; exists {
			for _, edge := range incoming {
				if node, exists := g.nodes[edge.From]; exists {
					neighbors[edge.From] = node
				}
			}
		}
	default:
		return nil, ErrInvalidDirection
	}

	result := make([]Node, 0, len(neighbors))
	for _, node := range neighbors {
		result = append(result, *node)
	}

	return result, nil
}

// Close closes the graph and prevents further operations
func (g *MemoryGraph) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.closed = true
	return nil
}

// makeEdgeKey creates a unique key for an edge
func (g *MemoryGraph) makeEdgeKey(from, to, label string) string {
	return from + ":" + to + ":" + label
}

// Stats returns statistics about the graph
func (g *MemoryGraph) Stats() map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return map[string]int{
		"nodes": len(g.nodes),
		"edges": len(g.edges),
		"types": len(g.nodesByType),
	}
}
