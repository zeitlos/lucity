// Package rybbit reads the sessions the platform's own sites report to its
// Rybbit web analytics.
package rybbit

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const sessionLimit = 100

type Client struct {
	baseURL string
	apiKey  string
	site    int
	http    *http.Client
}

func New(baseURL, apiKey string, site int) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		site:    site,
		http:    &http.Client{Timeout: time.Minute},
	}
}

type Session struct {
	ID              string
	StartedAt       time.Time
	Duration        time.Duration
	EntryPage       string
	Pageviews       int
	Country         string
	City            string
	Language        string
	Browser         string
	OperatingSystem string
	DeviceType      string
	Referrer        string
	HasReplay       bool
}

type session struct {
	SessionID       string `json:"session_id"`
	SessionStart    string `json:"session_start"`
	SessionDuration int64  `json:"session_duration"`
	EntryPage       string `json:"entry_page"`
	Pageviews       int    `json:"pageviews"`
	Country         string `json:"country"`
	City            string `json:"city"`
	Language        string `json:"language"`
	Browser         string `json:"browser"`
	OperatingSystem string `json:"operating_system"`
	DeviceType      string `json:"device_type"`
	Referrer        string `json:"referrer"`
	HasReplay       int    `json:"has_replay"`
}

// Sessions returns the newest sessions of an identified user, newest first.
func (c *Client) Sessions(ctx context.Context, userID string) ([]Session, error) {
	var payload struct {
		Data []session `json:"data"`
	}

	query := url.Values{"user_id": {userID}, "limit": {fmt.Sprint(sessionLimit)}}

	if err := c.get(ctx, fmt.Sprintf("/api/sites/%d/sessions?%s", c.site, query.Encode()), &payload); err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	sessions := make([]Session, 0, len(payload.Data))

	for _, raw := range payload.Data {
		startedAt, err := time.Parse(time.DateTime, raw.SessionStart)

		if err != nil {
			return nil, fmt.Errorf("parse start of session %s: %w", raw.SessionID, err)
		}

		sessions = append(sessions, Session{
			ID:              raw.SessionID,
			StartedAt:       startedAt,
			Duration:        time.Duration(raw.SessionDuration) * time.Second,
			EntryPage:       raw.EntryPage,
			Pageviews:       raw.Pageviews,
			Country:         raw.Country,
			City:            raw.City,
			Language:        raw.Language,
			Browser:         raw.Browser,
			OperatingSystem: raw.OperatingSystem,
			DeviceType:      raw.DeviceType,
			Referrer:        raw.Referrer,
			HasReplay:       raw.HasReplay == 1,
		})
	}

	return sessions, nil
}

// Replay returns the recorded rrweb events of a session as a JSON array.
func (c *Client) Replay(ctx context.Context, sessionID string) (json.RawMessage, error) {
	var payload struct {
		Events json.RawMessage `json:"events"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/sites/%d/session-replay/%s", c.site, url.PathEscape(sessionID)), &payload); err != nil {
		return nil, fmt.Errorf("load replay of session %s: %w", sessionID, err)
	}

	return payload.Events, nil
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)

	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	response, err := c.http.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		var problem struct {
			Error    string `json:"error"`
			Message  string `json:"message"`
			Required string `json:"required"`
		}

		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))

		if json.Unmarshal(body, &problem) != nil || problem.Error+problem.Message == "" {
			return fmt.Errorf("%s: %s", response.Status, strings.TrimSpace(string(body)))
		}

		message := cmp.Or(problem.Message, problem.Error)

		if problem.Required != "" {
			message += ", the API key needs the " + problem.Required + " scope"
		}

		return fmt.Errorf("%s: %s", response.Status, message)
	}

	return json.NewDecoder(response.Body).Decode(out)
}

// UserURL is the user's page in the Rybbit dashboard, which lists their
// sessions and plays the recorded ones.
func (c *Client) UserURL(userID string) string {
	return fmt.Sprintf("%s/%d/user/%s", c.baseURL, c.site, url.PathEscape(userID))
}
