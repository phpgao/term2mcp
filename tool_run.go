package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

// --- run_command ---

type RunCommandInput struct {
	WindowID    string `json:"window_id" jsonschema:"window ID to create tab in. Leave empty to create a new window"`
	ProfileName string `json:"profile" jsonschema:"iTerm2 profile name (default is 'Default')"`
	Command     string `json:"command" jsonschema:"required, the command to run"`
	Timeout     int    `json:"timeout" jsonschema:"max seconds to wait for output (default 10, max 60)"`
}

type RunCommandOutput struct {
	TabID     string `json:"tab_id"`
	SessionID string `json:"session_id"`
	WindowID  string `json:"window_id"`
	Output    string `json:"output"`
	ExitCode  int    `json:"exit_code,omitempty"`
}

func toolRunCommand() *mcp.Tool {
	return &mcp.Tool{
		Name:        "run_command",
		Description: "Create a new tab, run a command, and wait for the output to appear in the terminal buffer.",
	}
}

func readSessionBuffer(ctx context.Context, conn *term2go.Connection, sessionID string) string {
	trailing := int32(200)
	resp, err := term2go.GetBuffer(ctx, conn, sessionID, &iterm2.LineRange{
		TrailingLines: &trailing,
	})
	if err != nil {
		return ""
	}
	sc := term2go.NewScreenContents(resp)
	var buf []string
	for _, line := range sc.Lines() {
		buf = append(buf, line.Text())
	}
	return strings.Join(buf, "\n")
}

func handleRunCommand(ctx context.Context, req *mcp.CallToolRequest, input RunCommandInput) (
	*mcp.CallToolResult,
	*RunCommandOutput,
	error,
) {
	if input.Command == "" {
		return nil, nil, fmt.Errorf("command is required")
	}
	if input.Timeout <= 0 {
		input.Timeout = 10
	}
	if input.Timeout > 60 {
		input.Timeout = 60
	}
	if input.ProfileName == "" {
		input.ProfileName = "Default"
	}

	conn := ConnOrDie()

	// Step 1: Create the tab
	resp, err := term2go.CreateTab(ctx, conn, input.WindowID, input.ProfileName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tab: %w", err)
	}

	output := &RunCommandOutput{
		TabID:     fmt.Sprintf("%d", resp.GetTabId()),
		SessionID: resp.GetSessionId(),
		WindowID:  resp.GetWindowId(),
	}

	// Step 2: Send the command
	if err := term2go.SendText(ctx, conn, output.SessionID, input.Command+"\n"); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf(
					"Tab created:\n- Tab ID: `%s`\n- Session ID: `%s`\n- Window ID: `%s`\n\nCommand sent but output monitoring failed: %v",
					output.TabID, output.SessionID, output.WindowID, err,
				)},
			},
		}, output, nil
	}

	// Step 3: Poll buffer until output appears or timeout
	deadline := time.Now().Add(time.Duration(input.Timeout) * time.Second)
	var prevContent string

	for time.Now().Before(deadline) {
		current := readSessionBuffer(ctx, conn, output.SessionID)
		if current != "" && current != prevContent {
			prevContent = current
			if strings.Contains(current, input.Command) {
				// Command appeared in buffer - give it a moment to finish
				time.Sleep(500 * time.Millisecond)
				prevContent = readSessionBuffer(ctx, conn, output.SessionID)
				break
			}
		}

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}

	output.Output = strings.TrimRight(prevContent, "\x00 ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Tab created:\n- Tab ID: `%s`\n- Session ID: `%s`\n- Window ID: `%s`\n\n```\n%s\n```",
				output.TabID, output.SessionID, output.WindowID, output.Output,
			)},
		},
	}, output, nil
}
