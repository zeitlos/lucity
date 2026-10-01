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

type Client struct {
	token string
	http  *http.Client
}

func New(token string) *Client {
	return &Client{token: token, http: &http.Client{Timeout: 70 * time.Second}}
}

type User struct {
	ID       int64  `json:"id"`
	IsBot    bool   `json:"is_bot"`
	Username string `json:"username"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type ForumTopic struct {
	Name string `json:"name"`
}

type Message struct {
	MessageID         int64       `json:"message_id"`
	MessageThreadID   int64       `json:"message_thread_id"`
	IsTopicMessage    bool        `json:"is_topic_message"`
	Date              int64       `json:"date"`
	From              *User       `json:"from"`
	Chat              Chat        `json:"chat"`
	Text              string      `json:"text"`
	ReplyToMessage    *Message    `json:"reply_to_message"`
	ForumTopicCreated *ForumTopic `json:"forum_topic_created"`
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

type Button struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
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
	Description string
	RetryAfter  time.Duration
}

func (e *Error) Error() string {
	return fmt.Sprintf("telegram %s: %s", e.Method, e.Description)
}

func (c *Client) Me(ctx context.Context) (User, error) {
	var me User
	err := c.call(ctx, "getMe", map[string]any{}, &me)
	return me, err
}

func (c *Client) Updates(ctx context.Context, offset int64) ([]Update, error) {
	var updates []Update
	err := c.call(ctx, "getUpdates", map[string]any{
		"offset":          offset,
		"timeout":         50,
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

func (c *Client) CreateTopic(ctx context.Context, chatID int64, name string) (int64, error) {
	var topic struct {
		MessageThreadID int64 `json:"message_thread_id"`
	}
	err := c.call(ctx, "createForumTopic", map[string]any{"chat_id": chatID, "name": name}, &topic)
	return topic.MessageThreadID, err
}

func (c *Client) Copy(ctx context.Context, chatID, threadID int64, from *Message) (int64, error) {
	return c.relay(ctx, "copyMessage", chatID, threadID, from)
}

func (c *Client) Forward(ctx context.Context, chatID, threadID int64, from *Message) (int64, error) {
	return c.relay(ctx, "forwardMessage", chatID, threadID, from)
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

func (c *Client) relay(ctx context.Context, method string, chatID, threadID int64, from *Message) (int64, error) {
	params := map[string]any{"chat_id": chatID, "from_chat_id": from.Chat.ID, "message_id": from.MessageID}
	if threadID != 0 {
		params["message_thread_id"] = threadID
	}
	var result struct {
		MessageID int64 `json:"message_id"`
	}
	err := c.call(ctx, method, params, &result)
	return result.MessageID, err
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
			Description: envelope.Description,
			RetryAfter:  time.Duration(envelope.Parameters.RetryAfter) * time.Second,
		}
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(envelope.Result, result)
}
