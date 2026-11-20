package cockpit

import (
	"context"
	"fmt"
	"log/slog"
)

// MarkAbuse flags each workspace as abuse with the note, so stagehand
// suspends it on its next pass.
func (c *Client) MarkAbuse(ctx context.Context, workspaceIDs []string, note string) []Result {
	return c.updateEach(ctx, workspaceIDs, func(data map[string]interface{}) {
		data[abuseKey] = true

		if note != "" {
			data[abuseNoteKey] = note
		} else {
			delete(data, abuseNoteKey)
		}
	})
}

// ClearAbuse removes the abuse flag and note from each workspace. Stagehand
// then resumes the ones in good standing.
func (c *Client) ClearAbuse(ctx context.Context, workspaceIDs []string) []Result {
	return c.updateEach(ctx, workspaceIDs, func(data map[string]interface{}) {
		delete(data, abuseKey)
		delete(data, abuseNoteKey)
	})
}

func (c *Client) updateEach(ctx context.Context, workspaceIDs []string, change func(map[string]interface{})) []Result {
	results := make([]Result, 0, len(workspaceIDs))

	for _, workspaceID := range workspaceIDs {
		err := c.updateCustomData(ctx, workspaceID, change)

		if err != nil {
			slog.Error("update workspace", "workspace", workspaceID, "error", err)
		} else {
			slog.Info("updated workspace", "workspace", workspaceID)
		}

		results = append(results, Result{WorkspaceID: workspaceID, Err: err})
	}

	return results
}

func (c *Client) updateCustomData(ctx context.Context, workspaceID string, change func(map[string]interface{})) error {
	organization, err := c.identity.OrganizationByName(ctx, workspaceID)

	if err != nil {
		return fmt.Errorf("find workspace: %w", err)
	}

	data := organization.CustomData

	if data == nil {
		data = map[string]interface{}{}
	}

	change(data)

	if _, err := c.identity.UpdateOrganizationCustomData(ctx, organization.ID, data); err != nil {
		return fmt.Errorf("update workspace: %w", err)
	}

	return nil
}
