package edge

import "context"

type Interface interface {
	Register(ctx context.Context, workspaceID string, host Host) error
	Unregister(ctx context.Context, workspaceID, name string) error
	DeleteZone(ctx context.Context, workspaceID string) error
	Sync(ctx context.Context, workspaces map[string][]Host, routing []Host) error
	EdgeAddresses(ctx context.Context) ([]string, error)
}

type Disabled struct{}

func (Disabled) Register(context.Context, string, Host) error          { return nil }
func (Disabled) Unregister(context.Context, string, string) error      { return nil }
func (Disabled) DeleteZone(context.Context, string) error              { return nil }
func (Disabled) Sync(context.Context, map[string][]Host, []Host) error { return nil }
func (Disabled) EdgeAddresses(context.Context) ([]string, error)       { return nil, nil }
