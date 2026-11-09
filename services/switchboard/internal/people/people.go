// Package people tells switchboard who wrote a chat message: the Lucity user
// a chat account is linked to, if that user is a member of the chat app's
// workspace. It also finishes the links people start in Lucity.
package people

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/chat"
	"github.com/zeitlos/lucity/pkg/logto"
)

const refreshAtLeast = 30 * time.Second

var (
	ErrInvalidCode = errors.New("that code isn't valid")
	ErrExpired     = errors.New("that link expired")
)

type Member struct {
	ID   string
	Name string
}

type Directory struct {
	logto *logto.Client

	mu         sync.Mutex
	workspaces map[string]*workspace
}

type workspace struct {
	orgID    string
	fetched  time.Time
	members  map[string]Member
	accounts map[account]string
}

type account struct {
	kind   string
	server string
	id     string
}

func New(logto *logto.Client) *Directory {
	return &Directory{logto: logto, workspaces: map[string]*workspace{}}
}

// Find returns the member of workspace whose linked account is id on kind and
// server, or nil. An account two members claim counts as nobody's.
func (d *Directory) Find(ctx context.Context, workspaceName, kind, server, id string) (*Member, error) {
	key := account{kind: kind, server: server, id: id}

	return d.lookup(ctx, workspaceName, func(ws *workspace) (Member, bool) {
		userID, ok := ws.accounts[key]

		if !ok || userID == "" {
			return Member{}, false
		}

		member, ok := ws.members[userID]

		return member, ok
	})
}

// Member returns a member of workspace by user ID, or nil.
func (d *Directory) Member(ctx context.Context, workspaceName, userID string) (*Member, error) {
	return d.lookup(ctx, workspaceName, func(ws *workspace) (Member, bool) {
		member, ok := ws.members[userID]

		return member, ok
	})
}

func (d *Directory) lookup(ctx context.Context, name string, find func(*workspace) (Member, bool)) (*Member, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	ws, ok := d.workspaces[name]

	if ok {
		if member, found := find(ws); found {
			return &member, nil
		}

		if time.Since(ws.fetched) < refreshAtLeast {
			return nil, nil
		}
	}

	ws, err := d.fetch(ctx, name, ws)

	if err != nil {
		return nil, err
	}

	d.workspaces[name] = ws

	if member, found := find(ws); found {
		return &member, nil
	}

	return nil, nil
}

func (d *Directory) fetch(ctx context.Context, name string, previous *workspace) (*workspace, error) {
	orgID := ""

	if previous != nil {
		orgID = previous.orgID
	}

	if orgID == "" {
		org, err := d.logto.OrganizationByName(ctx, name)

		if err != nil {
			return nil, err
		}

		orgID = org.ID
	}

	members, err := d.logto.OrganizationMembers(ctx, orgID)

	if err != nil {
		return nil, err
	}

	ws := &workspace{orgID: orgID, fetched: time.Now(), members: map[string]Member{}, accounts: map[account]string{}}

	for _, member := range members {
		customData := member.CustomData

		if customData == nil {
			user, err := d.logto.User(ctx, member.ID)

			if err != nil {
				return nil, err
			}

			customData = user.CustomData
		}

		displayName := member.Name

		if displayName == "" {
			displayName = member.Email
		}

		ws.members[member.ID] = Member{ID: member.ID, Name: displayName}

		for _, linked := range chat.Accounts(customData) {
			key := account{kind: linked.Kind, server: linked.Server, id: linked.ID}

			if _, taken := ws.accounts[key]; taken {
				ws.accounts[key] = ""
				continue
			}

			ws.accounts[key] = member.ID
		}
	}

	return ws, nil
}

// Link finishes a link: code came from the chat account linked, and names the
// user who started it in Lucity. It returns that user's name.
func (d *Directory) Link(ctx context.Context, code string, linked chat.Account) (string, error) {
	userID, nonce, ok := chat.ParseLinkCode(code)

	if !ok {
		return "", ErrInvalidCode
	}

	user, err := d.logto.User(ctx, userID)

	if err != nil {
		return "", ErrInvalidCode
	}

	pending := chat.Pending(user.CustomData)

	if pending == nil || pending.Nonce != nonce {
		return "", ErrInvalidCode
	}

	if time.Now().After(pending.ExpiresAt) {
		return "", ErrExpired
	}

	linked.LinkedAt = time.Now().UTC()

	if err := d.logto.UpdateUserCustomData(ctx, userID, chat.WithAccount(user.CustomData, linked)); err != nil {
		return "", err
	}

	d.mu.Lock()
	clear(d.workspaces)
	d.mu.Unlock()

	if user.Name != "" {
		return user.Name, nil
	}

	return user.PrimaryEmail, nil
}
