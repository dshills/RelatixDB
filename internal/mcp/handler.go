package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dshills/RelatixDB/internal/graph"
)

// Handler manages MCP protocol communication via stdio
type Handler struct {
	graph       graph.Graph
	reader      *bufio.Scanner
	writer      io.Writer
	debug       bool
	initialized bool
}

// NewHandler creates a new MCP handler
func NewHandler(g graph.Graph, reader io.Reader, writer io.Writer, debug bool) *Handler {
	return &Handler{
		graph:  g,
		reader: bufio.NewScanner(reader),
		writer: writer,
		debug:  debug,
	}
}

// NewStdioHandler creates a handler that uses stdin/stdout
func NewStdioHandler(g graph.Graph, debug bool) *Handler {
	return NewHandler(g, os.Stdin, os.Stdout, debug)
}

// Run starts the MCP handler loop, processing JSON-RPC requests from stdin
func (h *Handler) Run(ctx context.Context) error {
	h.debugLog("Starting MCP server...")

	for h.reader.Scan() {
		select {
		case <-ctx.Done():
			h.debugLog("Context canceled, stopping handler")
			return ctx.Err()
		default:
		}

		line := h.reader.Text()
		if line == "" {
			continue
		}

		h.debugLog("Received request: %s", line)

		response := h.processRequest(ctx, line)

		if err := h.writeResponse(response); err != nil {
			h.debugLog("Failed to write response: %v", err)
			return fmt.Errorf("failed to write response: %w", err)
		}
	}

	if err := h.reader.Err(); err != nil {
		h.debugLog("Scanner error: %v", err)
		return fmt.Errorf("scanner error: %w", err)
	}

	h.debugLog("MCP handler finished")
	return nil
}

// processRequest processes a single JSON-RPC request and returns a response
func (h *Handler) processRequest(ctx context.Context, requestLine string) *JSONRPCResponse {
	// Parse the JSON-RPC request
	req, err := ParseJSONRPCRequest([]byte(requestLine))
	if err != nil {
		h.debugLog("Failed to parse JSON-RPC request: %v", err)
		return NewJSONRPCErrorResponse(nil, ParseError, "Parse error", err.Error())
	}

	// Handle the request based on method
	switch req.Method {
	case MethodInitialize:
		return h.handleInitialize(ctx, req)
	case MethodToolsList:
		return h.handleToolsList(ctx, req)
	case MethodToolsCall:
		return h.handleToolsCall(ctx, req)
	default:
		h.debugLog("Unknown method: %s", req.Method)
		return NewJSONRPCErrorResponse(req.ID, MethodNotFound, "Method not found", req.Method)
	}
}

// handleInitialize handles the initialize request
func (h *Handler) handleInitialize(_ context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var initReq InitializeRequest
	if req.Params != nil {
		paramsData, err := json.Marshal(req.Params)
		if err != nil {
			return NewJSONRPCErrorResponse(req.ID, InvalidParams, "Invalid params", err.Error())
		}
		if err := json.Unmarshal(paramsData, &initReq); err != nil {
			return NewJSONRPCErrorResponse(req.ID, InvalidParams, "Invalid initialize params", err.Error())
		}
	}

	h.debugLog("Initialize request: protocol=%s, client=%s %s",
		initReq.ProtocolVersion, initReq.ClientInfo.Name, initReq.ClientInfo.Version)

	// Create response
	response := InitializeResponse{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "RelatixDB",
			Version: "1.0.0",
		},
	}

	h.initialized = true
	h.debugLog("Server initialized successfully")

	return NewJSONRPCResponse(req.ID, response)
}

// handleToolsList handles the tools/list request
func (h *Handler) handleToolsList(_ context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	if !h.initialized {
		return NewJSONRPCErrorResponse(req.ID, InvalidRequest, "Server not initialized", nil)
	}

	h.debugLog("Listing available tools")

	// Define RelatixDB MCP tools
	tools := []Tool{
		{
			Name:        "add_node",
			Description: "Add a node to the graph with ID, optional type, and properties",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier for the node",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Optional type of the node (e.g., 'file', 'function', 'module')",
					},
					"props": map[string]interface{}{
						"type":        "object",
						"description": "Optional key/value properties for the node",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "add_edge",
			Description: "Add a directed, labeled edge between two nodes",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Source node ID",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "Target node ID",
					},
					"label": map[string]interface{}{
						"type":        "string",
						"description": "Edge label/relationship type",
					},
					"props": map[string]interface{}{
						"type":        "object",
						"description": "Optional key/value properties for the edge",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
				Required: []string{"from", "to", "label"},
			},
		},
		{
			Name:        "delete_node",
			Description: "Delete a node from the graph (and all connected edges)",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the node to delete",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "delete_edge",
			Description: "Delete a specific edge from the graph",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Source node ID",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "Target node ID",
					},
					"label": map[string]interface{}{
						"type":        "string",
						"description": "Edge label/relationship type",
					},
				},
				Required: []string{"from", "to", "label"},
			},
		},
		{
			Name:        "query_neighbors",
			Description: "Find neighboring nodes connected to a specific node",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"node": map[string]interface{}{
						"type":        "string",
						"description": "Node ID to find neighbors for",
					},
					"direction": map[string]interface{}{
						"type":        "string",
						"description": "Direction of edges to follow: 'in', 'out', or 'both'",
						"enum":        []string{"in", "out", "both"},
					},
					"label": map[string]interface{}{
						"type":        "string",
						"description": "Optional edge label filter",
					},
				},
				Required: []string{"node"},
			},
		},
		{
			Name:        "query_paths",
			Description: "Find paths between two nodes in the graph",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"from": map[string]interface{}{
						"type":        "string",
						"description": "Starting node ID",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"description": "Target node ID",
					},
					"max_depth": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum path depth to search (default: 4)",
						"minimum":     1,
						"maximum":     10,
					},
				},
				Required: []string{"from", "to"},
			},
		},
		{
			Name:        "query_find",
			Description: "Find nodes matching specific criteria (type and/or properties)",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Node type to search for",
					},
					"props": map[string]interface{}{
						"type":        "object",
						"description": "Key/value properties to match",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
	}

	response := ListToolsResponse{
		Tools: tools,
	}

	h.debugLog("Returning %d tools", len(tools))
	return NewJSONRPCResponse(req.ID, response)
}

// handleToolsCall handles the tools/call request
func (h *Handler) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	if !h.initialized {
		return NewJSONRPCErrorResponse(req.ID, InvalidRequest, "Server not initialized", nil)
	}

	var callReq CallToolRequest
	if req.Params != nil {
		paramsData, err := json.Marshal(req.Params)
		if err != nil {
			return NewJSONRPCErrorResponse(req.ID, InvalidParams, "Invalid params", err.Error())
		}
		if err := json.Unmarshal(paramsData, &callReq); err != nil {
			return NewJSONRPCErrorResponse(req.ID, InvalidParams, "Invalid call tool params", err.Error())
		}
	}

	h.debugLog("Calling tool: %s with args: %v", callReq.Name, callReq.Arguments)

	// Execute the tool
	result, err := h.executeTool(ctx, callReq.Name, callReq.Arguments)
	if err != nil {
		h.debugLog("Tool execution failed: %v", err)
		return NewJSONRPCResponse(req.ID, CallToolResponse{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("Error: %s", err.Error()),
				},
			},
			IsError: true,
		})
	}

	h.debugLog("Tool executed successfully")
	return NewJSONRPCResponse(req.ID, result)
}

// executeTool executes a specific tool with given arguments
func (h *Handler) executeTool(ctx context.Context, toolName string, args map[string]interface{}) (*CallToolResponse, error) {
	switch toolName {
	case "add_node":
		return h.executeAddNode(ctx, args)
	case "add_edge":
		return h.executeAddEdge(ctx, args)
	case "delete_node":
		return h.executeDeleteNode(ctx, args)
	case "delete_edge":
		return h.executeDeleteEdge(ctx, args)
	case "query_neighbors":
		return h.executeQueryNeighbors(ctx, args)
	case "query_paths":
		return h.executeQueryPaths(ctx, args)
	case "query_find":
		return h.executeQueryFind(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// executeAddNode executes the add_node tool
func (h *Handler) executeAddNode(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	nodeType, _ := args["type"].(string)

	props := make(map[string]string)
	if propsRaw, ok := args["props"].(map[string]interface{}); ok {
		for k, v := range propsRaw {
			if strVal, ok := v.(string); ok {
				props[k] = strVal
			}
		}
	}

	node := graph.Node{
		ID:    id,
		Type:  nodeType,
		Props: props,
	}

	if err := h.graph.AddNode(ctx, node); err != nil {
		return nil, err
	}

	return &CallToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully added node '%s' with type '%s'", id, nodeType),
			},
		},
	}, nil
}

// executeAddEdge executes the add_edge tool
func (h *Handler) executeAddEdge(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	from, ok := args["from"].(string)
	if !ok || from == "" {
		return nil, fmt.Errorf("from is required and must be a string")
	}

	to, ok := args["to"].(string)
	if !ok || to == "" {
		return nil, fmt.Errorf("to is required and must be a string")
	}

	label, ok := args["label"].(string)
	if !ok || label == "" {
		return nil, fmt.Errorf("label is required and must be a string")
	}

	props := make(map[string]string)
	if propsRaw, ok := args["props"].(map[string]interface{}); ok {
		for k, v := range propsRaw {
			if strVal, ok := v.(string); ok {
				props[k] = strVal
			}
		}
	}

	edge := graph.Edge{
		From:  from,
		To:    to,
		Label: label,
		Props: props,
	}

	if err := h.graph.AddEdge(ctx, edge); err != nil {
		return nil, err
	}

	return &CallToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully added edge '%s' -> '%s' with label '%s'", from, to, label),
			},
		},
	}, nil
}

// executeDeleteNode executes the delete_node tool
func (h *Handler) executeDeleteNode(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	if err := h.graph.DeleteNode(ctx, id); err != nil {
		return nil, err
	}

	return &CallToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully deleted node '%s'", id),
			},
		},
	}, nil
}

// executeDeleteEdge executes the delete_edge tool
func (h *Handler) executeDeleteEdge(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	from, ok := args["from"].(string)
	if !ok || from == "" {
		return nil, fmt.Errorf("from is required and must be a string")
	}

	to, ok := args["to"].(string)
	if !ok || to == "" {
		return nil, fmt.Errorf("to is required and must be a string")
	}

	label, ok := args["label"].(string)
	if !ok || label == "" {
		return nil, fmt.Errorf("label is required and must be a string")
	}

	if err := h.graph.DeleteEdge(ctx, from, to, label); err != nil {
		return nil, err
	}

	return &CallToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully deleted edge '%s' -> '%s' with label '%s'", from, to, label),
			},
		},
	}, nil
}

// executeQueryNeighbors executes the query_neighbors tool
func (h *Handler) executeQueryNeighbors(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	node, ok := args["node"].(string)
	if !ok || node == "" {
		return nil, fmt.Errorf("node is required and must be a string")
	}

	direction, _ := args["direction"].(string)
	if direction == "" {
		direction = "both"
	}

	label, _ := args["label"].(string)

	query := graph.Query{
		Type:      "neighbors",
		Node:      node,
		Direction: direction,
		Label:     label,
	}

	result, err := h.graph.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Format the result
	resultText := fmt.Sprintf("Found %d neighbors for node '%s':\n", len(result.Nodes), node)
	for _, n := range result.Nodes {
		resultText += fmt.Sprintf("- %s (type: %s)\n", n.ID, n.Type)
	}

	if len(result.Edges) > 0 {
		resultText += fmt.Sprintf("\nConnecting edges (%d):\n", len(result.Edges))
		for _, e := range result.Edges {
			resultText += fmt.Sprintf("- %s -> %s (%s)\n", e.From, e.To, e.Label)
		}
	}

	return &CallToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// executeQueryPaths executes the query_paths tool
func (h *Handler) executeQueryPaths(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	from, ok := args["from"].(string)
	if !ok || from == "" {
		return nil, fmt.Errorf("from is required and must be a string")
	}

	to, ok := args["to"].(string)
	if !ok || to == "" {
		return nil, fmt.Errorf("to is required and must be a string")
	}

	maxDepth := 4
	if maxDepthRaw, ok := args["max_depth"]; ok {
		if maxDepthFloat, ok := maxDepthRaw.(float64); ok {
			maxDepth = int(maxDepthFloat)
		}
	}

	query := graph.Query{
		Type:     "paths",
		From:     from,
		To:       to,
		MaxDepth: maxDepth,
	}

	result, err := h.graph.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Format the result
	resultText := fmt.Sprintf("Found %d paths from '%s' to '%s':\n", len(result.Paths), from, to)
	for i, path := range result.Paths {
		resultText += fmt.Sprintf("Path %d: ", i+1)
		for j, node := range path.Nodes {
			if j > 0 {
				resultText += " -> "
			}
			resultText += node.ID
		}
		resultText += "\n"
	}

	return &CallToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// executeQueryFind executes the query_find tool
func (h *Handler) executeQueryFind(ctx context.Context, args map[string]interface{}) (*CallToolResponse, error) {
	nodeType, _ := args["type"].(string)

	filters := make(map[string]string)
	if propsRaw, ok := args["props"].(map[string]interface{}); ok {
		for k, v := range propsRaw {
			if strVal, ok := v.(string); ok {
				filters[k] = strVal
			}
		}
	}

	// Add type to filters if specified
	if nodeType != "" {
		filters["type"] = nodeType
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("at least one filter (type or props) is required")
	}

	query := graph.Query{
		Type:    "find",
		Filters: filters,
	}

	result, err := h.graph.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Format the result
	resultText := fmt.Sprintf("Found %d nodes matching criteria:\n", len(result.Nodes))
	for _, n := range result.Nodes {
		resultText += fmt.Sprintf("- %s (type: %s)", n.ID, n.Type)
		if len(n.Props) > 0 {
			resultText += " {"
			first := true
			for k, v := range n.Props {
				if !first {
					resultText += ", "
				}
				resultText += fmt.Sprintf("%s: %s", k, v)
				first = false
			}
			resultText += "}"
		}
		resultText += "\n"
	}

	return &CallToolResponse{
		Content: []ContentItem{
			{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// writeResponse writes a JSON-RPC response to the output stream
func (h *Handler) writeResponse(response *JSONRPCResponse) error {
	data, err := response.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	data = append(data, '\n')

	if _, err := h.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

// debugLog logs debug messages to stderr if debug mode is enabled
func (h *Handler) debugLog(format string, args ...interface{}) {
	if h.debug {
		log.Printf("[MCP] "+format, args...)
	}
}

// ProcessSingleRequest processes a single JSON-RPC request and returns the response as JSON
// This is useful for testing and non-interactive usage
func (h *Handler) ProcessSingleRequest(ctx context.Context, requestJSON string) (string, error) {
	response := h.processRequest(ctx, requestJSON)

	data, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(data), nil
}
