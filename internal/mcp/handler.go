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
	graph  graph.Graph
	reader *bufio.Scanner
	writer io.Writer
	debug  bool
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

// Run starts the MCP handler loop, processing commands from stdin
func (h *Handler) Run(ctx context.Context) error {
	h.debugLog("Starting MCP handler...")

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

		h.debugLog("Received command: %s", line)

		response := h.processCommand(ctx, line)

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

// processCommand processes a single MCP command and returns a response
func (h *Handler) processCommand(ctx context.Context, commandLine string) *Response {
	// Parse the command
	cmd, err := ParseCommand([]byte(commandLine))
	if err != nil {
		h.debugLog("Failed to parse command: %v", err)
		return NewErrorResponse(fmt.Errorf("failed to parse command: %w", err))
	}

	// Parse command arguments
	args, err := cmd.ParseArgs()
	if err != nil {
		h.debugLog("Failed to parse args: %v", err)
		return NewErrorResponse(fmt.Errorf("failed to parse args: %w", err))
	}

	// Execute the command
	switch cmd.Cmd {
	case CmdAddNode:
		return h.handleAddNode(ctx, args.(AddNodeArgs))
	case CmdAddEdge:
		return h.handleAddEdge(ctx, args.(AddEdgeArgs))
	case CmdDeleteNode:
		return h.handleDeleteNode(ctx, args.(DeleteNodeArgs))
	case CmdDeleteEdge:
		return h.handleDeleteEdge(ctx, args.(DeleteEdgeArgs))
	case CmdQuery:
		return h.handleQuery(ctx, args.(QueryArgs))
	default:
		return NewErrorResponse(fmt.Errorf("unknown command: %s", cmd.Cmd))
	}
}

// handleAddNode handles the add_node command
func (h *Handler) handleAddNode(ctx context.Context, args AddNodeArgs) *Response {
	if err := args.Validate(); err != nil {
		return NewErrorResponse(err)
	}

	node := graph.Node{
		ID:    args.ID,
		Type:  args.Type,
		Props: args.Props,
	}

	if err := h.graph.AddNode(ctx, node); err != nil {
		h.debugLog("Failed to add node: %v", err)
		return NewErrorResponse(err)
	}

	h.debugLog("Added node: %s", args.ID)
	return NewSuccessResponse(map[string]interface{}{
		"node_id": args.ID,
		"action":  "added",
	})
}

// handleAddEdge handles the add_edge command
func (h *Handler) handleAddEdge(ctx context.Context, args AddEdgeArgs) *Response {
	if err := args.Validate(); err != nil {
		return NewErrorResponse(err)
	}

	edge := graph.Edge{
		From:  args.From,
		To:    args.To,
		Label: args.Label,
		Props: args.Props,
	}

	if err := h.graph.AddEdge(ctx, edge); err != nil {
		h.debugLog("Failed to add edge: %v", err)
		return NewErrorResponse(err)
	}

	h.debugLog("Added edge: %s -> %s (%s)", args.From, args.To, args.Label)
	return NewSuccessResponse(map[string]interface{}{
		"from":   args.From,
		"to":     args.To,
		"label":  args.Label,
		"action": "added",
	})
}

// handleDeleteNode handles the delete_node command
func (h *Handler) handleDeleteNode(ctx context.Context, args DeleteNodeArgs) *Response {
	if err := args.Validate(); err != nil {
		return NewErrorResponse(err)
	}

	if err := h.graph.DeleteNode(ctx, args.ID); err != nil {
		h.debugLog("Failed to delete node: %v", err)
		return NewErrorResponse(err)
	}

	h.debugLog("Deleted node: %s", args.ID)
	return NewSuccessResponse(map[string]interface{}{
		"node_id": args.ID,
		"action":  "deleted",
	})
}

// handleDeleteEdge handles the delete_edge command
func (h *Handler) handleDeleteEdge(ctx context.Context, args DeleteEdgeArgs) *Response {
	if err := args.Validate(); err != nil {
		return NewErrorResponse(err)
	}

	if err := h.graph.DeleteEdge(ctx, args.From, args.To, args.Label); err != nil {
		h.debugLog("Failed to delete edge: %v", err)
		return NewErrorResponse(err)
	}

	h.debugLog("Deleted edge: %s -> %s (%s)", args.From, args.To, args.Label)
	return NewSuccessResponse(map[string]interface{}{
		"from":   args.From,
		"to":     args.To,
		"label":  args.Label,
		"action": "deleted",
	})
}

// handleQuery handles the query command
func (h *Handler) handleQuery(ctx context.Context, args QueryArgs) *Response {
	if err := args.Validate(); err != nil {
		return NewErrorResponse(err)
	}

	query := graph.Query{
		Type:      args.Type,
		Node:      args.Node,
		Label:     args.Label,
		Direction: args.Direction,
		MaxDepth:  args.MaxDepth,
		Filters:   args.Filters,
		From:      args.From,
		To:        args.To,
	}

	result, err := h.graph.Query(ctx, query)
	if err != nil {
		h.debugLog("Failed to execute query: %v", err)
		return NewErrorResponse(err)
	}

	h.debugLog("Query executed successfully, returned %d nodes, %d edges, %d paths",
		len(result.Nodes), len(result.Edges), len(result.Paths))

	return NewSuccessResponse(result)
}

// writeResponse writes a response to the output stream
func (h *Handler) writeResponse(response *Response) error {
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

// ProcessSingleCommand processes a single command and returns the response as JSON
// This is useful for testing and non-interactive usage
func (h *Handler) ProcessSingleCommand(ctx context.Context, commandJSON string) (string, error) {
	response := h.processCommand(ctx, commandJSON)

	data, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(data), nil
}
