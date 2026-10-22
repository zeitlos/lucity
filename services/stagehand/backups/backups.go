// Package backups gives every archiving database a recovery floor. A scheduled
// backup fires the moment archiving is switched on, before the pods carrying
// the backup sidecar have rolled, so the first one always fails and the next is
// a week out. Until a base backup exists there is nothing to recover to,
// however much write-ahead log has been archived.
package backups

import (
	"context"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/pkg/platform"
)

const retryInterval = time.Hour

func Reconcile(ctx context.Context, client platform.Interface, environments []platform.EnvironmentID) error {
	for _, environment := range environments {
		if err := ctx.Err(); err != nil {
			return err
		}

		databases, err := client.Databases(ctx, environment)

		if err != nil {
			slog.Warn("reconcile backups: failed to list databases", "error", err, "environment", environment)
			continue
		}

		for _, database := range databases {
			if database.Status != platform.DatabaseHealthy {
				continue
			}

			if err := ensureRecoveryFloor(ctx, client, database.ID); err != nil {
				slog.Warn("reconcile backups: failed to ensure recovery floor", "error", err, "database", database.ID)
			}
		}
	}

	return nil
}

func ensureRecoveryFloor(ctx context.Context, client platform.Interface, id platform.DatabaseID) error {
	backups, err := client.DatabaseBackups(ctx, id)

	if err != nil {
		return err
	}

	if !backups.Enabled || backups.EarliestRestorePoint != nil {
		return nil
	}

	for _, backup := range backups.Backups {
		if backup.Status == platform.BackupRunning || backup.Status == platform.BackupPending {
			return nil
		}
	}

	// Sorted newest first, so this bounds retries when backups fail for a reason
	// retrying will not fix.
	if len(backups.Backups) > 0 && time.Since(backups.Backups[0].CreatedAt) < retryInterval {
		return nil
	}

	slog.Info("reconcile backups: no recovery point, starting a backup", "database", id)

	_, err = client.CreateDatabaseBackup(ctx, id)

	return err
}
