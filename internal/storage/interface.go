package storage

import (
	"context"
	"io"

	"github.com/dshills/RelatixDB/internal/graph"
)

// Backend defines the interface for persistent storage backends
type Backend interface {
	// Open opens or creates a database at the specified path
	Open(path string) error

	// Close closes the database connection
	Close() error

	// LoadGraph loads the entire graph from storage
	LoadGraph(ctx context.Context) (graph.Graph, error)

	// SaveGraph saves the entire graph to storage
	SaveGraph(ctx context.Context, g graph.Graph) error

	// Transaction operations for atomic updates
	BeginTransaction() (Transaction, error)
}

// Transaction represents an atomic database transaction
type Transaction interface {
	// Node operations
	SaveNode(node graph.Node) error
	DeleteNode(id string) error

	// Edge operations
	SaveEdge(edge graph.Edge) error
	DeleteEdge(from, to, label string) error

	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error
}

// Serializer handles conversion between graph objects and storage format
type Serializer interface {
	// Serialize converts a graph object to bytes
	SerializeNode(node graph.Node) ([]byte, error)
	SerializeEdge(edge graph.Edge) ([]byte, error)

	// Deserialize converts bytes back to graph objects
	DeserializeNode(data []byte) (graph.Node, error)
	DeserializeEdge(data []byte) (graph.Edge, error)
}

// Backup provides backup and restore functionality
type Backup interface {
	// Export exports the graph to a writer
	Export(ctx context.Context, g graph.Graph, writer io.Writer) error

	// Import imports a graph from a reader
	Import(ctx context.Context, reader io.Reader) (graph.Graph, error)
}

// Stats provides storage statistics
type Stats struct {
	DatabaseSize   int64 `json:"database_size"`
	NodeCount      int   `json:"node_count"`
	EdgeCount      int   `json:"edge_count"`
	LastSaved      int64 `json:"last_saved"`
	LastLoaded     int64 `json:"last_loaded"`
	TransactionLog int64 `json:"transaction_log_size"`
}

// StatsProvider provides storage statistics
type StatsProvider interface {
	GetStats() (*Stats, error)
}
