package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("term2mcp: shutting down...")
		cancel()
	}()

	// Create MCP server with all tools registered
	server := NewServer()

	log.Printf("term2mcp %s starting on stdio...", ServerVersion)
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("term2mcp: server error: %v", err)
	}
}
