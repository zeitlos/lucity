import { watch, ref, type Ref, type ComputedRef } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { ConnectDatabaseMutation } from '@/graphql/databases';
import type { DatabaseInstance } from './useEnvironment';

/**
 * Watches database instances and auto-calls `connectDatabase` when a database
 * transitions to ready. Creates shared variables (DATABASE_URL) so services
 * can reference them via "Reference Shared".
 *
 * Fire-and-forget: failures are silently retried on the next poll cycle.
 */
export function useDatabaseAutoConnect(
  projectId: Ref<string>,
  environment: Ref<string | null>,
  databases: ComputedRef<DatabaseInstance[]>,
) {
  const connected = ref(new Set<string>());

  watch(databases, (dbs) => {
    const envName = environment.value;
    if (!envName) return;

    for (const db of dbs) {
      const key = `${envName}-${db.name}`;
      if (db.ready && !connected.value.has(key)) {
        connected.value.add(key);
        apolloClient.mutate({
          mutation: ConnectDatabaseMutation,
          variables: {
            projectId: projectId.value,
            environment: envName,
            database: db.name,
          },
        }).catch(() => {
          // Remove from set so it retries on next poll cycle.
          connected.value.delete(key);
        });
      }
    }
  }, { deep: true });
}
