package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	pkglabels "github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	barmanPluginName     = "barman-cloud.cloudnative-pg.io"
	scheduledBackupLabel = "cnpg.io/scheduled-backup"
)

var (
	cnpgBackupGVR          = cnpgv1.SchemeGroupVersion.WithResource("backups")
	cnpgScheduledBackupGVR = cnpgv1.SchemeGroupVersion.WithResource("scheduledbackups")
)

func (c *Client) DatabaseBackups(ctx context.Context, id platform.DatabaseID) (*platform.DatabaseBackups, error) {
	cluster, err := c.cluster(ctx, id)

	if err != nil {
		return nil, err
	}

	clusterName := cluster.Name
	namespace := id.Namespace()

	list, err := c.dynamic.Resource(cnpgBackupGVR).Namespace(namespace).List(ctx, meta.ListOptions{})

	if err != nil {
		return nil, err
	}

	backups := make([]platform.DatabaseBackup, 0, len(list.Items))

	for _, item := range list.Items {
		backup, err := toBackup(item)

		if err != nil {
			return nil, err
		}

		if backup.Spec.Cluster.Name != clusterName {
			continue
		}

		backups = append(backups, toDatabaseBackup(*backup))
	}

	sort.Slice(backups, func(i, j int) bool {
		return backupOrder(backups[i]).After(backupOrder(backups[j]))
	})

	result := platform.DatabaseBackups{
		Enabled:              len(cluster.Spec.Plugins) > 0,
		ArchivingHealthy:     archivingHealthy(*cluster),
		ServerName:           serverNameOf(*cluster),
		Backups:              backups,
		EarliestRestorePoint: earliestRestorePoint(*cluster, backups),
		LastBackupAt:         lastBackupAt(backups),
	}

	schedule, err := c.scheduleFor(ctx, namespace, clusterName)

	if err != nil {
		return nil, err
	}

	result.Schedule = schedule

	return &result, nil
}

func (c *Client) CreateDatabaseBackup(ctx context.Context, id platform.DatabaseID) (*platform.DatabaseBackup, error) {
	cluster, err := c.cluster(ctx, id)

	if err != nil {
		return nil, err
	}

	if len(cluster.Spec.Plugins) == 0 {
		return nil, fmt.Errorf("database %q has no backups configured", id.Name)
	}

	namespace := id.Namespace()

	existing, err := c.DatabaseBackups(ctx, id)

	if err != nil {
		return nil, err
	}

	for _, backup := range existing.Backups {
		if backup.Status == platform.BackupRunning || backup.Status == platform.BackupPending {
			return &backup, nil
		}
	}

	backup := &cnpgv1.Backup{
		TypeMeta: meta.TypeMeta{
			APIVersion: cnpgv1.SchemeGroupVersion.String(),
			Kind:       "Backup",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      cluster.Name + "-" + rand.String(6),
			Namespace: namespace,
			Labels: map[string]string{
				databaseLabel:       id.Name,
				pkglabels.ManagedBy: pkglabels.ManagedByLucity,
			},
		},
		Spec: cnpgv1.BackupSpec{
			Cluster: cnpgv1.LocalObjectReference{Name: cluster.Name},
			Method:  cnpgv1.BackupMethodPlugin,
			PluginConfiguration: &cnpgv1.BackupPluginConfiguration{
				Name: barmanPluginName,
			},
		},
	}

	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(backup)

	if err != nil {
		return nil, err
	}

	created, err := c.dynamic.Resource(cnpgBackupGVR).Namespace(namespace).Create(ctx, &unstructured.Unstructured{Object: object}, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	result, err := toBackup(*created)

	if err != nil {
		return nil, err
	}

	return new(toDatabaseBackup(*result)), nil
}

func (c *Client) scheduleFor(ctx context.Context, namespace, clusterName string) (string, error) {
	list, err := c.dynamic.Resource(cnpgScheduledBackupGVR).Namespace(namespace).List(ctx, meta.ListOptions{})

	if err != nil {
		return "", err
	}

	for _, item := range list.Items {
		var scheduled cnpgv1.ScheduledBackup

		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &scheduled); err != nil {
			return "", err
		}

		if scheduled.Spec.Cluster.Name == clusterName {
			return scheduled.Spec.Schedule, nil
		}
	}

	return "", nil
}

func archivingHealthy(cluster cnpgv1.Cluster) bool {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == "ContinuousArchiving" {
			return condition.Status == "True"
		}
	}

	return false
}

func serverNameOf(cluster cnpgv1.Cluster) string {
	for _, plugin := range cluster.Spec.Plugins {
		if plugin.Name == barmanPluginName {
			return plugin.Parameters["serverName"]
		}
	}

	return ""
}

func toBackup(item unstructured.Unstructured) (*cnpgv1.Backup, error) {
	var backup cnpgv1.Backup
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &backup)

	return &backup, err
}

func toDatabaseBackup(backup cnpgv1.Backup) platform.DatabaseBackup {
	result := platform.DatabaseBackup{
		Name:      backup.Name,
		CreatedAt: backup.CreationTimestamp.Time,
		Status:    backupStatus(backup),
		Trigger:   backupTrigger(backup),
		Error:     backup.Status.Error,
	}

	if started := backup.Status.StartedAt; started != nil {
		result.StartedAt = &started.Time
	}

	if stopped := backup.Status.StoppedAt; stopped != nil {
		result.FinishedAt = &stopped.Time
	}

	return result
}

func backupStatus(backup cnpgv1.Backup) platform.BackupStatus {
	switch backup.Status.Phase {
	case cnpgv1.BackupPhaseCompleted:
		return platform.BackupCompleted
	case cnpgv1.BackupPhaseFailed, cnpgv1.BackupPhaseWalArchivingFailing:
		return platform.BackupFailed
	case cnpgv1.BackupPhaseRunning, cnpgv1.BackupPhaseFinalizing, cnpgv1.BackupPhaseStarted:
		return platform.BackupRunning
	case cnpgv1.BackupPhasePending, "":
		return platform.BackupPending
	}

	slog.Warn("unknown backup phase", "phase", backup.Status.Phase, "backup", backup.Name)

	return platform.BackupFailed
}

func backupTrigger(backup cnpgv1.Backup) platform.BackupTrigger {
	if _, ok := backup.Labels[scheduledBackupLabel]; ok {
		return platform.BackupScheduled
	}

	return platform.BackupManual
}

func earliestRestorePoint(cluster cnpgv1.Cluster, backups []platform.DatabaseBackup) *time.Time {
	if point := cluster.Status.FirstRecoverabilityPoint; point != "" {
		if parsed, err := time.Parse(time.RFC3339, point); err == nil {
			return &parsed
		}

		slog.Warn("unparseable first recoverability point", "value", point, "cluster", cluster.Name)
	}

	var earliest *time.Time

	for _, backup := range backups {
		if backup.Status != platform.BackupCompleted || backup.FinishedAt == nil {
			continue
		}

		if earliest == nil || backup.FinishedAt.Before(*earliest) {
			earliest = backup.FinishedAt
		}
	}

	return earliest
}

func lastBackupAt(backups []platform.DatabaseBackup) *time.Time {
	for _, backup := range backups {
		if backup.Status == platform.BackupCompleted && backup.FinishedAt != nil {
			return backup.FinishedAt
		}
	}

	return nil
}

func backupOrder(backup platform.DatabaseBackup) time.Time {
	if backup.StartedAt != nil {
		return *backup.StartedAt
	}

	return backup.CreatedAt
}
