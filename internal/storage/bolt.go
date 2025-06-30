package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.etcd.io/bbolt"

	"github.com/dshills/RelatixDB/internal/graph"
)

// BoltBackend implements the Backend interface using BoltDB
type BoltBackend struct {
	db         *bbolt.DB
	serializer Serializer
	stats      Stats
}

// BoltTransaction implements the Transaction interface for BoltDB
type BoltTransaction struct {
	tx         *bbolt.Tx
	backend    *BoltBackend
	serializer Serializer
}

const (
	nodesBucket = "nodes"
	edgesBucket = "edges"
	metaBucket  = "meta"
)

// NewBoltBackend creates a new BoltDB backend
func NewBoltBackend() *BoltBackend {
	return &BoltBackend{
		serializer: &JSONSerializer{},
	}
}

// Open opens or creates a BoltDB database
func (b *BoltBackend) Open(path string) error {
	var err error
	b.db, err = bbolt.Open(path, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create buckets if they don't exist
	err = b.db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(nodesBucket)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(edgesBucket)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(metaBucket)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		b.db.Close()
		return fmt.Errorf("failed to create buckets: %w", err)
	}

	// Update stats
	b.updateStats()

	return nil
}

// Close closes the BoltDB database
func (b *BoltBackend) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// LoadGraph loads the entire graph from BoltDB
func (b *BoltBackend) LoadGraph(ctx context.Context) (graph.Graph, error) {
	if b.db == nil {
		return nil, fmt.Errorf("database not opened")
	}

	memGraph := graph.NewMemoryGraph()

	err := b.db.View(func(tx *bbolt.Tx) error {
		// Load nodes
		nodesBucket := tx.Bucket([]byte(nodesBucket))
		if nodesBucket != nil {
			err := nodesBucket.ForEach(func(k, v []byte) error {
				node, err := b.serializer.DeserializeNode(v)
				if err != nil {
					return fmt.Errorf("failed to deserialize node %s: %w", k, err)
				}

				if err := memGraph.AddNode(ctx, node); err != nil {
					return fmt.Errorf("failed to add node %s: %w", k, err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

		// Load edges
		edgesBucket := tx.Bucket([]byte(edgesBucket))
		if edgesBucket != nil {
			err := edgesBucket.ForEach(func(k, v []byte) error {
				edge, err := b.serializer.DeserializeEdge(v)
				if err != nil {
					return fmt.Errorf("failed to deserialize edge %s: %w", k, err)
				}

				if err := memGraph.AddEdge(ctx, edge); err != nil {
					return fmt.Errorf("failed to add edge %s: %w", k, err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load graph: %w", err)
	}

	b.stats.LastLoaded = time.Now().Unix()
	return memGraph, nil
}

// SaveGraph saves the entire graph to BoltDB
func (b *BoltBackend) SaveGraph(ctx context.Context, g graph.Graph) error {
	if b.db == nil {
		return fmt.Errorf("database not opened")
	}

	// This is a simplified implementation - in production you'd want
	// to iterate over the graph more efficiently
	tx, err := b.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// TODO: Implement full graph iteration and saving
	// This requires extending the Graph interface to provide:
	// - GetAllNodes() method
	// - GetAllEdges() method
	// For now, individual operations are persisted immediately via transactions

	return fmt.Errorf("SaveGraph not fully implemented - requires Graph interface extension for iteration")
}

// BeginTransaction starts a new transaction
func (b *BoltBackend) BeginTransaction() (Transaction, error) {
	if b.db == nil {
		return nil, fmt.Errorf("database not opened")
	}

	tx, err := b.db.Begin(true)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &BoltTransaction{
		tx:         tx,
		backend:    b,
		serializer: b.serializer,
	}, nil
}

// SaveNode saves a node in the transaction
func (bt *BoltTransaction) SaveNode(node graph.Node) error {
	bucket := bt.tx.Bucket([]byte(nodesBucket))
	if bucket == nil {
		return fmt.Errorf("nodes bucket not found")
	}

	data, err := bt.serializer.SerializeNode(node)
	if err != nil {
		return fmt.Errorf("failed to serialize node: %w", err)
	}

	return bucket.Put([]byte(node.ID), data)
}

// DeleteNode deletes a node in the transaction
func (bt *BoltTransaction) DeleteNode(id string) error {
	bucket := bt.tx.Bucket([]byte(nodesBucket))
	if bucket == nil {
		return fmt.Errorf("nodes bucket not found")
	}

	return bucket.Delete([]byte(id))
}

// SaveEdge saves an edge in the transaction
func (bt *BoltTransaction) SaveEdge(edge graph.Edge) error {
	bucket := bt.tx.Bucket([]byte(edgesBucket))
	if bucket == nil {
		return fmt.Errorf("edges bucket not found")
	}

	data, err := bt.serializer.SerializeEdge(edge)
	if err != nil {
		return fmt.Errorf("failed to serialize edge: %w", err)
	}

	key := fmt.Sprintf("%s:%s:%s", edge.From, edge.To, edge.Label)
	return bucket.Put([]byte(key), data)
}

// DeleteEdge deletes an edge in the transaction
func (bt *BoltTransaction) DeleteEdge(from, to, label string) error {
	bucket := bt.tx.Bucket([]byte(edgesBucket))
	if bucket == nil {
		return fmt.Errorf("edges bucket not found")
	}

	key := fmt.Sprintf("%s:%s:%s", from, to, label)
	return bucket.Delete([]byte(key))
}

// Commit commits the transaction
func (bt *BoltTransaction) Commit() error {
	err := bt.tx.Commit()
	if err == nil {
		bt.backend.updateStats()
		bt.backend.stats.LastSaved = time.Now().Unix()
	}
	return err
}

// Rollback rolls back the transaction
func (bt *BoltTransaction) Rollback() error {
	return bt.tx.Rollback()
}

// updateStats updates internal statistics
func (b *BoltBackend) updateStats() {
	if b.db == nil {
		return
	}

	b.db.View(func(tx *bbolt.Tx) error {
		// Get database file size (simplified approach)
		if stat := b.db.Stats(); stat.TxStats.PageCount > 0 {
			// Use a reasonable page size estimate
			b.stats.DatabaseSize = int64(stat.TxStats.PageCount) * 4096
		}

		// Count nodes
		if bucket := tx.Bucket([]byte(nodesBucket)); bucket != nil {
			b.stats.NodeCount = bucket.Stats().KeyN
		}

		// Count edges
		if bucket := tx.Bucket([]byte(edgesBucket)); bucket != nil {
			b.stats.EdgeCount = bucket.Stats().KeyN
		}

		return nil
	})
}

// GetStats returns storage statistics
func (b *BoltBackend) GetStats() (*Stats, error) {
	b.updateStats()
	return &b.stats, nil
}

// JSONSerializer implements JSON serialization for graph objects
type JSONSerializer struct{}

// SerializeNode converts a node to JSON bytes
func (s *JSONSerializer) SerializeNode(node graph.Node) ([]byte, error) {
	return json.Marshal(node)
}

// SerializeEdge converts an edge to JSON bytes
func (s *JSONSerializer) SerializeEdge(edge graph.Edge) ([]byte, error) {
	return json.Marshal(edge)
}

// DeserializeNode converts JSON bytes to a node
func (s *JSONSerializer) DeserializeNode(data []byte) (graph.Node, error) {
	var node graph.Node
	err := json.Unmarshal(data, &node)
	return node, err
}

// DeserializeEdge converts JSON bytes to an edge
func (s *JSONSerializer) DeserializeEdge(data []byte) (graph.Edge, error) {
	var edge graph.Edge
	err := json.Unmarshal(data, &edge)
	return edge, err
}
