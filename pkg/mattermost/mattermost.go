// Package mattermost is a small client for a Mattermost server's REST API and
// WebSocket, acting as a bot account with a personal access token.
package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coder/websocket"
)

const maxBody = 4 << 20

type Client struct {
	server string
	token  string
	http   *http.Client
}

// New returns a client for the server at serverURL. The server is chosen by a
// workspace, so the caller decides which addresses httpClient may dial.
func New(serverURL, token string, httpClient *http.Client) *Client {
	return &Client{server: strings.TrimSuffix(serverURL, "/"), token: token, http: httpClient}
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsBot    bool   `json:"is_bot"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const (
	ChannelOpen    = "O"
	ChannelPrivate = "P"
	ChannelDirect  = "D"
	ChannelGroup   = "G"
)

type Channel struct {
	ID          string `json:"id"`
	TeamID      string `json:"team_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type Post struct {
	ID        string `json:"id,omitempty"`
	CreateAt  int64  `json:"create_at,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	ChannelID string `json:"channel_id"`
	RootID    string `json:"root_id,omitempty"`
	Message   string `json:"message"`
	Type      string `json:"type,omitempty"`
}

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("mattermost: %s (HTTP %d)", e.Message, e.StatusCode)
}

// IsUnauthorized reports whether the server turned the token down.
func IsUnauthorized(err error) bool {
	var apiErr *Error

	return errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden)
}

func (c *Client) Me(ctx context.Context) (User, error) {
	var me User

	err := c.call(ctx, http.MethodGet, "/api/v4/users/me", nil, &me)

	return me, err
}

func (c *Client) User(ctx context.Context, id string) (User, error) {
	var user User

	err := c.call(ctx, http.MethodGet, "/api/v4/users/"+url.PathEscape(id), nil, &user)

	return user, err
}

// Channels returns the public and private channels the bot is a member of,
// across all its teams.
func (c *Client) Channels(ctx context.Context) ([]Channel, error) {
	var teams []Team

	if err := c.call(ctx, http.MethodGet, "/api/v4/users/me/teams", nil, &teams); err != nil {
		return nil, err
	}

	var channels []Channel

	for _, team := range teams {
		var teamChannels []Channel

		if err := c.call(ctx, http.MethodGet, "/api/v4/users/me/teams/"+url.PathEscape(team.ID)+"/channels", nil, &teamChannels); err != nil {
			return nil, err
		}

		for _, channel := range teamChannels {
			if channel.Type == ChannelOpen || channel.Type == ChannelPrivate {
				channels = append(channels, channel)
			}
		}
	}

	return channels, nil
}

func (c *Client) Channel(ctx context.Context, id string) (Channel, error) {
	var channel Channel

	err := c.call(ctx, http.MethodGet, "/api/v4/channels/"+url.PathEscape(id), nil, &channel)

	return channel, err
}

func (c *Client) CreatePost(ctx context.Context, post Post) (Post, error) {
	var created Post

	err := c.call(ctx, http.MethodPost, "/api/v4/posts", post, &created)

	return created, err
}

func (c *Client) EditPost(ctx context.Context, id, message string) error {
	return c.call(ctx, http.MethodPut, "/api/v4/posts/"+url.PathEscape(id)+"/patch", map[string]string{"message": message}, nil)
}

func (c *Client) DeletePost(ctx context.Context, id string) error {
	return c.call(ctx, http.MethodDelete, "/api/v4/posts/"+url.PathEscape(id), nil, nil)
}

func (c *Client) Typing(ctx context.Context, userID, channelID, parentID string) error {
	return c.call(ctx, http.MethodPost, "/api/v4/users/"+url.PathEscape(userID)+"/typing", map[string]string{"channel_id": channelID, "parent_id": parentID}, nil)
}

// Permalink is the server's link to a post, which finds the post's team.
func (c *Client) Permalink(postID string) string {
	return c.server + "/_redirect/pl/" + url.PathEscape(postID)
}

func (c *Client) call(ctx context.Context, method, path string, body, result any) error {
	var reader io.Reader = http.NoBody

	if body != nil {
		data, err := json.Marshal(body)

		if err != nil {
			return err
		}

		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.server+path, reader)

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))

	if err != nil {
		return err
	}

	if len(data) > maxBody {
		return fmt.Errorf("mattermost: response exceeds %d bytes", maxBody)
	}

	if resp.StatusCode >= 300 {
		var apiErr struct {
			Message string `json:"message"`
		}

		_ = json.Unmarshal(data, &apiErr)

		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}

		return &Error{StatusCode: resp.StatusCode, Message: apiErr.Message}
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(data, result)
}

// Posted is a new post, as the WebSocket delivers it.
type Posted struct {
	Post        Post
	ChannelType string
}

// Listen connects to the server's WebSocket and calls posted for every new
// post the bot can see, until ctx is done or the connection drops.
func (c *Client) Listen(ctx context.Context, posted func(Posted)) error {
	wsURL := "wss" + strings.TrimPrefix(c.server, "https") + "/api/v4/websocket"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPClient: c.http, HTTPHeader: header})

	if err != nil {
		return err
	}

	defer conn.CloseNow()

	conn.SetReadLimit(maxBody)

	for {
		_, data, err := conn.Read(ctx)

		if err != nil {
			return err
		}

		var event struct {
			Event string `json:"event"`
			Data  struct {
				ChannelType string `json:"channel_type"`
				Post        string `json:"post"`
			} `json:"data"`
		}

		if json.Unmarshal(data, &event) != nil || event.Event != "posted" {
			continue
		}

		var post Post

		if json.Unmarshal([]byte(event.Data.Post), &post) != nil {
			continue
		}

		posted(Posted{Post: post, ChannelType: event.Data.ChannelType})
	}
}
