package upstream

import (
	"context"
	"errors"
	"strconv"

	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/mattermost"
	"github.com/zeitlos/lucity/pkg/telegram"
)

// ChatChannel is a chat or channel in a chat app, by the chat app's own ID.
type ChatChannel struct {
	ID   string
	Name string
}

// ChatBot calls a chat app as its bot and returns the bot's handle.
func (c *Client) ChatBot(ctx context.Context, kind, serverURL, token string) (string, Check) {
	switch kind {
	case chat.Telegram:
		me, err := telegram.New(token, c.http).Me(ctx)

		if err != nil {
			return "", chatCheck(err)
		}

		return me.Handle(), Check{Status: Valid}
	case chat.Mattermost:
		me, err := mattermost.New(serverURL, token, c.http).Me(ctx)

		if err != nil {
			return "", chatCheck(err)
		}

		return "@" + me.Username, Check{Status: Valid}
	default:
		return "", Check{Status: Rejected, Reason: "it is not a supported chat app"}
	}
}

// ChatChannels lists the channels a chat app's bot is a member of. Telegram
// can't list a bot's chats, so it has none.
func (c *Client) ChatChannels(ctx context.Context, kind, serverURL, token string) ([]ChatChannel, Check) {
	if kind != chat.Mattermost {
		return []ChatChannel{}, Check{Status: Valid}
	}

	channels, err := mattermost.New(serverURL, token, c.http).Channels(ctx)

	if err != nil {
		return nil, chatCheck(err)
	}

	result := make([]ChatChannel, 0, len(channels))

	for _, channel := range channels {
		result = append(result, ChatChannel{ID: channel.ID, Name: mattermostChannelName(channel)})
	}

	return result, Check{Status: Valid}
}

// ChatChannel looks up one chat or channel as the bot, which only works for
// one the bot is a member of.
func (c *Client) ChatChannel(ctx context.Context, kind, serverURL, token, id string) (ChatChannel, Check) {
	switch kind {
	case chat.Telegram:
		chatID, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			return ChatChannel{}, Check{Status: Rejected, Reason: "it is not a Telegram chat ID"}
		}

		found, err := telegram.New(token, c.http).Chat(ctx, chatID)

		if err != nil {
			return ChatChannel{}, chatCheck(err)
		}

		return ChatChannel{ID: id, Name: found.Name()}, Check{Status: Valid}
	case chat.Mattermost:
		found, err := mattermost.New(serverURL, token, c.http).Channel(ctx, id)

		if err != nil {
			return ChatChannel{}, chatCheck(err)
		}

		return ChatChannel{ID: id, Name: mattermostChannelName(found)}, Check{Status: Valid}
	default:
		return ChatChannel{}, Check{Status: Rejected, Reason: "it is not a supported chat app"}
	}
}

func mattermostChannelName(channel mattermost.Channel) string {
	if channel.DisplayName != "" {
		return channel.DisplayName
	}

	return channel.Name
}

func chatCheck(err error) Check {
	var telegramErr *telegram.Error
	var mattermostErr *mattermost.Error

	switch {
	case telegram.IsUnauthorized(err), mattermost.IsUnauthorized(err):
		return Check{Status: Rejected, Reason: "it turned the token down"}
	case errors.As(err, &telegramErr):
		return Check{Status: Rejected, Reason: "Telegram says: " + telegramErr.Description}
	case errors.As(err, &mattermostErr):
		return statusCheck(mattermostErr.StatusCode)
	default:
		return unreachable(err)
	}
}
