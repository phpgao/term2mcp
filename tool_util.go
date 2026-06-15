package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phpgao/term2go"
)

// --- focus ---

type FocusInput struct{}

type FocusOutput struct {
	ActiveSession string `json:"active_session"`
	SelectedTab   string `json:"selected_tab"`
	ActiveWindow  string `json:"active_window"`
}

func toolFocus() *mcp.Tool {
	return &mcp.Tool{
		Name:        "focus",
		Description: "Get the currently focused iTerm2 session, tab, and window",
	}
}

func handleFocus(ctx context.Context, req *mcp.CallToolRequest, input FocusInput) (
	*mcp.CallToolResult,
	*FocusOutput,
	error,
) {
	conn := ConnOrDie()
	resp, err := term2go.FocusRequest(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("focus request failed: %w", err)
	}

	output := &FocusOutput{}
	for _, n := range resp.GetNotifications() {
		if s := n.GetSession(); s != "" {
			output.ActiveSession = s
		}
		if t := n.GetSelectedTab(); t != "" {
			output.SelectedTab = t
		}
		if w := n.GetWindow(); w != nil {
			output.ActiveWindow = w.GetWindowId()
		}
	}

	msg := "Current focus:\n"
	if output.ActiveSession != "" {
		msg += fmt.Sprintf("- Active Session: `%s`\n", output.ActiveSession)
	}
	if output.SelectedTab != "" {
		msg += fmt.Sprintf("- Selected Tab: `%s`\n", output.SelectedTab)
	}
	if output.ActiveWindow != "" {
		msg += fmt.Sprintf("- Active Window: `%s`\n", output.ActiveWindow)
	}
	if output.ActiveSession == "" && output.SelectedTab == "" && output.ActiveWindow == "" {
		msg = "No focus information available"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, output, nil
}

// --- list_profiles ---

type ListProfilesInput struct{}

type ListProfilesOutput struct {
	Profiles []string `json:"profiles"`
}

func toolListProfiles() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_profiles",
		Description: "List all iTerm2 profile names",
	}
}

func handleListProfiles(ctx context.Context, req *mcp.CallToolRequest, input ListProfilesInput) (
	*mcp.CallToolResult,
	*ListProfilesOutput,
	error,
) {
	conn := ConnOrDie()
	resp, err := term2go.ListProfiles(ctx, conn, []string{"Name"}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	var profiles []string
	for _, p := range resp.GetProfiles() {
		for _, prop := range p.GetProperties() {
			if prop.GetKey() == "Name" {
				name := prop.GetJsonValue()
				profiles = append(profiles, name)
				break
			}
		}
	}

	msg := "# iTerm2 Profiles\n\n"
	for _, name := range profiles {
		msg += fmt.Sprintf("- %s\n", name)
	}
	if len(profiles) == 0 {
		msg += "(none)"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, &ListProfilesOutput{Profiles: profiles}, nil
}
