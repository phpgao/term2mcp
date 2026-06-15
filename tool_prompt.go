package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phpgao/term2go"
)

// --- get_prompt ---

type GetPromptInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID"`
}

type GetPromptOutput struct {
	Command          string `json:"command"`
	WorkingDirectory string `json:"working_directory"`
	ExitStatus       uint32 `json:"exit_status"`
	State            string `json:"state"`
}

func toolGetPrompt() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_prompt",
		Description: "Get information about the last command prompt: command text, working directory, exit status",
	}
}

func handleGetPrompt(ctx context.Context, req *mcp.CallToolRequest, input GetPromptInput) (
	*mcp.CallToolResult,
	*GetPromptOutput,
	error,
) {
	conn := ConnOrDie()

	prompt, err := term2go.GetLastPrompt(ctx, conn, input.SessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	output := &GetPromptOutput{
		Command:          prompt.Command(),
		WorkingDirectory: prompt.WorkingDirectory(),
		ExitStatus:       prompt.ExitStatus(),
	}

	switch prompt.State() {
	case term2go.PromptEditing:
		output.State = "editing"
	case term2go.PromptRunning:
		output.State = "running"
	case term2go.PromptFinished:
		output.State = "finished"
	default:
		output.State = "unknown"
	}

	var lines []string
	lines = append(lines, "# Last Command Prompt\n")
	if output.Command != "" {
		lines = append(lines, fmt.Sprintf("- Command: `%s`", output.Command))
	}
	if output.WorkingDirectory != "" {
		lines = append(lines, fmt.Sprintf("- Working Directory: `%s`", output.WorkingDirectory))
	}
	lines = append(lines, fmt.Sprintf("- State: %s", output.State))
	lines = append(lines, fmt.Sprintf("- Exit Status: %d", output.ExitStatus))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(lines, "\n")},
		},
	}, output, nil
}
