package main

import (
	"context"
	"log"
	"sync"

	"github.com/phpgao/term2go"
)

var (
	instance *term2go.Connection
	mu       sync.Mutex
)

// GetConn returns the iTerm2 connection, reconnecting if the current one is dead.
func GetConn() (*term2go.Connection, error) {
	mu.Lock()
	defer mu.Unlock()

	// Try to reuse existing connection
	if instance != nil {
		return instance, nil
	}

	conn, err := term2go.Connect(context.Background(), "term2mcp")
	if err != nil {
		log.Printf("term2mcp: failed to connect to iTerm2: %v", err)
		return nil, err
	}
	instance = conn
	instance.OnDisconnect(func() {
		mu.Lock()
		instance = nil
		mu.Unlock()
		log.Print("term2mcp: iTerm2 disconnected, will reconnect on next call")
	})
	log.Printf("term2mcp: connected to iTerm2 (type=%s)", instance.ConnType())
	return instance, nil
}

// DropConn drops the cached connection so GetConn will create a new one.
func DropConn() {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		_ = instance.Close()
		instance = nil
		log.Print("term2mcp: dropped iTerm2 connection")
	}
}

// ConnOrDie returns the connection or writes a fatal error to stderr.
func ConnOrDie() *term2go.Connection {
	conn, err := GetConn()
	if err != nil {
		log.Fatalf("term2mcp: cannot connect to iTerm2: %v\n"+
			"Make sure iTerm2 is running and the Python API is enabled.\n"+
			"Preferences → General → Magic → Enable Python API", err)
	}
	return conn
}
