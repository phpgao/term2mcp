package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phpgao/term2go"
)

// --- create_tab ---

type CreateTabInput struct {
	WindowID    string `json:"window_id" jsonschema:"window ID to create tab in. Leave empty to create a new window"`
	ProfileName string `json:"profile" jsonschema:"iTerm2 profile name (default is 'Default')"`
	Command     string `json:"command" jsonschema:"initial command to run in the new tab (optional)"`
}

type CreateTabOutput struct {
	TabID     string `json:"tab_id"`
	SessionID string `json:"session_id"`
	WindowID  string `json:"window_id"`
}

func toolCreateTab() *mcp.Tool {
	return &mcp.Tool{
		Name:        "create_tab",
		Description: "Create a new iTerm2 tab, optionally in an existing window. If no window_id is given, a new window is created.",
	}
}

func handleCreateTab(ctx context.Context, req *mcp.CallToolRequest, input CreateTabInput) (
	*mcp.CallToolResult,
	*CreateTabOutput,
	error,
) {
	conn := ConnOrDie()
	if input.ProfileName == "" {
		input.ProfileName = "Default"
	}

	resp, err := term2go.CreateTab(ctx, conn, input.WindowID, input.ProfileName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tab: %w", err)
	}

	output := &CreateTabOutput{
		TabID:     fmt.Sprintf("%d", resp.GetTabId()),
		SessionID: resp.GetSessionId(),
		WindowID:  resp.GetWindowId(),
	}

	// Run initial command if provided
	if input.Command != "" {
		if err := term2go.SendText(ctx, conn, output.SessionID, input.Command+"\n"); err != nil {
			// Don't fail — tab was created successfully
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf(
						"Tab created:\n- Tab ID: `%s`\n- Session ID: `%s`\n- Window ID: `%s`\n\nWarning: initial command failed: %v",
						output.TabID, output.SessionID, output.WindowID, err,
					)},
				},
			}, output, nil
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Tab created:\n- Tab ID: `%s`\n- Session ID: `%s`\n- Window ID: `%s`",
				output.TabID, output.SessionID, output.WindowID,
			)},
		},
	}, output, nil
}

// --- split_pane ---

type SplitPaneInput struct {
	SessionID   string `json:"session_id" jsonschema:"required, the session ID to split"`
	Vertical    bool   `json:"vertical" jsonschema:"true for vertical split (left-right), false for horizontal (top-bottom)"`
	Before      bool   `json:"before" jsonschema:"if true, place the new pane before the current one"`
	ProfileName string `json:"profile" jsonschema:"iTerm2 profile name (default is 'Default')"`
}

type SplitPaneOutput struct {
	SessionID  string `json:"new_session_id"`
	OriginalID string `json:"original_session_id"`
}

func toolSplitPane() *mcp.Tool {
	return &mcp.Tool{
		Name:        "split_pane",
		Description: "Split an iTerm2 session into two panes. Returns the new session ID.",
	}
}

func handleSplitPane(ctx context.Context, req *mcp.CallToolRequest, input SplitPaneInput) (
	*mcp.CallToolResult,
	*SplitPaneOutput,
	error,
) {
	conn := ConnOrDie()
	if input.ProfileName == "" {
		input.ProfileName = "Default"
	}

	resp, err := term2go.SplitPane(ctx, conn, input.SessionID, input.Vertical, input.Before, input.ProfileName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to split pane: %w", err)
	}

	sessionID := ""
	if ids := resp.GetSessionId(); len(ids) > 0 {
		sessionID = ids[0]
	}

	output := &SplitPaneOutput{
		SessionID:  sessionID,
		OriginalID: input.SessionID,
	}

	direction := "horizontally"
	if input.Vertical {
		direction = "vertically"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf(
				"Pane split %s from `%s`\nNew session: `%s`",
				direction, input.SessionID, output.SessionID,
			)},
		},
	}, output, nil
}
