<script setup lang="ts">
import NameIdField from '@/components/NameIdField.vue';
import { isValidId } from '@/lib/id';
import { useResourceIds } from '@/composables/useResourceIds';
import QueryError from '@/components/QueryError.vue';
import { ref, computed, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { useIntervalFn, useNow } from '@vueuse/core';
import { CalendarClock, MousePointerClick, Timer, DatabaseBackup } from '@lucide/vue';
import { graphql } from '@/gql';
import { BackupStatus, BackupTrigger } from '@/gql/graphql';

const DatabaseBackupsDocument = graphql(`
  query DatabaseBackups($database: DatabaseID!) {
    database(id: $database) {
      id
      name
      backups {
        enabled
        retentionDays
        earliestRestorePoint
        latestRestorePoint
        lastBackupAt
        backups {
          id
          status
          trigger
          createdAt
          startedAt
          finishedAt
          error
        }
      }
    }
  }
`);

const CreateDatabaseBackupDocument = graphql(`
  mutation CreateDatabaseBackup($database: DatabaseID!) {
    createDatabaseBackup(database: $database) {
      id
      status
    }
  }
`);

const RestoreDatabaseDocument = graphql(`
  mutation RestoreDatabase($input: RestoreDatabaseInput!) {
    restoreDatabase(input: $input) {
      clampedToLatest
      database {
        id
        name
        status
      }
    }
  }
`);

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { failureToast, successToast } from '@/components/ui/sonner';
import EmptyState from '@/components/EmptyState.vue';
import RecoveryTimeline from './RecoveryTimeline.vue';
import { Status } from '@/components/ui/status';
import { backupStatus } from '@/lib/status';
import { formatDateTime, formatDuration, formatTime } from '@/lib/format';

const props = defineProps<{
  databaseId: string;
  databaseName: string;
}>();

const emit = defineEmits<{ (e: 'database-restored'): void }>();

const now = useNow({ scheduler: (cb) => useIntervalFn(cb, 30_000) });
const polling = ref(false);

const { result, loading, error, refetch } = useQuery(
  DatabaseBackupsDocument,
  () => ({ database: props.databaseId }),
  () => ({ enabled: !!props.databaseId, pollInterval: polling.value ? 5000 : 0 }),
);

const backups = computed(() => result.value?.database.backups ?? null);
const entries = computed(() => backups.value?.backups ?? []);

watch(entries, list => {
  polling.value = list.some(
    b => b.status === BackupStatus.Running || b.status === BackupStatus.Pending,
  );
});

const earliest = computed(() => {
  const point = backups.value?.earliestRestorePoint;
  return point ? new Date(point).getTime() : null;
});

const archiveFrontier = computed(() => {
  const point = backups.value?.latestRestorePoint;
  return point ? new Date(point).getTime() : null;
});

const latest = computed(() => archiveFrontier.value ?? now.value.getTime());

const markers = computed(() =>
  entries.value
    .filter(b => b.status === BackupStatus.Completed && b.finishedAt)
    .map(b => new Date(b.finishedAt as string).getTime()),
);

// Three shapes: no floor yet, a real window to scrub, or a window whose only
// reachable point is the newest one because nothing was written after it.
const degenerate = computed(
  () => earliest.value !== null && archiveFrontier.value !== null && archiveFrontier.value <= earliest.value,
);

const mode = ref<'moment' | 'latest'>('moment');

watch(degenerate, isDegenerate => {
  if (isDegenerate) mode.value = 'latest';
}, { immediate: true });
const hasWindow = computed(() => earliest.value !== null && earliest.value < latest.value);

const selected = ref<number | null>(null);

const targetInput = computed({
  get: () => (selected.value === null ? '' : toLocalInput(selected.value)),
  set: (value: string) => {
    const parsed = value ? new Date(value).getTime() : NaN;
    if (!Number.isNaN(parsed)) selected.value = parsed;
  },
});

function toLocalInput(time: number): string {
  const date = new Date(time - new Date(time).getTimezoneOffset() * 60_000);
  return date.toISOString().slice(0, 19);
}

const inputMin = computed(() => (earliest.value === null ? undefined : toLocalInput(earliest.value)));
const inputMax = computed(() => toLocalInput(latest.value));

const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

const restoreName = ref('');
const restoreId = ref('');
const takenIds = useResourceIds(() => props.databaseId.split('/').slice(0, 3).join('/'));

watch(
  () => props.databaseName,
  name => {
    restoreName.value = `${name} restored`;
  },
  { immediate: true },
);

const nameValid = computed(
  () => restoreName.value.trim().length > 0 && isValidId(restoreId.value) && !takenIds.value.has(restoreId.value),
);

const { mutate: createBackup, loading: backingUp } = useMutation(CreateDatabaseBackupDocument);
const { mutate: restoreMutate } = useMutation(RestoreDatabaseDocument);
const restoring = ref(false);

const canRestore = computed(() => {
  if (!nameValid.value || restoring.value) return false;
  if (mode.value === 'latest') return archiveFrontier.value !== null;
  return hasWindow.value && selected.value !== null;
});

async function handleBackupNow() {
  try {
    await createBackup({ database: props.databaseId });
    await refetch();
  } catch (e: unknown) {
    failureToast('Failed to start backup', e);
  }
}

async function handleRestore() {
  const targetTime = mode.value === 'latest' || selected.value === null
    ? null
    : new Date(selected.value).toISOString();

  if (mode.value === 'moment' && targetTime === null) return;

  restoring.value = true;
  try {
    const response = await restoreMutate({
      input: {
        database: props.databaseId,
        name: restoreName.value.trim(),
        id: restoreId.value,
        targetTime,
      },
    });

    const clamped = response?.data?.restoreDatabase.clampedToLatest;

    successToast(`Restoring into "${restoreName.value.trim()}"`, {
      description: clamped
        ? 'Nothing was written after the most recent archived point, so it restores to there.'
        : 'It appears on the canvas as soon as the data is back.',
    });
    emit('database-restored');
  } catch (e: unknown) {
    failureToast('Failed to restore', e);
  } finally {
    restoring.value = false;
  }
}

function duration(entry: { startedAt?: string | null; finishedAt?: string | null }): string {
  if (!entry.startedAt || !entry.finishedAt) return '';
  return formatDuration(new Date(entry.finishedAt).getTime() - new Date(entry.startedAt).getTime());
}

const fullBackupSchedule = (() => {
  const now = new Date();
  const sunday = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() - now.getUTCDay(), 2));
  return `${sunday.toLocaleDateString('en-US', { weekday: 'long' })} at ${formatTime(sunday)}`;
})();

const hasRunning = computed(() => entries.value.some(entry => entry.status === BackupStatus.Running));
const { now: clock, pause: pauseClock, resume: resumeClock } = useNow({ scheduler: cb => useIntervalFn(cb, 1000), controls: true });
watch(hasRunning, running => (running ? resumeClock() : pauseClock()), { immediate: true });

function statusText(entry: { status: BackupStatus; startedAt?: string | null }): string {
  const { label } = backupStatus(entry.status);
  if (entry.status !== BackupStatus.Running || !entry.startedAt) return label;
  return `${label} ${formatDuration(clock.value.getTime() - new Date(entry.startedAt).getTime())}`;
}

const showAll = ref(false);
const visibleEntries = computed(() => (showAll.value ? entries.value : entries.value.slice(0, 5)));
</script>

<template>
  <div class="space-y-8">
    <template v-if="loading && !backups">
      <section class="space-y-4">
        <div class="space-y-1">
          <h3 class="text-sm font-medium text-muted-foreground">Restore to a point in time</h3>
          <div v-for="lines in [['w-full max-w-lg', 'w-40'], ['w-full max-w-md', 'w-24']]" :key="lines[0]">
            <div v-for="width in lines" :key="width" class="flex h-5 items-center">
              <Skeleton :class="['h-3', width]" />
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div class="space-y-3">
            <div class="flex items-start gap-2.5 text-sm">
              <Skeleton class="mt-1 size-3.5 shrink-0 rounded-full" />
              <div class="flex-1">
                <span class="font-medium text-foreground">A specific moment</span>
                <div class="flex h-4 items-center">
                  <Skeleton class="h-2.5 w-56" />
                </div>
              </div>
            </div>
            <div class="space-y-4 pl-6">
              <div class="space-y-1.5">
                <div class="flex h-4 items-center justify-between">
                  <Skeleton class="h-2.5 w-12" />
                  <Skeleton class="h-2.5 w-10" />
                </div>
                <Skeleton class="h-8 w-full" />
                <div class="flex h-4 items-center justify-between">
                  <Skeleton class="h-2.5 w-20" />
                  <Skeleton class="h-2.5 w-20" />
                </div>
              </div>
              <div class="h-9 max-w-sm rounded-md border border-input bg-background" />
            </div>
            <div class="flex items-start gap-2.5 text-sm">
              <Skeleton class="mt-1 size-3.5 shrink-0 rounded-full" />
              <div class="flex-1">
                <span class="font-medium text-foreground">The most recent state</span>
                <div class="flex h-4 items-center">
                  <Skeleton class="h-2.5 w-48" />
                </div>
              </div>
            </div>
          </div>
          <div class="space-y-1.5">
            <p class="text-sm font-medium text-foreground">New database name</p>
            <div class="h-9 max-w-xs rounded-md border border-input bg-background" />
            <p class="text-xs text-muted-foreground">
              Up to 16 characters. Lowercase letters, numbers and dashes.
            </p>
          </div>
          <div class="flex justify-end">
            <Skeleton class="h-9 w-20 rounded-none" />
          </div>
        </div>
      </section>

      <section class="space-y-4">
        <div class="flex items-start justify-between gap-4">
          <div class="flex-1 space-y-1">
            <h3 class="text-sm font-medium text-muted-foreground">Base backups</h3>
            <div class="flex h-5 items-center">
              <Skeleton class="h-3 w-full max-w-sm" />
            </div>
            <div class="flex h-5 items-center">
              <Skeleton class="h-3 w-48" />
            </div>
          </div>
          <Skeleton class="h-9 w-28 shrink-0 rounded-none" />
        </div>
        <div class="-mx-4 divide-y border-y">
          <div v-for="row in 3" :key="row" class="flex items-start justify-between gap-4 px-4 py-3">
            <div class="space-y-1">
              <div class="flex h-5 items-center">
                <Skeleton class="h-3 w-28" />
              </div>
              <div class="flex h-4 items-center gap-3">
                <Skeleton class="h-2.5 w-16" />
                <Skeleton class="h-2.5 w-20" />
              </div>
            </div>
            <div class="flex h-5 items-center">
              <Skeleton class="h-3 w-16" />
            </div>
          </div>
        </div>
      </section>
    </template>

    <QueryError
      v-else-if="error"
      title="Couldn't load backups"
      :error="error"
      @retry="refetch()"
    />

    <template v-else-if="backups">
      <EmptyState
        v-if="!backups.enabled"
        size="compact"
        :icon="DatabaseBackup"
        title="Backups are not switched on yet"
        description="Continuous backups are rolling out across the platform. This database picks them up automatically."
      />

      <template v-else>
        <section class="space-y-4">
          <div class="space-y-1 text-sm text-muted-foreground/80">
            <h3 class="text-sm font-medium text-muted-foreground">Restore to a point in time</h3>
            <p v-if="hasWindow">
              Every write is archived as it happens, so you can restore to any moment in the last
              {{ backups.retentionDays }} days, not just when a backup ran.
            </p>
            <p v-else-if="degenerate">
              Every write is archived as it happens. Nothing has been written since the last
              backup, so there is a single state to restore to.
            </p>
            <p v-else>
              The first backup is still running. Once it finishes, everything from that moment
              onward becomes restorable.
            </p>
            <p>
              Restoring creates a second database holding the data as it was then. This database
              keeps running untouched.
            </p>
          </div>

          <div class="space-y-4">
            <p v-if="degenerate" class="text-sm text-muted-foreground">
              Nothing has been written since
              {{ formatDateTime(backups.latestRestorePoint) }}, so that is the only state there is
              to restore to.
            </p>

            <div v-else class="space-y-3">
              <label class="flex items-start gap-2.5 text-sm">
                <input v-model="mode" type="radio" value="moment" class="mt-1" />
                <span>
                  <span class="font-medium text-foreground">A specific moment</span>
                  <span class="block text-xs text-muted-foreground">
                    Anywhere in the window below. Times are in {{ timeZone }}.
                  </span>
                </span>
              </label>

              <div v-if="mode === 'moment'" class="space-y-4 pl-6">
                <RecoveryTimeline
                  v-if="hasWindow"
                  v-model="selected"
                  :from="earliest as number"
                  :to="latest"
                  :markers="markers"
                />
                <Input
                  id="restore-timestamp"
                  v-model="targetInput"
                  type="datetime-local"
                  step="1"
                  :min="inputMin"
                  :max="inputMax"
                  :disabled="!hasWindow"
                  class="h-9 max-w-sm tabular-nums"
                />
              </div>

              <label class="flex items-start gap-2.5 text-sm">
                <input v-model="mode" type="radio" value="latest" class="mt-1" />
                <span>
                  <span class="font-medium text-foreground">The most recent state</span>
                  <span class="block text-xs text-muted-foreground">
                    Everything in the archive, up to
                    {{ formatDateTime(backups.latestRestorePoint) }}.
                  </span>
                </span>
              </label>
            </div>

            <NameIdField
              v-model:name="restoreName"
              v-model:id="restoreId"
              fallback="database"
              :taken="takenIds"
              scope="environment"
              :disabled="!hasWindow"
              name-label="New database name"
              name-placeholder="e.g. Orders restored"
              class="max-w-sm"
            />

            <div class="flex justify-end">
              <Button :pending="restoring" :disabled="!canRestore" @click="handleRestore">
                {{ restoring ? 'Restoring…' : 'Restore' }}
              </Button>
            </div>
          </div>
        </section>


        <section class="space-y-4">
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-1 text-sm text-muted-foreground/80">
              <h3 class="text-sm font-medium text-muted-foreground">Base backups</h3>
              <p>
                A full copy runs every {{ fullBackupSchedule }} (your time), with every write in between archived
                continuously. Copies older than {{ backups.retentionDays }} days are removed.
              </p>
            </div>
            <Button
              variant="outline"
              class="shrink-0"
              :pending="backingUp"
              @click="handleBackupNow"
            >
              {{ backingUp ? 'Starting…' : 'Back up now' }}
            </Button>
          </div>

          <EmptyState
            v-if="entries.length === 0"
            size="compact"
            :icon="DatabaseBackup"
            title="No backups yet"
            description="The first one runs as soon as the database is ready."
          />

          <div v-else class="-mx-4 border-y">
            <div class="divide-y">
              <div v-for="entry in visibleEntries" :key="entry.id" class="px-4 py-3">
                <div class="flex items-start justify-between gap-4">
                  <div class="min-w-0 flex-1 space-y-1">
                    <p class="text-sm text-foreground">
                      {{ formatDateTime(entry.startedAt ?? entry.createdAt) }}
                    </p>
                    <div class="flex items-center gap-3 text-xs text-muted-foreground">
                      <span class="inline-flex items-center gap-1">
                        <CalendarClock
                          v-if="entry.trigger === BackupTrigger.Scheduled"
                          :size="13"
                        />
                        <MousePointerClick v-else :size="13" />
                        {{ entry.trigger === BackupTrigger.Scheduled ? 'Scheduled' : 'Manual' }}
                      </span>
                      <span v-if="duration(entry)" class="inline-flex items-center gap-1">
                        <Timer :size="13" />
                        took {{ duration(entry) }}
                      </span>
                    </div>
                  </div>
                  <span class="shrink-0 text-sm text-muted-foreground">
                    <Status :tone="backupStatus(entry.status).tone" variant="text">
                      {{ statusText(entry) }}
                    </Status>
                  </span>
                </div>
                <p
                  v-if="entry.error"
                  class="mt-2 rounded border border-destructive/20 bg-destructive/5 px-2 py-1 font-mono text-xs text-destructive"
                >
                  {{ entry.error }}
                </p>
              </div>

              <button
                v-if="entries.length > 5 && !showAll"
                type="button"
                class="w-full px-4 py-2 text-center text-xs text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground"
                @click="showAll = true"
              >
                Show {{ entries.length - 5 }} older
              </button>
            </div>
          </div>
        </section>
      </template>
    </template>
  </div>
</template>
