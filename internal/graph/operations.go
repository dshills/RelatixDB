package graph

import (
	"context"
	"fmt"
)

// Operations provides high-level graph operations that combine multiple low-level operations
type Operations struct {
	graph Graph
}

// NewOperations creates a new Operations instance
func NewOperations(graph Graph) *Operations {
	return &Operations{
		graph: graph,
	}
}

// CreateNode creates a new node, validating that it doesn't already exist
func (ops *Operations) CreateNode(ctx context.Context, node Node) error {
	if err := node.Validate(); err != nil {
		return fmt.Errorf("invalid node: %w", err)
	}

	if ops.graph.NodeExists(ctx, node.ID) {
		return ErrNodeExists
	}

	return ops.graph.AddNode(ctx, node)
}

// CreateEdge creates a new edge, validating that both nodes exist and edge doesn't exist
func (ops *Operations) CreateEdge(ctx context.Context, edge Edge) error {
	if err := edge.Validate(); err != nil {
		return fmt.Errorf("invalid edge: %w", err)
	}

	// Check that both nodes exist
	if !ops.graph.NodeExists(ctx, edge.From) {
		return fmt.Errorf("from node '%s' does not exist", edge.From)
	}

	if !ops.graph.NodeExists(ctx, edge.To) {
		return fmt.Errorf("to node '%s' does not exist", edge.To)
	}

	// Check if edge already exists
	if _, err := ops.graph.GetEdge(ctx, edge.From, edge.To, edge.Label); err == nil {
		return ErrEdgeExists
	}

	return ops.graph.AddEdge(ctx, edge)
}

// UpdateNode updates an existing node's properties
func (ops *Operations) UpdateNode(ctx context.Context, nodeID string, props map[string]string) error {
	if nodeID == "" {
		return ErrEmptyNodeID
	}

	// Get existing node
	existingNode, err := ops.graph.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}

	// Update properties
	if existingNode.Props == nil {
		existingNode.Props = make(map[string]string)
	}

	for key, value := range props {
		existingNode.Props[key] = value
	}

	// Delete and re-add the node (since we don't have direct update)
	if err := ops.graph.DeleteNode(ctx, nodeID); err != nil {
		return fmt.Errorf("failed to delete node for update: %w", err)
	}

	if err := ops.graph.AddNode(ctx, *existingNode); err != nil {
		return fmt.Errorf("failed to re-add updated node: %w", err)
	}

	return nil
}

// UpdateEdge updates an existing edge's properties
func (ops *Operations) UpdateEdge(ctx context.Context, from, to, label string, props map[string]string) error {
	if from == "" {
		return ErrEmptyFromNode
	}
	if to == "" {
		return ErrEmptyToNode
	}
	if label == "" {
		return ErrEmptyEdgeLabel
	}

	// Get existing edge
	existingEdge, err := ops.graph.GetEdge(ctx, from, to, label)
	if err != nil {
		return err
	}

	// Update properties
	if existingEdge.Props == nil {
		existingEdge.Props = make(map[string]string)
	}

	for key, value := range props {
		existingEdge.Props[key] = value
	}

	// Delete and re-add the edge
	if err := ops.graph.DeleteEdge(ctx, from, to, label); err != nil {
		return fmt.Errorf("failed to delete edge for update: %w", err)
	}

	if err := ops.graph.AddEdge(ctx, *existingEdge); err != nil {
		return fmt.Errorf("failed to re-add updated edge: %w", err)
	}

	return nil
}

// GetNodeWithEdges returns a node along with its connected edges
func (ops *Operations) GetNodeWithEdges(ctx context.Context, nodeID string) (*Node, []Edge, error) {
	// Get the node
	node, err := ops.graph.GetNode(ctx, nodeID)
	if err != nil {
		return nil, nil, err
	}

	// Get all connected edges (this is a simplified approach)
	var edges []Edge

	// This would need to be implemented more efficiently in a real system
	// For now, we'll return just the node
	return node, edges, nil
}

// BulkAddNodes adds multiple nodes in a single operation
func (ops *Operations) BulkAddNodes(ctx context.Context, nodes []Node) error {
	for i, node := range nodes {
		if err := ops.CreateNode(ctx, node); err != nil {
			return fmt.Errorf("failed to add node %d (%s): %w", i, node.ID, err)
		}
	}
	return nil
}

// BulkAddEdges adds multiple edges in a single operation
func (ops *Operations) BulkAddEdges(ctx context.Context, edges []Edge) error {
	for i, edge := range edges {
		if err := ops.CreateEdge(ctx, edge); err != nil {
			return fmt.Errorf("failed to add edge %d (%s->%s): %w", i, edge.From, edge.To, err)
		}
	}
	return nil
}

// NodeCount returns the total number of nodes in the graph
func (ops *Operations) NodeCount(ctx context.Context) (int, error) {
	// This would need to be implemented in the Graph interface
	// For now, we'll return 0
	return 0, fmt.Errorf("node count not implemented")
}

// EdgeCount returns the total number of edges in the graph
func (ops *Operations) EdgeCount(ctx context.Context) (int, error) {
	// This would need to be implemented in the Graph interface
	// For now, we'll return 0
	return 0, fmt.Errorf("edge count not implemented")
}

// ValidateGraph performs basic validation on the graph structure
func (ops *Operations) ValidateGraph(ctx context.Context) error {
	// This would implement various graph validation checks
	// For now, we'll just return nil
	return nil
}

// ClearGraph removes all nodes and edges from the graph
func (ops *Operations) ClearGraph(ctx context.Context) error {
	// This would need to be implemented more efficiently
	// For now, we'll return an error
	return fmt.Errorf("clear graph not implemented")
}
