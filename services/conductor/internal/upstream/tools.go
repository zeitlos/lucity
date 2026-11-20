package upstream

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/zeitlos/lucity/pkg/deployer/values"
)

const (
	protocolVersion       = "2025-06-18"
	maxTools              = 1000
	maxToolPages          = 10
	maxToolDescriptionLen = 1024
)

type Tool struct {
	Name        string
	Description string
}

type rpcResponse struct {
	ID     json.RawMessage `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type mcpSession struct {
	client     *Client
	url        string
	credential string
	id         string
	version    string
	nextID     int
}

// Tools connects to the remote MCP server at url over Streamable HTTP, with
// credential as a bearer token when set, and lists its tools.
func (c *Client) Tools(ctx context.Context, url, credential string) ([]Tool, Check) {
	session := &mcpSession{client: c, url: url, credential: credential}

	var initialized struct {
		ProtocolVersion string `json:"protocolVersion"`
	}

	if check := session.call(ctx, "initialize", map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "lucity", "version": "1"},
	}, &initialized); check.Status != Valid {
		return nil, check
	}

	session.version = initialized.ProtocolVersion

	defer session.close()

	if check := session.notify(ctx, "notifications/initialized"); check.Status != Valid {
		return nil, check
	}

	var tools []Tool

	cursor := ""

	for range maxToolPages {
		params := map[string]any{}

		if cursor != "" {
			params["cursor"] = cursor
		}

		var page struct {
			Tools []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"tools"`
			NextCursor string `json:"nextCursor"`
		}

		if check := session.call(ctx, "tools/list", params, &page); check.Status != Valid {
			return nil, check
		}

		for _, tool := range page.Tools {
			if len(tools) >= maxTools {
				break
			}

			if !values.IsValidToolName(tool.Name) {
				continue
			}

			tools = append(tools, Tool{Name: tool.Name, Description: truncate(tool.Description, maxToolDescriptionLen)})
		}

		if page.NextCursor == "" || len(tools) >= maxTools {
			break
		}

		cursor = page.NextCursor
	}

	return tools, Check{Status: Valid}
}

func (s *mcpSession) request(ctx context.Context, method string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, s.url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	if s.credential != "" {
		req.Header.Set("Authorization", "Bearer "+s.credential)
	}

	if s.id != "" {
		req.Header.Set("Mcp-Session-Id", s.id)
	}

	if s.version != "" {
		req.Header.Set("MCP-Protocol-Version", s.version)
	}

	return req, nil
}

func (s *mcpSession) call(ctx context.Context, method string, params any, result any) Check {
	s.nextID++
	id := s.nextID

	req, err := s.request(ctx, http.MethodPost, map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})

	if err != nil {
		return unreachable(err)
	}

	resp, body, err := s.client.do(req)

	if err != nil {
		return unreachable(err)
	}

	if resp.StatusCode != http.StatusOK {
		return statusCheck(resp.StatusCode)
	}

	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" && s.id == "" {
		if !isPrintable(sessionID, 256) {
			return Check{Status: Unreachable, Reason: "it isn't a remote MCP server"}
		}

		s.id = sessionID
	}

	response, ok := findResponse(resp.Header.Get("Content-Type"), body, id)

	if !ok {
		return Check{Status: Unreachable, Reason: "it isn't a remote MCP server"}
	}

	if response.Error != nil {
		return Check{Status: Unreachable, Reason: "it answered " + method + " with an error"}
	}

	if err := json.Unmarshal(response.Result, result); err != nil {
		return Check{Status: Unreachable, Reason: "it answered " + method + " with something unexpected"}
	}

	return Check{Status: Valid}
}

func (s *mcpSession) notify(ctx context.Context, method string) Check {
	req, err := s.request(ctx, http.MethodPost, map[string]any{"jsonrpc": "2.0", "method": method})

	if err != nil {
		return unreachable(err)
	}

	resp, _, err := s.client.do(req)

	if err != nil {
		return unreachable(err)
	}

	if resp.StatusCode >= 300 {
		return statusCheck(resp.StatusCode)
	}

	return Check{Status: Valid}
}

func (s *mcpSession) close() {
	if s.id == "" {
		return
	}

	req, err := s.request(context.Background(), http.MethodDelete, nil)

	if err != nil {
		return
	}

	req.Body = http.NoBody
	req.ContentLength = 0

	s.client.do(req)
}

func findResponse(contentType string, body []byte, id int) (rpcResponse, bool) {
	mediaType, _, _ := mime.ParseMediaType(contentType)

	if mediaType != "text/event-stream" {
		return matchResponse(body, id)
	}

	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 64*1024), maxBody)

	var data strings.Builder

	flush := func() (rpcResponse, bool) {
		defer data.Reset()

		if data.Len() == 0 {
			return rpcResponse{}, false
		}

		return matchResponse([]byte(data.String()), id)
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if response, ok := flush(); ok {
				return response, true
			}

			continue
		}

		if value, ok := strings.CutPrefix(line, "data:"); ok {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}

			data.WriteString(strings.TrimPrefix(value, " "))
		}
	}

	return flush()
}

func matchResponse(data []byte, id int) (rpcResponse, bool) {
	var response rpcResponse

	if err := json.Unmarshal(data, &response); err != nil {
		return rpcResponse{}, false
	}

	var got int

	if err := json.Unmarshal(response.ID, &got); err != nil || got != id {
		return rpcResponse{}, false
	}

	return response, response.Result != nil || response.Error != nil
}

func truncate(s string, maxRunes int) string {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}

	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}

	return string([]rune(s)[:maxRunes])
}
