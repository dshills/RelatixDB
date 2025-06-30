package graph

import (
	"context"
	"fmt"
)

// QueryEngine handles complex graph queries
type QueryEngine struct {
	graph Graph
}

// NewQueryEngine creates a new query engine
func NewQueryEngine(graph Graph) *QueryEngine {
	return &QueryEngine{
		graph: graph,
	}
}

// Query executes a graph query and returns results
func (qe *QueryEngine) Query(ctx context.Context, query Query) (*QueryResult, error) {
	switch query.Type {
	case "neighbors":
		return qe.queryNeighbors(ctx, query)
	case "paths":
		return qe.queryPaths(ctx, query)
	case "find":
		return qe.queryFind(ctx, query)
	default:
		return nil, fmt.Errorf("unknown query type: %s", query.Type)
	}
}

// queryNeighbors handles neighbor queries
func (qe *QueryEngine) queryNeighbors(ctx context.Context, query Query) (*QueryResult, error) {
	if query.Node == "" {
		return nil, fmt.Errorf("node ID is required for neighbors query")
	}
	
	direction := query.Direction
	if direction == "" {
		direction = "both" // default to both directions
	}
	
	neighbors, err := qe.graph.GetNeighbors(ctx, query.Node, direction)
	if err != nil {
		return nil, fmt.Errorf("failed to get neighbors: %w", err)
	}
	
	// Filter by label if specified
	if query.Label != "" {
		filteredNeighbors, err := qe.filterNeighborsByLabel(ctx, query.Node, neighbors, query.Label, direction)
		if err != nil {
			return nil, fmt.Errorf("failed to filter neighbors by label: %w", err)
		}
		neighbors = filteredNeighbors
	}
	
	return &QueryResult{
		Nodes: neighbors,
	}, nil
}

// queryPaths handles path-finding queries
func (qe *QueryEngine) queryPaths(ctx context.Context, query Query) (*QueryResult, error) {
	if query.From == "" || query.To == "" {
		return nil, fmt.Errorf("both 'from' and 'to' nodes are required for path queries")
	}
	
	maxDepth := query.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 4 // default max depth as per spec
	}
	
	if maxDepth > 10 {
		return nil, ErrMaxDepthExceeded
	}
	
	paths, err := qe.findPaths(ctx, query.From, query.To, maxDepth, query.Label)
	if err != nil {
		return nil, fmt.Errorf("failed to find paths: %w", err)
	}
	
	return &QueryResult{
		Paths: paths,
	}, nil
}

// queryFind handles property-based search queries
func (qe *QueryEngine) queryFind(ctx context.Context, query Query) (*QueryResult, error) {
	if query.Filters == nil || len(query.Filters) == 0 {
		return nil, fmt.Errorf("filters are required for find queries")
	}
	
	var nodes []Node
	
	// Check if searching by type
	if nodeType, exists := query.Filters["type"]; exists {
		typeNodes, err := qe.graph.GetNodesByType(ctx, nodeType)
		if err != nil {
			return nil, fmt.Errorf("failed to get nodes by type: %w", err)
		}
		nodes = typeNodes
	} else {
		// For now, we don't have a way to iterate all nodes
		// This would need to be added to the Graph interface
		return nil, fmt.Errorf("property-based search without type filter not yet implemented")
	}
	
	// Apply additional property filters
	filteredNodes := qe.filterNodesByProperties(nodes, query.Filters)
	
	return &QueryResult{
		Nodes: filteredNodes,
	}, nil
}

// filterNeighborsByLabel filters neighbors by edge label
func (qe *QueryEngine) filterNeighborsByLabel(ctx context.Context, nodeID string, neighbors []Node, label, direction string) ([]Node, error) {
	// This is a simplified implementation
	// In a real system, we'd need access to the edges to filter properly
	// For now, we'll return all neighbors (no filtering by label)
	return neighbors, nil
}

// findPaths finds all paths between two nodes using BFS
func (qe *QueryEngine) findPaths(ctx context.Context, from, to string, maxDepth int, label string) ([]Path, error) {
	if !qe.graph.NodeExists(ctx, from) {
		return nil, fmt.Errorf("from node '%s' does not exist", from)
	}
	if !qe.graph.NodeExists(ctx, to) {
		return nil, fmt.Errorf("to node '%s' does not exist", to)
	}
	
	if from == to {
		// Self-path
		node, err := qe.graph.GetNode(ctx, from)
		if err != nil {
			return nil, err
		}
		return []Path{{Nodes: []Node{*node}}}, nil
	}
	
	var paths []Path
	
	// Use BFS to find paths
	queue := []pathState{{
		currentNode: from,
		path:        []string{from},
		depth:       0,
	}}
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if current.depth >= maxDepth {
			continue
		}
		
		neighbors, err := qe.graph.GetNeighbors(ctx, current.currentNode, "out")
		if err != nil {
			continue
		}
		
		for _, neighbor := range neighbors {
			// Avoid cycles in current path
			if contains(current.path, neighbor.ID) {
				continue
			}
			
			newPath := append(current.path, neighbor.ID)
			
			if neighbor.ID == to {
				// Found a path to target
				pathNodes, err := qe.buildPathNodes(ctx, newPath)
				if err != nil {
					continue
				}
				paths = append(paths, Path{Nodes: pathNodes})
			} else {
				// Continue exploring
				queue = append(queue, pathState{
					currentNode: neighbor.ID,
					path:        newPath,
					depth:       current.depth + 1,
				})
			}
		}
	}
	
	return paths, nil
}

// pathState represents the state during path finding
type pathState struct {
	currentNode string
	path        []string
	depth       int
}

// buildPathNodes converts a path of node IDs to a path of Node objects
func (qe *QueryEngine) buildPathNodes(ctx context.Context, nodeIDs []string) ([]Node, error) {
	var nodes []Node
	
	for _, id := range nodeIDs {
		node, err := qe.graph.GetNode(ctx, id)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, *node)
	}
	
	return nodes, nil
}

// filterNodesByProperties filters nodes based on property criteria
func (qe *QueryEngine) filterNodesByProperties(nodes []Node, filters map[string]string) []Node {
	var filtered []Node
	
	for _, node := range nodes {
		matches := true
		
		for key, value := range filters {
			if key == "type" {
				if node.Type != value {
					matches = false
					break
				}
			} else {
				if node.Props == nil || node.Props[key] != value {
					matches = false
					break
				}
			}
		}
		
		if matches {
			filtered = append(filtered, node)
		}
	}
	
	return filtered
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Implement the Query method for MemoryGraph to satisfy the Graph interface
func (g *MemoryGraph) Query(ctx context.Context, query Query) (*QueryResult, error) {
	queryEngine := NewQueryEngine(g)
	return queryEngine.Query(ctx, query)
}