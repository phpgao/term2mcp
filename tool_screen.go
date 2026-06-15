package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phpgao/term2go"
)

// --- screenshot ---

type ScreenshotInput struct {
	Target     string `json:"target" jsonschema:"required, what to capture: 'window' or 'session'"`
	ID         string `json:"id" jsonschema:"required, the window ID or session ID to capture"`
	OutputPath string `json:"output_path" jsonschema:"required, file path to save the screenshot (e.g. /tmp/screenshot.png)"`
}

type ScreenshotOutput struct {
	Path string `json:"path"`
	Size int64  `json:"size_bytes"`
}

func toolScreenshot() *mcp.Tool {
	return &mcp.Tool{
		Name:        "screenshot",
		Description: "Take a screenshot of an iTerm2 window or session and save it as a PNG file",
	}
}

func handleScreenshot(ctx context.Context, req *mcp.CallToolRequest, input ScreenshotInput) (
	*mcp.CallToolResult,
	*ScreenshotOutput,
	error,
) {
	conn := ConnOrDie()

	switch input.Target {
	case "window":
		app, err := term2go.GetApp(ctx, conn)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get app: %w", err)
		}
		for _, w := range app.Windows {
			if w.ID == input.ID {
				if err := w.Screenshot(ctx, input.OutputPath); err != nil {
					return nil, nil, fmt.Errorf("window screenshot failed: %w", err)
				}
				return textResult(fmt.Sprintf("Window screenshot saved to %s", input.OutputPath)),
					&ScreenshotOutput{Path: input.OutputPath}, nil
			}
		}
		return nil, nil, fmt.Errorf("window not found: %s", input.ID)

	case "session":
		app, err := term2go.GetApp(ctx, conn)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get app: %w", err)
		}
		for _, w := range app.Windows {
			for _, t := range w.Tabs {
				for _, s := range t.Root.Sessions() {
					if s.GetID() == input.ID {
						if err := s.Screenshot(ctx, input.OutputPath); err != nil {
							return nil, nil, fmt.Errorf("session screenshot failed: %w", err)
						}
						return textResult(fmt.Sprintf("Session screenshot saved to %s", input.OutputPath)),
							&ScreenshotOutput{Path: input.OutputPath}, nil
					}
				}
			}
		}
		return nil, nil, fmt.Errorf("session not found: %s", input.ID)

	default:
		return nil, nil, fmt.Errorf("invalid target: %q (must be 'window' or 'session')", input.Target)
	}
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}
