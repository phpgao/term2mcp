package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phpgao/term2go"
)

// --- get_variable ---

type GetVariableInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID"`
	Name      string `json:"name" jsonschema:"required, the variable name (e.g. session.name, session.tmuxWindowTitle)"`
}

type GetVariableOutput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func toolGetVariable() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_variable",
		Description: "Get the value of an iTerm2 session variable. Useful variables include: session.name, session.jobName, session.pid, session.path, session.tmuxWindowTitle, session.tmuxStatus, etc.",
	}
}

func handleGetVariable(ctx context.Context, req *mcp.CallToolRequest, input GetVariableInput) (
	*mcp.CallToolResult,
	*GetVariableOutput,
	error,
) {
	conn := ConnOrDie()
	value, err := term2go.GetVariable(ctx, conn, input.SessionID, []string{input.Name})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get variable: %w", err)
	}

	raw := ""
	if len(value) > 0 {
		raw = value[0]
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("`%s` = %s", input.Name, raw)},
		},
	}, &GetVariableOutput{Name: input.Name, Value: raw}, nil
}

// --- set_variable ---

type SetVariableInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID"`
	Name      string `json:"name" jsonschema:"required, the variable name to set"`
	Value     string `json:"value" jsonschema:"required, the value to set"`
}

type SetVariableOutput struct {
	Ok bool `json:"ok"`
}

func toolSetVariable() *mcp.Tool {
	return &mcp.Tool{
		Name:        "set_variable",
		Description: "Set the value of an iTerm2 session variable",
	}
}

func handleSetVariable(ctx context.Context, req *mcp.CallToolRequest, input SetVariableInput) (
	*mcp.CallToolResult,
	*SetVariableOutput,
	error,
) {
	conn := ConnOrDie()
	if err := term2go.SetVariable(ctx, conn, input.SessionID, input.Name, input.Value); err != nil {
		return nil, nil, fmt.Errorf("failed to set variable: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Set `%s` = %s on %s", input.Name, input.Value, input.SessionID)},
		},
	}, &SetVariableOutput{Ok: true}, nil
}
