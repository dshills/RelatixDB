package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dshills/RelatixDB/internal/graph"
	"github.com/dshills/RelatixDB/internal/mcp"
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
	
	// Print banner to stderr so it doesn't interfere with MCP communication
	if *debug {
		fmt.Fprintf(os.Stderr, banner, version)
	}
	
	// Create context that can be cancelled
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
		// TODO: Initialize persistent graph when storage layer is implemented
		if *debug {
			log.Printf("Persistent storage not yet implemented, using in-memory graph")
		}
		g = graph.NewMemoryGraph()
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