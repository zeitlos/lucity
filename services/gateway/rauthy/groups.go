package rauthy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Group represents a Rauthy group.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Groups returns all Rauthy groups.
func (c *Client) Groups(ctx context.Context) ([]Group, error) {
	var groups []Group
	if err := c.doJSON(ctx, "GET", "/groups", nil, &groups); err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	return groups, nil
}

// CreateGroup creates a new Rauthy group and returns it.
// Idempotent: if the group already exists, returns the existing group.
func (c *Client) CreateGroup(ctx context.Context, name string) (*Group, error) {
	payload, _ := json.Marshal(map[string]string{"group": name})

	var group Group
	if err := c.doJSON(ctx, "POST", "/groups", bytes.NewReader(payload), &group); err != nil {
		// Handle "already exists" — return the existing group
		if strings.Contains(err.Error(), "already exists") {
			existing, findErr := c.GroupByName(ctx, name)
			if findErr != nil {
				return nil, fmt.Errorf("group %q already exists but failed to look up: %w", name, findErr)
			}
			if existing != nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("failed to create group %q: %w", name, err)
	}
	return &group, nil
}

// DeleteGroup deletes a Rauthy group by ID.
// This also removes the group from all assigned users.
func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	if err := c.doNoContent(ctx, "DELETE", "/groups/"+id, nil); err != nil {
		return fmt.Errorf("failed to delete group %q: %w", id, err)
	}
	return nil
}

// FindGroupsByPrefix returns groups whose name starts with prefix.
func (c *Client) FindGroupsByPrefix(ctx context.Context, prefix string) ([]Group, error) {
	all, err := c.Groups(ctx)
	if err != nil {
		return nil, err
	}

	var matched []Group
	for _, g := range all {
		if strings.HasPrefix(g.Name, prefix) {
			matched = append(matched, g)
		}
	}
	return matched, nil
}

// GroupByName finds a group by its exact name. Returns nil if not found.
func (c *Client) GroupByName(ctx context.Context, name string) (*Group, error) {
	all, err := c.Groups(ctx)
	if err != nil {
		return nil, err
	}

	for _, g := range all {
		if g.Name == name {
			return &g, nil
		}
	}
	return nil, nil
}
