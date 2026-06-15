package main

import (
	"context"
	"log"
	"sync"

	"github.com/phpgao/term2go"
)

var (
	instance *term2go.Connection
	once     sync.Once
	initErr  error
)

// GetConn returns the singleton iTerm2 connection, creating it on first call.
// The connection is lazily established, so term2mcp starts fast even if iTerm2
// is not running.
func GetConn() (*term2go.Connection, error) {
	once.Do(func() {
		instance, initErr = term2go.Connect(context.Background(), "term2mcp")
		if initErr != nil {
			log.Printf("term2mcp: failed to connect to iTerm2: %v", initErr)
		} else {
			log.Printf("term2mcp: connected to iTerm2 (type=%s)", instance.ConnType())
		}
	})
	return instance, initErr
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
