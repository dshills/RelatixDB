package mcp

import (
	"encoding/json"
	"fmt"
)

// Command constants
const (
	CmdAddNode    = "add_node"
	CmdAddEdge    = "add_edge"
	CmdDeleteNode = "delete_node"
	CmdDeleteEdge = "delete_edge"
	CmdQuery      = "query"
)

// Command represents an MCP command received via stdin
type Command struct {
	Cmd  string          `json:"cmd"`
	Args json.RawMessage `json:"args"`
}

// Response represents an MCP response sent via stdout
type Response struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// AddNodeArgs represents arguments for add_node command
type AddNodeArgs struct {
	ID    string            `json:"id"`
	Type  string            `json:"type,omitempty"`
	Props map[string]string `json:"props,omitempty"`
}

// AddEdgeArgs represents arguments for add_edge command
type AddEdgeArgs struct {
	From  string            `json:"from"`
	To    string            `json:"to"`
	Label string            `json:"label"`
	Props map[string]string `json:"props,omitempty"`
}

// DeleteNodeArgs represents arguments for delete_node command
type DeleteNodeArgs struct {
	ID string `json:"id"`
}

// DeleteEdgeArgs represents arguments for delete_edge command
type DeleteEdgeArgs struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

// QueryArgs represents arguments for query command
type QueryArgs struct {
	Type      string            `json:"type"` // "neighbors", "paths", "find"
	Node      string            `json:"node,omitempty"`
	Label     string            `json:"label,omitempty"`
	Direction string            `json:"direction,omitempty"` // "in", "out", "both"
	MaxDepth  int               `json:"max_depth,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
	From      string            `json:"from,omitempty"` // for path queries
	To        string            `json:"to,omitempty"`   // for path queries
}

// ParseCommand parses a JSON command string into a Command struct
func ParseCommand(data []byte) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	if cmd.Cmd == "" {
		return nil, fmt.Errorf("command field is required")
	}

	return &cmd, nil
}

// ParseArgs parses the command arguments into the appropriate struct
func (c *Command) ParseArgs() (interface{}, error) {
	switch c.Cmd {
	case CmdAddNode:
		var args AddNodeArgs
		if err := json.Unmarshal(c.Args, &args); err != nil {
			return nil, fmt.Errorf("failed to parse add_node args: %w", err)
		}
		return args, nil

	case CmdAddEdge:
		var args AddEdgeArgs
		if err := json.Unmarshal(c.Args, &args); err != nil {
			return nil, fmt.Errorf("failed to parse add_edge args: %w", err)
		}
		return args, nil

	case CmdDeleteNode:
		var args DeleteNodeArgs
		if err := json.Unmarshal(c.Args, &args); err != nil {
			return nil, fmt.Errorf("failed to parse delete_node args: %w", err)
		}
		return args, nil

	case CmdDeleteEdge:
		var args DeleteEdgeArgs
		if err := json.Unmarshal(c.Args, &args); err != nil {
			return nil, fmt.Errorf("failed to parse delete_edge args: %w", err)
		}
		return args, nil

	case CmdQuery:
		var args QueryArgs
		if err := json.Unmarshal(c.Args, &args); err != nil {
			return nil, fmt.Errorf("failed to parse query args: %w", err)
		}
		return args, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", c.Cmd)
	}
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(result interface{}) *Response {
	return &Response{
		OK:     true,
		Result: result,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) *Response {
	return &Response{
		OK:    false,
		Error: err.Error(),
	}
}

// ToJSON converts the response to JSON bytes
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// Validate validates AddNodeArgs
func (a *AddNodeArgs) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("node ID is required")
	}
	return nil
}

// Validate validates AddEdgeArgs
func (a *AddEdgeArgs) Validate() error {
	if a.From == "" {
		return fmt.Errorf("from node is required")
	}
	if a.To == "" {
		return fmt.Errorf("to node is required")
	}
	if a.Label == "" {
		return fmt.Errorf("edge label is required")
	}
	return nil
}

// Validate validates DeleteNodeArgs
func (a *DeleteNodeArgs) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("node ID is required")
	}
	return nil
}

// Validate validates DeleteEdgeArgs
func (a *DeleteEdgeArgs) Validate() error {
	if a.From == "" {
		return fmt.Errorf("from node is required")
	}
	if a.To == "" {
		return fmt.Errorf("to node is required")
	}
	if a.Label == "" {
		return fmt.Errorf("edge label is required")
	}
	return nil
}

// Validate validates QueryArgs
func (a *QueryArgs) Validate() error {
	if a.Type == "" {
		return fmt.Errorf("query type is required")
	}

	switch a.Type {
	case "neighbors":
		if a.Node == "" {
			return fmt.Errorf("node is required for neighbors query")
		}
		if a.Direction != "" && a.Direction != "in" && a.Direction != "out" && a.Direction != "both" {
			return fmt.Errorf("invalid direction: must be 'in', 'out', or 'both'")
		}
	case "paths":
		if a.From == "" {
			return fmt.Errorf("from node is required for paths query")
		}
		if a.To == "" {
			return fmt.Errorf("to node is required for paths query")
		}
	case "find":
		if len(a.Filters) == 0 {
			return fmt.Errorf("filters are required for find query")
		}
	default:
		return fmt.Errorf("unknown query type: %s", a.Type)
	}

	return nil
}
