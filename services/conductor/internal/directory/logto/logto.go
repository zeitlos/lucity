package logto

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/zeitlos/lucity/pkg/logto"
)

type Provider struct {
	api *logto.Client

	adminRoleID  string
	memberRoleID string

	orgIDCache   map[string]string
	orgIDCacheMu sync.RWMutex
}

func New(client *logto.Client) (*Provider, error) {
	ctx := context.Background()

	roles, err := client.OrganizationRoles(ctx)

	if err != nil {
		return nil, err
	}

	p := Provider{
		api: client,
	}

	for _, r := range roles {
		switch r.Name {
		case "admin":
			p.adminRoleID = r.ID
		case "member":
			p.memberRoleID = r.ID
		}
	}

	if p.adminRoleID == "" || p.memberRoleID == "" {
		return nil, fmt.Errorf("missing org roles: admin=%q member=%q", p.adminRoleID, p.memberRoleID)
	}

	slog.Info("logto org roles cached", "admin", p.adminRoleID, "member", p.memberRoleID)

	return &p, nil
}
