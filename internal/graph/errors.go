package graph

import "errors"

// Error definitions for graph operations
var (
	// Node errors
	ErrEmptyNodeID  = errors.New("node ID cannot be empty")
	ErrNodeNotFound = errors.New("node not found")
	ErrNodeExists   = errors.New("node already exists")

	// Edge errors
	ErrEmptyFromNode  = errors.New("edge 'from' node cannot be empty")
	ErrEmptyToNode    = errors.New("edge 'to' node cannot be empty")
	ErrEmptyEdgeLabel = errors.New("edge label cannot be empty")
	ErrEdgeNotFound   = errors.New("edge not found")
	ErrEdgeExists     = errors.New("edge already exists")

	// Query errors
	ErrInvalidQuery     = errors.New("invalid query")
	ErrInvalidDirection = errors.New("invalid direction: must be 'in', 'out', or 'both'")
	ErrMaxDepthExceeded = errors.New("maximum query depth exceeded")

	// General errors
	ErrGraphClosed = errors.New("graph is closed")
)
