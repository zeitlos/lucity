import { onScopeDispose, ref, watch, type Ref } from 'vue';
import { graphql } from '@/gql';
import { apolloClient } from '@/lib/apollo';
import type { LogsPanelKind } from '@/composables/useBuildLogsPanel';

const BuildLogsDocument = graphql(`
  subscription BuildLogs($id: BuildID!) {
    buildLogs(id: $id)
  }
`);

const DeployLogsDocument = graphql(`
  subscription DeployLogs($id: DeployID!) {
    deployLogs(id: $id)
  }
`);

const ScanLogsDocument = graphql(`
  subscription ScanLogs($id: ScanID!) {
    scanLogs(id: $id)
  }
`);

export type LogStreamState = 'connecting' | 'streaming' | 'finished' | 'error';

function subscribe(kind: LogsPanelKind, id: string) {
  switch (kind) {
    case 'deploy':
      return apolloClient.subscribe({ query: DeployLogsDocument, variables: { id } })
        .map(result => result.data?.deployLogs);
    case 'scan':
      return apolloClient.subscribe({ query: ScanLogsDocument, variables: { id } })
        .map(result => result.data?.scanLogs);
    default:
      return apolloClient.subscribe({ query: BuildLogsDocument, variables: { id } })
        .map(result => result.data?.buildLogs);
  }
}

/**
 * Streams the logs of one build, deploy or scan job. The stream ends when the job does.
 */
export function useJobLogs(kind: Ref<LogsPanelKind>, id: Ref<string | null>) {
  const lines = ref<string[]>([]);
  const state = ref<LogStreamState>('connecting');
  const error = ref<unknown>(null);
  let subscription: { unsubscribe(): void } | null = null;

  function stop() {
    subscription?.unsubscribe();
    subscription = null;
  }

  function start() {
    stop();
    lines.value = [];
    error.value = null;
    state.value = 'connecting';
    if (!id.value) return;

    subscription = subscribe(kind.value, id.value).subscribe({
      next: line => {
        if (line == null) return;
        lines.value.push(line);
        state.value = 'streaming';
      },
      error: e => {
        error.value = e;
        state.value = 'error';
      },
      complete: () => {
        state.value = 'finished';
      },
    });
  }

  watch([kind, id], start, { immediate: true });
  onScopeDispose(stop);

  return { lines, state, error, retry: start };
}
