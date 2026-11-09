// Package telegram is a small client for the Telegram Bot API: the calls the
// platform makes as a workspace's bot, and nothing else.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiURL = "https://api.telegram.org"

// PollTimeout is how long one getUpdates call waits for an update. The
// http.Client given to New must allow a little more than that.
const PollTimeout = 50 * time.Second

type Client struct {
	token string
	http  *http.Client
}

func New(token string, httpClient *http.Client) *Client {
	return &Client{token: token, http: httpClient}
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// Handle is the user's @name, or the first name for a user without one.
func (u User) Handle() string {
	if u.Username != "" {
		return "@" + u.Username
	}

	return u.FirstName
}

type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	IsForum   bool   `json:"is_forum"`
}

// Name is how the chat reads in Lucity: a group's title, or the person a
// private chat is with.
func (c Chat) Name() string {
	if c.Title != "" {
		return c.Title
	}

	if c.Username != "" {
		return "@" + c.Username
	}

	return c.FirstName
}

type Message struct {
	MessageID       int64    `json:"message_id"`
	MessageThreadID int64    `json:"message_thread_id"`
	IsTopicMessage  bool     `json:"is_topic_message"`
	Date            int64    `json:"date"`
	From            *User    `json:"from"`
	Chat            Chat     `json:"chat"`
	Text            string   `json:"text"`
	ReplyToMessage  *Message `json:"reply_to_message"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

// Button is an inline keyboard button. It either sends CallbackData back to
// the bot or opens URL.
type Button struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

type Outgoing struct {
	ChatID   int64
	ThreadID int64
	ReplyTo  int64
	Text     string
	HTML     bool
	Buttons  []Button
}

type Error struct {
	Method      string
	Code        int
	Description string
	RetryAfter  time.Duration
}

func (e *Error) Error() string {
	return fmt.Sprintf("telegram %s: %s", e.Method, e.Description)
}

// IsUnauthorized reports whether Telegram turned the bot token down.
func IsUnauthorized(err error) bool {
	var apiErr *Error

	return errors.As(err, &apiErr) && apiErr.Code == http.StatusUnauthorized
}

func (c *Client) Me(ctx context.Context) (User, error) {
	var me User

	err := c.call(ctx, "getMe", map[string]any{}, &me)

	return me, err
}

// Chat returns a chat the bot is in. Telegram answers with an error for a
// chat the bot isn't a member of.
func (c *Client) Chat(ctx context.Context, chatID int64) (Chat, error) {
	var chat Chat

	err := c.call(ctx, "getChat", map[string]any{"chat_id": chatID}, &chat)

	return chat, err
}

func (c *Client) Updates(ctx context.Context, offset int64) ([]Update, error) {
	var updates []Update

	err := c.call(ctx, "getUpdates", map[string]any{
		"offset":          offset,
		"timeout":         int(PollTimeout.Seconds()),
		"allowed_updates": []string{"message", "callback_query"},
	}, &updates)

	return updates, err
}

func (c *Client) Send(ctx context.Context, message Outgoing) (int64, error) {
	params := message.params()

	if message.ThreadID != 0 {
		params["message_thread_id"] = message.ThreadID
	}

	if message.ReplyTo != 0 {
		params["reply_parameters"] = map[string]any{"message_id": message.ReplyTo, "allow_sending_without_reply": true}
	}

	var sent Message

	err := c.call(ctx, "sendMessage", params, &sent)

	return sent.MessageID, err
}

func (c *Client) Edit(ctx context.Context, messageID int64, message Outgoing) error {
	params := message.params()
	params["message_id"] = messageID

	err := c.call(ctx, "editMessageText", params, nil)

	var apiErr *Error

	if errors.As(err, &apiErr) && strings.Contains(apiErr.Description, "message is not modified") {
		return nil
	}

	return err
}

func (c *Client) Delete(ctx context.Context, chatID, messageID int64) error {
	return c.call(ctx, "deleteMessage", map[string]any{"chat_id": chatID, "message_id": messageID}, nil)
}

func (c *Client) Answer(ctx context.Context, callbackID, text string) error {
	return c.call(ctx, "answerCallbackQuery", map[string]any{"callback_query_id": callbackID, "text": text}, nil)
}

func (c *Client) Typing(ctx context.Context, chatID, threadID int64) error {
	params := map[string]any{"chat_id": chatID, "action": "typing"}

	if threadID != 0 {
		params["message_thread_id"] = threadID
	}

	return c.call(ctx, "sendChatAction", params, nil)
}

func (m Outgoing) params() map[string]any {
	params := map[string]any{
		"chat_id":              m.ChatID,
		"text":                 m.Text,
		"link_preview_options": map[string]bool{"is_disabled": true},
	}

	if m.HTML {
		params["parse_mode"] = "HTML"
	}

	if len(m.Buttons) > 0 {
		params["reply_markup"] = map[string]any{"inline_keyboard": [][]Button{m.Buttons}}
	}

	return params
}

func (c *Client) call(ctx context.Context, method string, params, result any) error {
	body, err := json.Marshal(params)

	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/bot"+c.token+"/"+method, bytes.NewReader(body))

	if err != nil {
		return fmt.Errorf("telegram %s: build request failed", method)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)

	if err != nil {
		var urlErr *url.Error

		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}

		return fmt.Errorf("telegram %s: %w", method, err)
	}

	defer response.Body.Close()

	var envelope struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		ErrorCode   int             `json:"error_code"`
		Description string          `json:"description"`
		Parameters  struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}

	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("telegram %s: decode response: %w", method, err)
	}

	if !envelope.OK {
		return &Error{
			Method:      method,
			Code:        envelope.ErrorCode,
			Description: envelope.Description,
			RetryAfter:  time.Duration(envelope.Parameters.RetryAfter) * time.Second,
		}
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(envelope.Result, result)
}
