package graph

import (
	"context"
	"encoding/json"
)

// Node represents a graph node with unique ID, optional type, and properties
type Node struct {
	ID    string            `json:"id"`
	Type  string            `json:"type,omitempty"`
	Props map[string]string `json:"props,omitempty"`
}

// Edge represents a directed, labeled edge between two nodes
type Edge struct {
	From  string            `json:"from"`
	To    string            `json:"to"`
	Label string            `json:"label"`
	Props map[string]string `json:"props,omitempty"`
}

// Query represents a graph query with various parameters
type Query struct {
	Type      string            `json:"type"`      // "neighbors", "paths", "find"
	Node      string            `json:"node,omitempty"`
	Label     string            `json:"label,omitempty"`
	Direction string            `json:"direction,omitempty"` // "in", "out", "both"
	MaxDepth  int               `json:"max_depth,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
	From      string            `json:"from,omitempty"` // for path queries
	To        string            `json:"to,omitempty"`   // for path queries
}

// QueryResult represents the result of a graph query
type QueryResult struct {
	Nodes []Node `json:"nodes,omitempty"`
	Edges []Edge `json:"edges,omitempty"`
	Paths []Path `json:"paths,omitempty"`
}

// Path represents a path through the graph
type Path struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Graph defines the interface for graph operations
type Graph interface {
	// Node operations
	AddNode(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (*Node, error)
	DeleteNode(ctx context.Context, id string) error
	
	// Edge operations
	AddEdge(ctx context.Context, edge Edge) error
	GetEdge(ctx context.Context, from, to, label string) (*Edge, error)
	DeleteEdge(ctx context.Context, from, to, label string) error
	
	// Query operations
	Query(ctx context.Context, query Query) (*QueryResult, error)
	
	// Utility operations
	NodeExists(ctx context.Context, id string) bool
	GetNodesByType(ctx context.Context, nodeType string) ([]Node, error)
	GetNeighbors(ctx context.Context, nodeID, direction string) ([]Node, error)
}

// Validate checks if a Node is valid
func (n *Node) Validate() error {
	if n.ID == "" {
		return ErrEmptyNodeID
	}
	return nil
}

// Validate checks if an Edge is valid
func (e *Edge) Validate() error {
	if e.From == "" {
		return ErrEmptyFromNode
	}
	if e.To == "" {
		return ErrEmptyToNode
	}
	if e.Label == "" {
		return ErrEmptyEdgeLabel
	}
	return nil
}

// String returns a string representation of the Node
func (n *Node) String() string {
	data, _ := json.Marshal(n)
	return string(data)
}

// String returns a string representation of the Edge
func (e *Edge) String() string {
	data, _ := json.Marshal(e)
	return string(data)
}