package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/dshills/RelatixDB/internal/graph"
	"github.com/dshills/RelatixDB/internal/mcp"
	"github.com/dshills/RelatixDB/internal/storage"
)

const (
	version = "0.1.0"
	banner  = `
RelatixDB v%s
High-performance local graph database for MCP tools
`
)

func main() {
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
		debug       = flag.Bool("debug", false, "Enable debug logging")
		dbPath      = flag.String("db", "", "Database file path (optional, uses in-memory if not specified)")
		dumpPath    = flag.String("dump", "", "Pretty print contents of database file and exit")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("RelatixDB v%s\n", version)
		return
	}

	if *showHelp {
		showUsage()
		return
	}

	if *dumpPath != "" {
		dumpDatabase(*dumpPath, *debug)
		return
	}

	// Print banner to stderr so it doesn't interfere with MCP communication
	if *debug {
		fmt.Fprintf(os.Stderr, banner, version)
	}

	// Create context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		if *debug {
			log.Printf("Received signal: %v, shutting down...", sig)
		}
		cancel()
	}()

	// Initialize the graph
	var g graph.Graph
	if *dbPath != "" {
		// Initialize persistent graph with BoltDB backend
		backend := storage.NewBoltBackend()
		if err := backend.Open(*dbPath); err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer backend.Close()

		// Disable auto-save until SaveGraph is fully implemented
		persistentGraph := storage.NewPersistentGraph(backend, false, 30*time.Second)
		if err := persistentGraph.Load(ctx); err != nil {
			if *debug {
				log.Printf("No existing database found, starting with empty graph: %v", err)
			}
		}
		defer persistentGraph.Close()

		g = persistentGraph
		if *debug {
			log.Printf("Using persistent graph storage at %s", *dbPath)
		}
	} else {
		g = graph.NewMemoryGraph()
		if *debug {
			log.Printf("Using in-memory graph storage")
		}
	}

	// Create MCP handler
	handler := mcp.NewStdioHandler(g, *debug)

	// Run the MCP handler
	if err := handler.Run(ctx); err != nil {
		if err == context.Canceled {
			if *debug {
				log.Println("Shutdown completed")
			}
			return
		}
		log.Fatalf("Handler error: %v", err)
	}
}

func dumpDatabase(dbPath string, debug bool) {
	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Database file does not exist: %s\n", dbPath)
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize persistent graph with BoltDB backend
	backend := storage.NewBoltBackend()
	if err := backend.Open(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer backend.Close()

	// Create persistent graph
	persistentGraph := storage.NewPersistentGraph(backend, false, 30*time.Second)
	if err := persistentGraph.Load(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load database: %v\n", err)
		os.Exit(1)
	}
	defer persistentGraph.Close()

	// Get all nodes and edges
	nodes, err := persistentGraph.GetAllNodes(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get nodes: %v\n", err)
		os.Exit(1)
	}

	edges, err := persistentGraph.GetAllEdges(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get edges: %v\n", err)
		os.Exit(1)
	}

	// Pretty print the database contents
	fmt.Printf("RelatixDB Database Contents: %s\n", dbPath)
	fmt.Printf("=====================================\n\n")

	// Print statistics
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Nodes: %d\n", len(nodes))
	fmt.Printf("  Edges: %d\n", len(edges))

	// Count types
	typeCount := make(map[string]int)
	for _, node := range nodes {
		if node.Type != "" {
			typeCount[node.Type]++
		} else {
			typeCount["<no-type>"]++
		}
	}
	fmt.Printf("  Types: %d\n", len(typeCount))
	fmt.Printf("\n")

	// Print type breakdown
	if len(typeCount) > 0 {
		fmt.Printf("Node Types:\n")
		var types []string
		for t := range typeCount {
			types = append(types, t)
		}
		sort.Strings(types)
		for _, t := range types {
			fmt.Printf("  %s: %d nodes\n", t, typeCount[t])
		}
		fmt.Printf("\n")
	}

	// Print nodes
	if len(nodes) > 0 {
		fmt.Printf("Nodes:\n")
		fmt.Printf("------\n")

		// Sort nodes by ID for consistent output
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].ID < nodes[j].ID
		})

		for _, node := range nodes {
			printNode(node, debug)
		}
		fmt.Printf("\n")
	}

	// Print edges
	if len(edges) > 0 {
		fmt.Printf("Edges:\n")
		fmt.Printf("------\n")

		// Sort edges by from, to, label for consistent output
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].From != edges[j].From {
				return edges[i].From < edges[j].From
			}
			if edges[i].To != edges[j].To {
				return edges[i].To < edges[j].To
			}
			return edges[i].Label < edges[j].Label
		})

		for _, edge := range edges {
			printEdge(edge, debug)
		}
	}

	if len(nodes) == 0 && len(edges) == 0 {
		fmt.Printf("Database is empty.\n")
	}
}

func printNode(node graph.Node, debug bool) {
	fmt.Printf("Node: %s", node.ID)
	if node.Type != "" {
		fmt.Printf(" (type: %s)", node.Type)
	}
	fmt.Printf("\n")

	printProperties(node.Props, debug)
	fmt.Printf("\n")
}

func printEdge(edge graph.Edge, debug bool) {
	fmt.Printf("Edge: %s -> %s [%s]\n", edge.From, edge.To, edge.Label)

	printProperties(edge.Props, debug)
	fmt.Printf("\n")
}

func printProperties(props map[string]string, debug bool) {
	if len(props) == 0 {
		return
	}

	// Sort properties for consistent output
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("  Properties:\n")
	for _, key := range keys {
		value := props[key]
		if debug {
			// In debug mode, show raw JSON representation
			if jsonData, err := json.Marshal(value); err == nil {
				fmt.Printf("    %s: %s\n", key, string(jsonData))
			} else {
				fmt.Printf("    %s: %s\n", key, value)
			}
		} else {
			// Pretty print with truncation for long values
			if len(value) > 100 {
				fmt.Printf("    %s: %s...\n", key, value[:97])
			} else {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
	}
}

func showUsage() {
	fmt.Printf(banner, version)
	fmt.Println("USAGE:")
	fmt.Println("  relatixdb [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -version      Show version information")
	fmt.Println("  -help         Show this help message")
	fmt.Println("  -debug        Enable debug logging to stderr")
	fmt.Println("  -db PATH      Database file path (optional, uses in-memory if not specified)")
	fmt.Println("  -dump PATH    Pretty print contents of database file and exit")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  RelatixDB is a high-performance local graph database designed for use as an")
	fmt.Println("  MCP (Model Context Protocol) tool server. It operates via JSON commands on")
	fmt.Println("  stdin with JSON responses on stdout.")
	fmt.Println()
	fmt.Println("  The database supports:")
	fmt.Println("  - Nodes with unique IDs, optional types, and key/value properties")
	fmt.Println("  - Directed, labeled edges between nodes with optional metadata")
	fmt.Println("  - Fast queries for neighbors, paths, and property-based searches")
	fmt.Println()
	fmt.Println("EXAMPLE COMMANDS:")
	fmt.Println("  Add a node:")
	fmt.Println(`    {"cmd": "add_node", "args": {"id": "user:1", "type": "user", "props": {"name": "Alice"}}}`)
	fmt.Println()
	fmt.Println("  Add an edge:")
	fmt.Println(`    {"cmd": "add_edge", "args": {"from": "user:1", "to": "user:2", "label": "follows"}}`)
	fmt.Println()
	fmt.Println("  Query neighbors:")
	fmt.Println(`    {"cmd": "query", "args": {"type": "neighbors", "node": "user:1", "direction": "out"}}`)
	fmt.Println()
	fmt.Println("  Find nodes by type:")
	fmt.Println(`    {"cmd": "query", "args": {"type": "find", "filters": {"type": "user"}}}`)
	fmt.Println()
	fmt.Println("PERFORMANCE TARGETS:")
	fmt.Println("  - Node insertion: < 100µs")
	fmt.Println("  - Edge insertion: < 150µs")
	fmt.Println("  - Neighborhood query (1-hop): < 1ms")
	fmt.Println("  - Path query (depth ≤ 4): < 10ms")
	fmt.Println()
}
