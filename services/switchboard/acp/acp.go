package acp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const protocolVersion = 1

var ErrExited = errors.New("agent exited")

type Handler interface {
	Update(sessionID string, update Update)
	Permission(ctx context.Context, request PermissionRequest) string
}

type Command struct {
	Path string
	Args []string
	Dir  string
	Env  []string
}

type Update struct {
	Kind       string          `json:"sessionUpdate"`
	Content    json.RawMessage `json:"content"`
	ToolCallID string          `json:"toolCallId"`
	Title      string          `json:"title"`
	ToolKind   string          `json:"kind"`
	Status     string          `json:"status"`
	RawInput   json.RawMessage `json:"rawInput"`
}

func (u Update) Text() string {
	var block struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(u.Content, &block) != nil || block.Type != "text" {
		return ""
	}
	return block.Text
}

type ToolCall struct {
	ToolCallID string          `json:"toolCallId"`
	Title      string          `json:"title"`
	Kind       string          `json:"kind"`
	RawInput   json.RawMessage `json:"rawInput"`
}

type PermissionOption struct {
	OptionID string `json:"optionId"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
}

type PermissionRequest struct {
	SessionID string             `json:"sessionId"`
	ToolCall  ToolCall           `json:"toolCall"`
	Options   []PermissionOption `json:"options"`
}

type Client struct {
	handler   Handler
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
	writeMu   sync.Mutex
	nextID    atomic.Int64
	mu        sync.Mutex
	pending   map[int64]chan reply
	exitErr   error
	canResume bool
}

type incoming struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type outgoing struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  any             `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("%s (code %d)", e.Message, e.Code)
}

type reply struct {
	result json.RawMessage
	err    error
}

func Start(ctx context.Context, command Command, handler Handler) (*Client, error) {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, command.Path, command.Args...)
	cmd.Dir = command.Dir
	cmd.Env = command.Env
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 10 * time.Second

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start %s: %w", command.Path, err)
	}

	client := &Client{
		handler: handler,
		cmd:     cmd,
		stdin:   stdin,
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
		pending: map[int64]chan reply{},
	}
	go client.read(stdout)

	if err := client.initialize(ctx); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func (c *Client) NewSession(ctx context.Context, cwd string, mcpServers []json.RawMessage) (string, error) {
	var result struct {
		SessionID string `json:"sessionId"`
	}
	err := c.call(ctx, "session/new", map[string]any{"cwd": cwd, "mcpServers": servers(mcpServers)}, &result)
	return result.SessionID, err
}

func (c *Client) ResumeSession(ctx context.Context, sessionID, cwd string, mcpServers []json.RawMessage) error {
	method := "session/load"
	if c.canResume {
		method = "session/resume"
	}
	return c.call(ctx, method, map[string]any{"sessionId": sessionID, "cwd": cwd, "mcpServers": servers(mcpServers)}, nil)
}

func (c *Client) Prompt(ctx context.Context, sessionID, text string) (string, error) {
	var result struct {
		StopReason string `json:"stopReason"`
	}
	params := map[string]any{
		"sessionId": sessionID,
		"prompt":    []map[string]string{{"type": "text", "text": text}},
	}

	finished := make(chan error, 1)
	go func() { finished <- c.call(c.ctx, "session/prompt", params, &result) }()

	select {
	case err := <-finished:
		return result.StopReason, err
	case <-ctx.Done():
		if err := c.Cancel(sessionID); err != nil {
			slog.Warn("failed to cancel prompt", "session", sessionID, "error", err)
		}
		err := <-finished
		return result.StopReason, err
	}
}

func (c *Client) Cancel(sessionID string) error {
	return c.write(outgoing{JSONRPC: "2.0", Method: "session/cancel", Params: map[string]string{"sessionId": sessionID}})
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.exitErr
}

func (c *Client) Close() {
	c.cancel()
	<-c.done
}

func (c *Client) initialize(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	var result struct {
		ProtocolVersion   int `json:"protocolVersion"`
		AgentCapabilities struct {
			SessionCapabilities struct {
				Resume *struct{} `json:"resume"`
			} `json:"sessionCapabilities"`
		} `json:"agentCapabilities"`
	}
	err := c.call(ctx, "initialize", map[string]any{
		"protocolVersion": protocolVersion,
		"clientCapabilities": map[string]any{
			"fs":       map[string]bool{"readTextFile": false, "writeTextFile": false},
			"terminal": false,
		},
		"clientInfo": map[string]string{"name": "switchboard", "version": "0.1.0"},
	}, &result)
	if err != nil {
		return err
	}
	if result.ProtocolVersion != protocolVersion {
		return fmt.Errorf("agent speaks ACP version %d, switchboard speaks %d", result.ProtocolVersion, protocolVersion)
	}
	c.canResume = result.AgentCapabilities.SessionCapabilities.Resume != nil
	return nil
}

func (c *Client) call(ctx context.Context, method string, params, result any) error {
	id := c.nextID.Add(1)
	replies := make(chan reply, 1)

	c.mu.Lock()
	if c.pending == nil {
		c.mu.Unlock()
		return ErrExited
	}
	c.pending[id] = replies
	c.mu.Unlock()

	request := outgoing{JSONRPC: "2.0", ID: json.RawMessage(strconv.FormatInt(id, 10)), Method: method, Params: params}
	if err := c.write(request); err != nil {
		c.forget(id)
		return fmt.Errorf("%s: %w", method, err)
	}

	select {
	case r := <-replies:
		if r.err != nil {
			return fmt.Errorf("%s: %w", method, r.err)
		}
		if result == nil {
			return nil
		}
		return json.Unmarshal(r.result, result)
	case <-ctx.Done():
		c.forget(id)
		return ctx.Err()
	}
}

func (c *Client) forget(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pending, id)
}

func (c *Client) write(message outgoing) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err = c.stdin.Write(append(payload, '\n'))
	return err
}

func (c *Client) read(stdout io.Reader) {
	reader := bufio.NewReaderSize(stdout, 1<<20)
	for {
		line, err := reader.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			c.dispatch(line)
		}
		if err != nil {
			break
		}
	}

	waitErr := c.cmd.Wait()
	c.mu.Lock()
	c.exitErr = waitErr
	pending := c.pending
	c.pending = nil
	c.mu.Unlock()

	for _, replies := range pending {
		replies <- reply{err: ErrExited}
	}
	c.cancel()
	close(c.done)
}

func (c *Client) dispatch(line []byte) {
	var message incoming
	if err := json.Unmarshal(line, &message); err != nil {
		slog.Warn("undecodable message from agent", "error", err)
		return
	}
	switch {
	case message.Method != "" && len(message.ID) > 0:
		go c.serve(message)
	case message.Method != "":
		c.notification(message)
	default:
		c.resolve(message)
	}
}

func (c *Client) resolve(message incoming) {
	id, err := strconv.ParseInt(string(message.ID), 10, 64)
	if err != nil {
		return
	}
	c.mu.Lock()
	replies, ok := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()
	if !ok {
		return
	}
	if message.Error != nil {
		replies <- reply{err: message.Error}
		return
	}
	replies <- reply{result: message.Result}
}

func (c *Client) notification(message incoming) {
	if message.Method != "session/update" {
		return
	}
	var notification struct {
		SessionID string `json:"sessionId"`
		Update    Update `json:"update"`
	}
	if err := json.Unmarshal(message.Params, &notification); err != nil {
		slog.Warn("undecodable session update", "error", err)
		return
	}
	c.handler.Update(notification.SessionID, notification.Update)
}

func (c *Client) serve(message incoming) {
	if message.Method != "session/request_permission" {
		c.respond(message.ID, nil, &rpcError{Code: -32601, Message: "method not found: " + message.Method})
		return
	}
	var request PermissionRequest
	if err := json.Unmarshal(message.Params, &request); err != nil {
		c.respond(message.ID, nil, &rpcError{Code: -32602, Message: "invalid params"})
		return
	}

	outcome := map[string]string{"outcome": "cancelled"}
	if optionID := c.handler.Permission(c.ctx, request); optionID != "" {
		outcome = map[string]string{"outcome": "selected", "optionId": optionID}
	}
	c.respond(message.ID, map[string]any{"outcome": outcome}, nil)
}

func (c *Client) respond(id json.RawMessage, result any, rpcErr *rpcError) {
	if err := c.write(outgoing{JSONRPC: "2.0", ID: id, Result: result, Error: rpcErr}); err != nil {
		slog.Warn("failed to answer agent request", "error", err)
	}
}

func servers(mcpServers []json.RawMessage) []json.RawMessage {
	if mcpServers == nil {
		return []json.RawMessage{}
	}
	return mcpServers
}
