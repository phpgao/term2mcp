# term2mcp

MCP (Model Context Protocol) server for iTerm2 — empowered by [term2go](https://github.com/phpgao/term2go).

Let AI assistants (CodeBuddy Code, Claude Desktop, Continue, etc.) control your iTerm2 sessions programmatically: list sessions, send commands, read output, split panes, take screenshots, and more.

🌐 **Website**: https://phpgao.github.io/term2mcp/

## Prerequisites

- macOS
- [iTerm2](https://iterm2.com/) with Python API enabled
  - Open iTerm2 → **Preferences** → **General** → **Magic** → check **Enable Python API**
- Go 1.23+ (only needed for `go install`)

## Installation

### From source (recommended)

```bash
go install github.com/phpgao/term2mcp@latest
```

After installation, the binary `term2mcp` will be in your `$GOPATH/bin` (usually `~/go/bin`). Make sure it's in your PATH.

### From release (future)

Pre-built binaries will be available on the [Releases](https://github.com/phpgao/term2mcp/releases) page.

## Configuration

Choose your MCP client below and follow the step-by-step setup.

### CodeBuddy Code

**Step 1 — Find your config file**

Open CodeBuddy Code, run:

```
/config
```

This opens `mcp.json` in your editor. The actual file path is:

```
~/.codebuddy/mcp.json
```

**Step 2 — Add term2mcp**

```json
{
  "mcpServers": {
    "term2mcp": {
      "command": "term2mcp",
      "args": []
    }
  }
}
```

**Step 3 — Optional: speed up startup with env vars**

```json
{
  "mcpServers": {
    "term2mcp": {
      "command": "term2mcp",
      "args": [],
      "env": {
        "ITERM2_COOKIE": "YOUR_COOKIE_HERE",
        "ITERM2_KEY": "YOUR_KEY_HERE"
      }
    }
  }
}
```

How to get these values? See [iTerm2 Credentials](#iterm2-credentials) below.

**Step 4 — Verify**

Restart CodeBuddy Code. Run `/mcp` to list servers. You should see `term2mcp` with status **connected**.

In the chat, try: "List my iTerm2 sessions" — it should show your windows and tabs.

### Claude Desktop

**Step 1 — Find your config file**

```bash
# Config file location
open ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

If the file doesn't exist, create it:

```bash
touch ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

**Step 2 — Add term2mcp**

```json
{
  "mcpServers": {
    "term2mcp": {
      "command": "term2mcp",
      "args": []
    }
  }
}
```

**Step 3 — Restart Claude Desktop**

Quit and reopen the app. You should see a hammer icon in the chat input — click it to see available tools.

### iTerm2 Credentials

term2mcp auto-detects credentials via AppleScript when iTerm2 is running. This works well but adds ~1-2s latency on first connection.

To eliminate this latency, set environment variables:

```bash
export ITERM2_COOKIE="xxx"
export ITERM2_KEY="xxx"
```

**How to get the values:**

They are stored in iTerm2's private socket directory. You can extract them with:

```bash
# Cookie
defaults read com.googlecode.iterm2 Cookie

# Key (base64 encoded)
ls ~/Library/Application\ Support/iTerm2/private/
```

Or search your environment when inside iTerm2:

```bash
# In an iTerm2 session, run:
echo $ITERM2_COOKIE
echo $ITERM2_KEY
```

These values change occasionally (especially after iTerm2 restart). If connection fails, the auto-detection fallback will handle it transparently.

## Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `list_sessions` | List all iTerm2 windows, tabs, and sessions with IDs | — |
| `send_text` | Send text/commands to a session (use `\n` for Enter) | `session_id`, `text`, `suppress_broadcast?` |
| `get_buffer` | Read terminal output from a session | `session_id`, `lines?` (default 50, max 500) |
| `create_tab` | Create a new tab (or window) with optional initial command | `window_id?`, `profile?`, `command?` |
| `split_pane` | Split a session vertically or horizontally | `session_id`, `vertical`, `before?`, `profile?` |
| `screenshot` | Capture a window or session as a PNG file | `target` (window/session), `id`, `output_path` |
| `get_variable` | Get an iTerm2 session variable | `session_id`, `name` |
| `set_variable` | Set an iTerm2 session variable | `session_id`, `name`, `value` |
| `focus` | Get the currently focused session/tab/window | — |
| `activate` | Focus and bring a specific session to the front | `session_id` |
| `inject_keystrokes` | Inject raw bytes/keystrokes into a session | `session_id`, `keys` |
| `close_session` | Close an iTerm2 session (pane) | `session_id`, `force?` |
| `set_name` | Set the name of a session (displayed in tab/title) | `session_id`, `name` |
| `set_badge` | Set the badge text overlay on a session | `session_id`, `text` |
| `get_prompt` | Get last command info (text, dir, exit code, state) | `session_id` |
| `list_profiles` | List all iTerm2 profile names | — |

## Example Conversations

```
You: List my iTerm2 sessions
AI: [list_sessions] → Window 0 has tab 0 with session "w0_t0_s0"

You: Run ls -la in that session
AI: [send_text: session="w0_t0_s0", text="ls -la\n"]

You: What's the output?
AI: [get_buffer: session="w0_t0_s0"] → total 32
                                    drwxr-xr-x  5 user  staff  160 ...
                                    -rw-r--r--  1 user  staff  123 main.go

You: Open a new tab and run htop
AI: [create_tab: command="htop"] → New tab created, session: "w0_t1_s0"

You: What was the last command's exit code?
AI: [get_prompt: session="w0_t0_s0"] → state: finished, exit: 0

You: Take a screenshot of the window
AI: [screenshot: target="window", id="...", output="/tmp/screen.png"]
```

## Troubleshooting

### "cannot connect to iTerm2"

Make sure:
1. iTerm2 is running
2. **Python API** is enabled (Preferences → General → Magic)

### Tool not showing up in MCP client

Check `~/.codebuddy/mcp.json` or Claude config for typos. The `command` must be the exact path to the `term2mcp` binary (run `which term2mcp` to check).

### Connection is slow on first use

This is normal — AppleScript auth takes ~1-2s. Set `ITERM2_COOKIE`/`ITERM2_KEY` env vars in the MCP config (see [iTerm2 Credentials](#iterm2-credentials)) to eliminate this delay.

## Development

```bash
git clone https://github.com/phpgao/term2mcp.git
cd term2mcp

make build   # Build binary
make test    # Run tests
make run     # Run directly
make install # Install to $GOPATH/bin
```

## License

MIT — see [LICENSE](LICENSE) file.
