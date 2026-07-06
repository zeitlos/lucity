package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/zeitlos/lucity/cli/internal/api"
)

const (
	subscriptionIdleTimeout  = 4 * time.Second
	subscriptionTotalTimeout = 20 * time.Second
)

type wsMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func (s *server) subscribeLogs(ctx context.Context, query string, variables map[string]any, selector func(map[string]any) []string, tail int) ([]string, bool, error) {
	token, err := s.manager.Token(ctx)
	if err != nil {
		return nil, false, err
	}

	endpoint := wsEndpoint(s.manager.APIURL())

	ctx, cancel := context.WithTimeout(ctx, subscriptionTotalTimeout)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{
		Subprotocols: []string{"graphql-transport-ws"},
	})
	if err != nil {
		return nil, false, fmt.Errorf("connect log stream at %s: %w", endpoint, err)
	}
	defer conn.CloseNow()
	conn.SetReadLimit(16 << 20)

	initPayload, _ := json.Marshal(map[string]any{
		"Authorization":     "Bearer " + token,
		api.WorkspaceHeader: s.manager.Workspace(),
	})
	if err := writeMessage(ctx, conn, wsMessage{Type: "connection_init", Payload: initPayload}); err != nil {
		return nil, false, err
	}

	if err := awaitAck(ctx, conn); err != nil {
		return nil, false, err
	}

	subPayload, _ := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err := writeMessage(ctx, conn, wsMessage{ID: "1", Type: "subscribe", Payload: subPayload}); err != nil {
		return nil, false, err
	}

	lines, complete, err := collectLogs(ctx, conn, selector, tail)
	_ = conn.Close(websocket.StatusNormalClosure, "")
	return lines, complete, err
}

func wsEndpoint(apiURL string) string {
	endpoint := strings.TrimSuffix(apiURL, "/") + "/graphql"
	switch {
	case strings.HasPrefix(endpoint, "https://"):
		return "wss://" + strings.TrimPrefix(endpoint, "https://")
	case strings.HasPrefix(endpoint, "http://"):
		return "ws://" + strings.TrimPrefix(endpoint, "http://")
	default:
		return endpoint
	}
}

func writeMessage(ctx context.Context, conn *websocket.Conn, message wsMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		return fmt.Errorf("write %s: %w", message.Type, err)
	}
	return nil
}

func awaitAck(ctx context.Context, conn *websocket.Conn) error {
	for {
		message, err := readMessage(ctx, conn)
		if err != nil {
			return err
		}
		switch message.Type {
		case "connection_ack":
			return nil
		case "ping":
			if err := writeMessage(ctx, conn, wsMessage{Type: "pong"}); err != nil {
				return err
			}
		case "connection_error", "error":
			return fmt.Errorf("log stream rejected the connection: %s", string(message.Payload))
		default:
		}
	}
}

func collectLogs(ctx context.Context, conn *websocket.Conn, selector func(map[string]any) []string, tail int) ([]string, bool, error) {
	var lines []string

	appendLines := func(values []string) {
		lines = append(lines, values...)
		if len(lines) > tail {
			lines = lines[len(lines)-tail:]
		}
	}

	for {
		idleCtx, cancel := context.WithTimeout(ctx, subscriptionIdleTimeout)
		message, err := readMessage(idleCtx, conn)
		cancel()
		if err != nil {
			return lines, false, nil
		}

		switch message.Type {
		case "next":
			var payload struct {
				Data   map[string]any `json:"data"`
				Errors []any          `json:"errors"`
			}
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				continue
			}
			if payload.Data != nil {
				appendLines(selector(payload.Data))
			}
		case "complete":
			return lines, true, nil
		case "error":
			if len(lines) > 0 {
				return lines, false, nil
			}
			return nil, false, fmt.Errorf("log stream error: %s", string(message.Payload))
		case "ping":
			if err := writeMessage(ctx, conn, wsMessage{Type: "pong"}); err != nil {
				return lines, false, nil
			}
		default:
		}
	}
}

func readMessage(ctx context.Context, conn *websocket.Conn) (wsMessage, error) {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return wsMessage{}, err
	}
	var message wsMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return wsMessage{}, fmt.Errorf("decode ws message: %w", err)
	}
	return message, nil
}
