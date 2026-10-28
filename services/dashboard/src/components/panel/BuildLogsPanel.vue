<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { BuildStatus } from '@/gql/graphql';
import LogViewer from '@/components/LogViewer.vue';
import type { LogsPanelKind } from '@/composables/useBuildLogsPanel';
import { useEscapeToClose } from '@/composables/useEscapeToClose';
import { useJobLogs } from '@/composables/useJobLogs';
import { formatDuration } from '@/lib/format';

const BuildLogsStatusDocument = graphql(`
  query BuildLogsStatus($id: BuildID!) {
    build(id: $id) {
      id
      status
      startedAt
      finishedAt
    }
  }
`);

const props = defineProps<{
  id: string;
  kind: LogsPanelKind;
  serviceName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

useEscapeToClose(() => emit('close'));

const titles: Record<LogsPanelKind, string> = { build: 'Build logs', deploy: 'Deploy logs', scan: 'Secret scan logs' };
const names: Record<LogsPanelKind, string> = { build: 'Build', deploy: 'Deploy', scan: 'Secret scan' };

const kindRef = computed(() => props.kind);
const idRef = computed(() => props.id);
const { lines, state, error, retry } = useJobLogs(kindRef, idRef);

const entries = computed(() => lines.value.map(text => ({ text })));

const { result: buildResult } = useQuery(
  BuildLogsStatusDocument,
  () => ({ id: props.id }),
  () => ({ enabled: props.kind === 'build' && state.value === 'finished', fetchPolicy: 'network-only' }),
);

const end = computed(() => {
  const name = names[props.kind];
  const build = props.kind === 'build' ? buildResult.value?.build : null;
  const took = build?.startedAt && build.finishedAt
    ? formatDuration(new Date(build.finishedAt).getTime() - new Date(build.startedAt).getTime())
    : null;
  switch (build?.status) {
    case BuildStatus.Succeeded:
      return { text: took ? `${name} succeeded in ${took}` : `${name} succeeded`, tone: 'ok' as const };
    case BuildStatus.Failed:
      return { text: took ? `${name} failed after ${took}` : `${name} failed`, tone: 'danger' as const };
    case BuildStatus.Cancelled:
      return { text: `${name} was cancelled`, tone: 'neutral' as const };
    default:
      return { text: `${name} finished`, tone: 'neutral' as const };
  }
});
</script>

<template>
  <LogViewer
    :title="titles[kind]"
    :lines="entries"
    :state="state"
    :error="error"
    :end="end"
    :file-name="`${serviceName}-${kind}-logs.txt`"
    @close="emit('close')"
    @retry="retry"
  />
</template>
