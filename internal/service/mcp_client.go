package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// MCPTool is a model-visible tool exposed by the MCP server.
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPToolClient lists and calls tools from an MCP server.
type MCPToolClient interface {
	ListTools(ctx context.Context) ([]MCPTool, error)
	CallTool(ctx context.Context, name string, arguments map[string]any) (map[string]any, error)
}

type stdioMCPClient struct {
	command        string
	args           []string
	backendBaseURL string
	timeout        time.Duration
}

// NewStdioMCPClient creates a stdio MCP client for the configured server command.
func NewStdioMCPClient(command string, args []string, backendBaseURL string, timeout time.Duration) MCPToolClient {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &stdioMCPClient{
		command:        command,
		args:           append([]string(nil), args...),
		backendBaseURL: backendBaseURL,
		timeout:        timeout,
	}
}

func (client *stdioMCPClient) ListTools(ctx context.Context) ([]MCPTool, error) {
	session, err := client.startSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.close()

	if err := session.initialize(); err != nil {
		return nil, err
	}

	response, err := session.request("tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}

	var result struct {
		Tools []MCPTool `json:"tools"`
	}
	if err := decodeMCPResult(response.Result, &result); err != nil {
		return nil, err
	}

	return result.Tools, nil
}

func (client *stdioMCPClient) CallTool(ctx context.Context, name string, arguments map[string]any) (map[string]any, error) {
	session, err := client.startSession(ctx)
	if err != nil {
		return nil, err
	}
	defer session.close()

	if err := session.initialize(); err != nil {
		return nil, err
	}

	response, err := session.request("tools/call", map[string]any{
		"name":      name,
		"arguments": arguments,
	})
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := decodeMCPResult(response.Result, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (client *stdioMCPClient) startSession(ctx context.Context) (*mcpSession, error) {
	if client.command == "" {
		return nil, errors.New("MCP server command is not configured")
	}

	sessionCtx, cancel := context.WithTimeout(ctx, client.timeout)
	cmd := exec.CommandContext(sessionCtx, client.command, client.args...)
	cmd.Env = os.Environ()
	if client.backendBaseURL != "" {
		cmd.Env = append(cmd.Env, "MINIBIGC_API_BASE_URL="+client.backendBaseURL)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("open MCP stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("open MCP stdout: %w", err)
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start MCP server: %w%s", err, formatMCPStderr(stderr))
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	return &mcpSession{
		cmd:     cmd,
		cancel:  cancel,
		stdin:   stdin,
		stderr:  stderr,
		scanner: scanner,
		nextID:  1,
	}, nil
}

type mcpSession struct {
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	stdin   io.WriteCloser
	stderr  *bytes.Buffer
	scanner *bufio.Scanner
	nextID  int
	mu      sync.Mutex
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (session *mcpSession) initialize() error {
	if _, err := session.request("initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "minibigc-backend",
			"version": "1.0.0",
		},
	}); err != nil {
		return err
	}

	return session.notification("notifications/initialized", map[string]any{})
}

func (session *mcpSession) request(method string, params map[string]any) (*mcpResponse, error) {
	session.mu.Lock()
	id := session.nextID
	session.nextID++
	session.mu.Unlock()

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	if err := session.writeJSON(request); err != nil {
		return nil, err
	}

	for session.scanner.Scan() {
		var response mcpResponse
		if err := json.Unmarshal(session.scanner.Bytes(), &response); err != nil {
			return nil, fmt.Errorf("decode MCP response: %w%s", err, formatMCPStderr(session.stderr))
		}
		if response.ID != id {
			continue
		}
		if response.Error != nil {
			return nil, fmt.Errorf("MCP %s failed: %s%s", method, response.Error.Message, formatMCPStderr(session.stderr))
		}
		return &response, nil
	}
	if err := session.scanner.Err(); err != nil {
		return nil, fmt.Errorf("read MCP response: %w%s", err, formatMCPStderr(session.stderr))
	}
	return nil, fmt.Errorf("%w%s", io.ErrUnexpectedEOF, formatMCPStderr(session.stderr))
}

func (session *mcpSession) notification(method string, params map[string]any) error {
	return session.writeJSON(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
}

func (session *mcpSession) writeJSON(value map[string]any) error {
	if err := json.NewEncoder(session.stdin).Encode(value); err != nil {
		return fmt.Errorf("write MCP request: %w", err)
	}
	return nil
}

func (session *mcpSession) close() {
	_ = session.stdin.Close()
	session.cancel()
	_ = session.cmd.Wait()
}

func formatMCPStderr(stderr *bytes.Buffer) string {
	if stderr == nil || stderr.Len() == 0 {
		return ""
	}

	text := strings.TrimSpace(stderr.String())
	if text == "" {
		return ""
	}
	if len(text) > 2000 {
		text = text[:2000] + "... truncated"
	}

	return ": " + text
}

func decodeMCPResult(raw json.RawMessage, target any) error {
	if len(raw) == 0 {
		return errors.New("empty MCP result")
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode MCP result: %w", err)
	}
	return nil
}
