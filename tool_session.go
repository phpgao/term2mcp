package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/phpgao/term2go"
	iterm2 "github.com/phpgao/term2go/proto"
)

type (
	SessionsInput  struct{}
	SessionsOutput struct {
		Windows []WindowInfo `json:"windows"`
		Total   int          `json:"total_sessions"`
	}

	WindowInfo struct {
		WindowID string    `json:"window_id"`
		Number   int32     `json:"number"`
		Tabs     []TabInfo `json:"tabs"`
	}

	TabInfo struct {
		TabID    string        `json:"tab_id"`
		Sessions []SessionInfo `json:"sessions"`
	}

	SessionInfo struct {
		SessionID string `json:"session_id"`
		Name      string `json:"name,omitempty"`
		JobName   string `json:"job_name,omitempty"`
	}
)

func toolListSessions() *mcp.Tool {
	return &mcp.Tool{
		Name:        "list_sessions",
		Description: "List all iTerm2 sessions (windows, tabs, and panes) with their IDs and names",
	}
}

func handleListSessions(ctx context.Context, req *mcp.CallToolRequest, input SessionsInput) (
	*mcp.CallToolResult,
	*SessionsOutput,
	error,
) {
	conn := ConnOrDie()
	app, err := term2go.GetApp(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get app: %w", err)
	}

	if len(app.Windows) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No iTerm2 windows found."},
			},
		}, &SessionsOutput{}, nil
	}

	var output SessionsOutput
	var lines []string
	lines = append(lines, "# iTerm2 Sessions\n")

	for _, w := range app.Windows {
		wi := WindowInfo{WindowID: w.ID, Number: w.Number}
		lines = append(lines, fmt.Sprintf("## Window %d (%s)\n", w.Number, w.ID))

		for _, t := range w.Tabs {
			ti := TabInfo{TabID: t.ID}
			lines = append(lines, fmt.Sprintf("- **Tab** `%s`", t.ID))

			for _, s := range t.Root.Sessions() {
				si := SessionInfo{SessionID: s.GetID()}
				// Try to get name and job
				if name, err := s.GetVariable(ctx, "session.name"); err == nil {
					si.Name = name
				}
				if job, err := s.GetVariable(ctx, "session.jobName"); err == nil {
					si.JobName = job
				}
				info := fmt.Sprintf("  - Session `%s`", si.SessionID)
				if si.Name != "" {
					info += fmt.Sprintf(" [%s]", si.Name)
				}
				if si.JobName != "" {
					info += fmt.Sprintf(" (%s)", si.JobName)
				}
				lines = append(lines, info)
				ti.Sessions = append(ti.Sessions, si)
				output.Total++
			}
			wi.Tabs = append(wi.Tabs, ti)
		}
		lines = append(lines, "")
		output.Windows = append(output.Windows, wi)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(lines, "\n")},
		},
	}, &output, nil
}

// --- send_text ---

type SendTextInput struct {
	SessionID         string `json:"session_id" jsonschema:"required, the session ID (e.g. window_0_tab_0_session_0)"`
	Text              string `json:"text" jsonschema:"required, the text to send (use \\n for Return/Enter)"`
	SuppressBroadcast bool   `json:"suppress_broadcast" jsonschema:"if true, suppress broadcast to other sessions"`
}

type SendTextOutput struct {
	Ok bool `json:"ok"`
}

func toolSendText() *mcp.Tool {
	return &mcp.Tool{
		Name:        "send_text",
		Description: "Send text (or a command) to an iTerm2 session as if typed. Use \\n at the end to execute the command.",
	}
}

func handleSendText(ctx context.Context, req *mcp.CallToolRequest, input SendTextInput) (
	*mcp.CallToolResult,
	*SendTextOutput,
	error,
) {
	conn := ConnOrDie()
	var opts []term2go.SendTextOption
	if input.SuppressBroadcast {
		opts = append(opts, term2go.WithSendTextSuppressBroadcast(true))
	}
	if err := term2go.SendText(ctx, conn, input.SessionID, input.Text, opts...); err != nil {
		return nil, nil, fmt.Errorf("failed to send text: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Sent to %s: %q", input.SessionID, input.Text)},
		},
	}, &SendTextOutput{Ok: true}, nil
}

// --- get_buffer ---

type GetBufferInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID"`
	Lines     int32  `json:"lines" jsonschema:"number of lines to fetch from the end (default 50, max 500)"`
}

type GetBufferOutput struct {
	Lines []string `json:"lines"`
	Count int      `json:"total_lines_returned"`
}

func toolGetBuffer() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_buffer",
		Description: "Read the terminal buffer content from a session. Returns recent output lines.",
	}
}

func handleGetBuffer(ctx context.Context, req *mcp.CallToolRequest, input GetBufferInput) (
	*mcp.CallToolResult,
	*GetBufferOutput,
	error,
) {
	if input.Lines <= 0 {
		input.Lines = 50
	}
	if input.Lines > 500 {
		input.Lines = 500
	}

	conn := ConnOrDie()
	lines := input.Lines
	resp, err := term2go.GetBuffer(ctx, conn, input.SessionID, &iterm2.LineRange{
		TrailingLines: &lines,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get buffer: %w", err)
	}

	sc := term2go.NewScreenContents(resp)
	var output GetBufferOutput
	var textLines []string

	for _, line := range sc.Lines() {
		text := line.Text()
		textLines = append(textLines, text)
		output.Lines = append(output.Lines, text)
	}
	output.Count = len(output.Lines)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(textLines, "\n")},
		},
	}, &output, nil
}

// --- activate ---

type ActivateInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID to activate/focus"`
}

type ActivateOutput struct {
	Ok string `json:"ok"`
}

func toolActivate() *mcp.Tool {
	return &mcp.Tool{
		Name:        "activate",
		Description: "Activate (focus) a specific iTerm2 session, bringing its window and tab to the front",
	}
}

func handleActivate(ctx context.Context, req *mcp.CallToolRequest, input ActivateInput) (
	*mcp.CallToolResult,
	*ActivateOutput,
	error,
) {
	conn := ConnOrDie()
	if err := term2go.Activate(ctx, conn, input.SessionID, true, true); err != nil {
		return nil, nil, fmt.Errorf("failed to activate: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Activated session %s", input.SessionID)},
		},
	}, &ActivateOutput{Ok: input.SessionID}, nil
}

// --- inject_keystrokes ---

type InjectKeystrokesInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID to inject keystrokes into"`
	Keys      string `json:"keys" jsonschema:"required, the keystrokes/bytes to inject"`
}

type InjectKeystrokesOutput struct {
	Ok bool `json:"ok"`
}

func toolInjectKeystrokes() *mcp.Tool {
	return &mcp.Tool{
		Name:        "inject_keystrokes",
		Description: "Inject raw bytes/keystrokes into a session. More low-level than send_text.",
	}
}

func handleInjectKeystrokes(ctx context.Context, req *mcp.CallToolRequest, input InjectKeystrokesInput) (
	*mcp.CallToolResult,
	*InjectKeystrokesOutput,
	error,
) {
	conn := ConnOrDie()
	if err := term2go.Inject(ctx, conn, []string{input.SessionID}, []byte(input.Keys)); err != nil {
		return nil, nil, fmt.Errorf("failed to inject keystrokes: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Injected %q into %s", input.Keys, input.SessionID)},
		},
	}, &InjectKeystrokesOutput{Ok: true}, nil
}

// --- close_session ---

type CloseSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID to close"`
	Force     bool   `json:"force" jsonschema:"if true, skip confirmation prompt"`
}

type CloseSessionOutput struct {
	Ok bool `json:"ok"`
}

func toolCloseSession() *mcp.Tool {
	return &mcp.Tool{
		Name:        "close_session",
		Description: "Close an iTerm2 session (pane). Use force=true to skip confirmation.",
	}
}

func handleCloseSession(ctx context.Context, req *mcp.CallToolRequest, input CloseSessionInput) (
	*mcp.CallToolResult,
	*CloseSessionOutput,
	error,
) {
	conn := ConnOrDie()
	var opts []term2go.CloseOption
	if input.Force {
		opts = append(opts, term2go.WithCloseForce(true))
	}
	if err := term2go.Close(ctx, conn, input.SessionID, opts...); err != nil {
		return nil, nil, fmt.Errorf("failed to close session: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Closed session %s", input.SessionID)},
		},
	}, &CloseSessionOutput{Ok: true}, nil
}

// --- set_name ---

type SetNameInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID to rename"`
	Name      string `json:"name" jsonschema:"required, the new name for the session"`
}

type SetNameOutput struct {
	Ok bool `json:"ok"`
}

func toolSetName() *mcp.Tool {
	return &mcp.Tool{
		Name:        "set_name",
		Description: "Set the name of an iTerm2 session (displayed in the tab/title bar)",
	}
}

func handleSetName(ctx context.Context, req *mcp.CallToolRequest, input SetNameInput) (
	*mcp.CallToolResult,
	*SetNameOutput,
	error,
) {
	conn := ConnOrDie()
	app, err := term2go.GetApp(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get app: %w", err)
	}
	for _, w := range app.Windows {
		for _, t := range w.Tabs {
			for _, s := range t.Root.Sessions() {
				if s.GetID() == input.SessionID {
					if err := s.SetName(ctx, input.Name); err != nil {
						return nil, nil, fmt.Errorf("failed to set name: %w", err)
					}
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Session %s renamed to %q", input.SessionID, input.Name)},
						},
					}, &SetNameOutput{Ok: true}, nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("session not found: %s", input.SessionID)
}

// --- set_badge ---

type SetBadgeInput struct {
	SessionID string `json:"session_id" jsonschema:"required, the session ID"`
	Text      string `json:"text" jsonschema:"required, the badge text to display. Empty string to clear."`
}

type SetBadgeOutput struct {
	Ok bool `json:"ok"`
}

func toolSetBadge() *mcp.Tool {
	return &mcp.Tool{
		Name:        "set_badge",
		Description: "Set the badge text of an iTerm2 session (displayed as an overlay)",
	}
}

func handleSetBadge(ctx context.Context, req *mcp.CallToolRequest, input SetBadgeInput) (
	*mcp.CallToolResult,
	*SetBadgeOutput,
	error,
) {
	conn := ConnOrDie()
	app, err := term2go.GetApp(ctx, conn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get app: %w", err)
	}
	for _, w := range app.Windows {
		for _, t := range w.Tabs {
			for _, s := range t.Root.Sessions() {
				if s.GetID() == input.SessionID {
					if err := s.SetBadge(ctx, input.Text); err != nil {
						return nil, nil, fmt.Errorf("failed to set badge: %w", err)
					}
					msg := fmt.Sprintf("Set badge on %s to %q", input.SessionID, input.Text)
					if input.Text == "" {
						msg = fmt.Sprintf("Cleared badge on %s", input.SessionID)
					}
					return &mcp.CallToolResult{
						Content: []mcp.Content{&mcp.TextContent{Text: msg}},
					}, &SetBadgeOutput{Ok: true}, nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("session not found: %s", input.SessionID)
}
