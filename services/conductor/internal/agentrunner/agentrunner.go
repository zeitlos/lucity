// Package agentrunner talks to the runner in an agent's pod, through the
// apiserver's service proxy. The runner executes arbitrary code next to the
// harness, so everything it returns is checked and size-limited.
package agentrunner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

const (
	port        = 8080
	callTimeout = 15 * time.Second

	maxBody        = 1 << 20
	maxSessionBody = 16 << 20
	maxEventBytes  = 1 << 20
)

var (
	// ErrUnreachable means the runner didn't answer, usually because the
	// agent is still starting.
	ErrUnreachable = usererror.New("the agent can't be reached yet, try again in a moment")

	errInvalidResponse = errors.New("agent runner returned an invalid response")
)

type Client struct {
	kubernetes kubernetes.Interface
	http       *http.Client
	base       *url.URL
}

func New(config *rest.Config, kubernetes kubernetes.Interface) (*Client, error) {
	httpClient, err := rest.HTTPClientFor(config)

	if err != nil {
		return nil, err
	}

	base, _, err := rest.DefaultServerUrlFor(config)

	if err != nil {
		return nil, err
	}

	return &Client{kubernetes: kubernetes, http: httpClient, base: base}, nil
}

// ServiceName is the name of the Service in front of an agent's runner.
func ServiceName(agent string) string {
	return "lucity-app-agent-" + agent
}

func (c *Client) url(agent platform.AgentID, segments []string, query url.Values) string {
	path := "/api/v1/namespaces/" + url.PathEscape(agent.Namespace()) + "/services/" + url.PathEscape(fmt.Sprintf("%s:%d", ServiceName(agent.Name), port)) + "/proxy"

	for _, segment := range segments {
		path += "/" + url.PathEscape(segment)
	}

	target := strings.TrimSuffix(c.base.String(), "/") + path

	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	return target
}

func (c *Client) call(ctx context.Context, agent platform.AgentID, method string, segments []string, body any, limit int64, result any) error {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	var reader io.Reader = http.NoBody

	if body != nil {
		data, err := json.Marshal(body)

		if err != nil {
			return err
		}

		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.url(agent, segments, nil), reader)

	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return ErrUnreachable
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))

	if err != nil {
		return ErrUnreachable
	}

	if int64(len(data)) > limit {
		return errInvalidResponse
	}

	if resp.StatusCode >= 300 {
		return responseError(resp.StatusCode, data)
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(data, result); err != nil {
		return errInvalidResponse
	}

	return nil
}

func responseError(code int, data []byte) error {
	var body struct {
		Error string `json:"error"`
	}

	switch code {
	case http.StatusNotFound:
		if json.Unmarshal(data, &body) == nil && body.Error != "" {
			return usererror.Errorf("the agent says: %s", printable(body.Error))
		}

		return usererror.New("not found")
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return ErrUnreachable
	}

	if code >= 400 && code < 500 && json.Unmarshal(data, &body) == nil && body.Error != "" {
		return usererror.Errorf("the agent says: %s", printable(body.Error))
	}

	return fmt.Errorf("agent runner answered HTTP %d", code)
}

func printable(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}

		return r
	}, s)

	return cut(s, maxAPIMessage)
}

func (c *Client) Status(ctx context.Context, agent platform.AgentID) (*Status, error) {
	var status Status

	if err := c.call(ctx, agent, http.MethodGet, []string{"status"}, nil, maxBody, &status); err != nil {
		return nil, err
	}

	if !sanitizeStatus(&status) {
		return nil, errInvalidResponse
	}

	return &status, nil
}

func (c *Client) Sessions(ctx context.Context, agent platform.AgentID) ([]Session, error) {
	var sessions []Session

	if err := c.call(ctx, agent, http.MethodGet, []string{"sessions"}, nil, maxBody, &sessions); err != nil {
		return nil, err
	}

	result := make([]Session, 0, min(len(sessions), maxSessions))

	for _, session := range sessions {
		if len(result) >= maxSessions {
			break
		}

		session.Events = nil

		if sanitizeSession(&session) {
			result = append(result, session)
		}
	}

	return result, nil
}

func (c *Client) Session(ctx context.Context, id platform.AgentSessionID) (*Session, error) {
	return c.session(ctx, id.Agent, http.MethodGet, []string{"sessions", id.Session}, nil, maxSessionBody)
}

type StartRequest struct {
	Message   string `json:"message"`
	Model     Model  `json:"model"`
	Effort    string `json:"effort"`
	StartedBy string `json:"startedBy"`
}

func (c *Client) StartSession(ctx context.Context, agent platform.AgentID, request StartRequest) (*Session, error) {
	return c.session(ctx, agent, http.MethodPost, []string{"sessions"}, request, maxBody)
}

func (c *Client) SendMessage(ctx context.Context, id platform.AgentSessionID, message, author string) (*Session, error) {
	return c.session(ctx, id.Agent, http.MethodPost, []string{"sessions", id.Session, "messages"}, map[string]string{"message": message, "author": author}, maxBody)
}

func (c *Client) SetModel(ctx context.Context, id platform.AgentSessionID, model Model, effort string) (*Session, error) {
	return c.session(ctx, id.Agent, http.MethodPost, []string{"sessions", id.Session, "model"}, map[string]any{"model": model, "effort": effort}, maxBody)
}

func (c *Client) AnswerApproval(ctx context.Context, id platform.AgentSessionID, approval, decision, decidedBy string) (*Session, error) {
	return c.session(ctx, id.Agent, http.MethodPost, []string{"sessions", id.Session, "approvals", approval}, map[string]string{"decision": decision, "decidedBy": decidedBy}, maxBody)
}

func (c *Client) Stop(ctx context.Context, id platform.AgentSessionID) (*Session, error) {
	return c.session(ctx, id.Agent, http.MethodPost, []string{"sessions", id.Session, "stop"}, nil, maxBody)
}

func (c *Client) session(ctx context.Context, agent platform.AgentID, method string, segments []string, body any, limit int64) (*Session, error) {
	var session Session

	if err := c.call(ctx, agent, method, segments, body, limit, &session); err != nil {
		return nil, err
	}

	if !sanitizeSession(&session) {
		return nil, errInvalidResponse
	}

	return &session, nil
}

// Events streams the events of a session after the event after, or from the
// start when it is empty. The channel closes when the stream ends or ctx is
// done.
func (c *Client) Events(ctx context.Context, id platform.AgentSessionID, after string) (<-chan Event, error) {
	query := url.Values{}

	if after != "" {
		query.Set("after", after)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(id.Agent, []string{"sessions", id.Session, "events"}, query), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)

	if err != nil {
		return nil, ErrUnreachable
	}

	if resp.StatusCode >= 300 {
		defer resp.Body.Close()

		data, _ := io.ReadAll(io.LimitReader(resp.Body, maxBody))

		return nil, responseError(resp.StatusCode, data)
	}

	events := make(chan Event, 64)

	go func() {
		defer close(events)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 64*1024), maxEventBytes)

		var data bytes.Buffer

		for scanner.Scan() {
			line := scanner.Bytes()

			if len(line) > 0 {
				if value, ok := bytes.CutPrefix(line, []byte("data:")); ok {
					if data.Len()+len(value) > maxEventBytes {
						return
					}

					if data.Len() > 0 {
						data.WriteByte('\n')
					}

					data.Write(bytes.TrimPrefix(value, []byte(" ")))
				}

				continue
			}

			if data.Len() == 0 {
				continue
			}

			var event Event

			err := json.Unmarshal(data.Bytes(), &event)
			data.Reset()

			if err != nil || !sanitizeEvent(&event) {
				continue
			}

			select {
			case events <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, nil
}

// Failed reports whether an agent's pod can't start, as opposed to one that
// is still starting.
func (c *Client) Failed(ctx context.Context, agent platform.AgentID) (bool, error) {
	pods, err := c.kubernetes.CoreV1().Pods(agent.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: labels.Selector(labels.Agent, agent.Name),
	})

	if err != nil {
		return false, err
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase == core.PodFailed {
			return true, nil
		}

		for _, status := range pod.Status.ContainerStatuses {
			if waiting := status.State.Waiting; waiting != nil {
				switch waiting.Reason {
				case "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull", "CreateContainerConfigError", "InvalidImageName":
					return true, nil
				}
			}
		}
	}

	return false, nil
}
