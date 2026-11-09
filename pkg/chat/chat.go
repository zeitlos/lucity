// Package chat holds what the conductor and switchboard share about chat apps:
// their kinds, the chat accounts people link to their Lucity user, and the code
// that proves a chat account belongs to that user.
package chat

import (
	"encoding/json"
	"strings"
	"time"
)

const (
	Telegram   = "telegram"
	Mattermost = "mattermost"
)

// LinkTTL is how long a started link waits for its code to arrive.
const LinkTTL = 10 * time.Minute

const (
	accountsKey = "chatAccounts"
	pendingKey  = "pendingChatLink"
)

// Account is a chat identity that counts as the Lucity user it is linked to.
// Server is the Mattermost server it lives on, empty for Telegram.
type Account struct {
	Kind     string    `json:"kind"`
	Server   string    `json:"server,omitempty"`
	ID       string    `json:"id"`
	Handle   string    `json:"handle"`
	LinkedAt time.Time `json:"linkedAt"`
}

// PendingLink is a link someone started in Lucity and hasn't finished in the
// chat app yet.
type PendingLink struct {
	Nonce     string    `json:"nonce"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Accounts reads the linked accounts from a user's custom data.
func Accounts(customData map[string]any) []Account {
	var accounts []Account

	decode(customData[accountsKey], &accounts)

	return accounts
}

// Pending reads the started link from a user's custom data, if any.
func Pending(customData map[string]any) *PendingLink {
	var pending PendingLink

	if !decode(customData[pendingKey], &pending) || pending.Nonce == "" {
		return nil
	}

	return &pending
}

// WithAccount returns the custom data changes that link account, replacing an
// earlier account of the same kind on the same server.
func WithAccount(customData map[string]any, account Account) map[string]any {
	accounts := []Account{account}

	for _, existing := range Accounts(customData) {
		if existing.Kind != account.Kind || existing.Server != account.Server {
			accounts = append(accounts, existing)
		}
	}

	return map[string]any{accountsKey: accounts, pendingKey: nil}
}

// WithoutAccount returns the custom data changes that remove the account of
// kind on server.
func WithoutAccount(customData map[string]any, kind, server string) map[string]any {
	accounts := []Account{}

	for _, existing := range Accounts(customData) {
		if existing.Kind != kind || existing.Server != server {
			accounts = append(accounts, existing)
		}
	}

	return map[string]any{accountsKey: accounts}
}

// WithPending returns the custom data changes that start a link.
func WithPending(pending PendingLink) map[string]any {
	return map[string]any{pendingKey: pending}
}

// LinkCode is what a person sends to a bot to link their chat account: their
// Lucity user ID and the nonce of the link they started. It only uses
// characters a Telegram start parameter allows.
func LinkCode(userID, nonce string) string {
	return userID + "_" + nonce
}

// ParseLinkCode splits a code made by LinkCode.
func ParseLinkCode(code string) (userID, nonce string, ok bool) {
	i := strings.LastIndex(code, "_")

	if i <= 0 || i == len(code)-1 {
		return "", "", false
	}

	return code[:i], code[i+1:], true
}

func decode(value any, target any) bool {
	if value == nil {
		return false
	}

	data, err := json.Marshal(value)

	if err != nil {
		return false
	}

	return json.Unmarshal(data, target) == nil
}
