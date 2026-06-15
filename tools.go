package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ServerName    = "term2mcp"
	ServerVersion = "v0.1.0"
)

// NewServer creates an MCP server with all tools registered.
func NewServer() *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    ServerName,
			Version: ServerVersion,
		},
		nil,
	)

	// Session tools
	mcp.AddTool(server, toolListSessions(), handleListSessions)
	mcp.AddTool(server, toolSendText(), handleSendText)
	mcp.AddTool(server, toolGetBuffer(), handleGetBuffer)
	mcp.AddTool(server, toolActivate(), handleActivate)
	mcp.AddTool(server, toolInjectKeystrokes(), handleInjectKeystrokes)
	mcp.AddTool(server, toolCloseSession(), handleCloseSession)
	mcp.AddTool(server, toolSetName(), handleSetName)
	mcp.AddTool(server, toolSetBadge(), handleSetBadge)

	// Pane/Tab tools
	mcp.AddTool(server, toolCreateTab(), handleCreateTab)
	mcp.AddTool(server, toolSplitPane(), handleSplitPane)

	// Variable tools
	mcp.AddTool(server, toolGetVariable(), handleGetVariable)
	mcp.AddTool(server, toolSetVariable(), handleSetVariable)

	// Screen tools
	mcp.AddTool(server, toolScreenshot(), handleScreenshot)

	// Prompt tools
	mcp.AddTool(server, toolGetPrompt(), handleGetPrompt)

	// Utility tools
	mcp.AddTool(server, toolFocus(), handleFocus)
	mcp.AddTool(server, toolListProfiles(), handleListProfiles)

	return server
}
