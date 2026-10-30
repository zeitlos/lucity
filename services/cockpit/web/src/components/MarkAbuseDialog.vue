<script setup lang="ts">
import { computed, onMounted, ref, useTemplateRef } from 'vue';
import { Ban, TriangleAlert } from '@lucide/vue';

import { avatarStyle, initial, signalLabel, type Signal, type Workspace } from '@/lib/workspaces';

const props = defineProps<{
  workspaces: Workspace[];
  running: boolean;
}>();

const emit = defineEmits<{
  close: [];
  confirm: [workspaceIDs: string[], note: string];
}>();

const dialog = useTemplateRef<HTMLDialogElement>('dialog');
const note = ref('');

const pods = computed(() => props.workspaces.reduce((sum, workspace) => sum + workspace.runningPods + workspace.pendingPods, 0));
const environments = computed(() => props.workspaces.reduce((sum, workspace) => sum + workspace.environments.length, 0));

function plural(count: number, word: string): string {
  return `${count} ${count === 1 ? word : word + 's'}`;
}

function firstSignal(workspace: Workspace): Signal | undefined {
  return workspace.signals.find((signal) => signal.strong) ?? workspace.signals[0];
}

onMounted(() => dialog.value?.showModal());
</script>

<template>
  <dialog
    ref="dialog"
    aria-labelledby="mark-title"
    class="m-auto w-[540px] max-w-[calc(100vw-32px)] rounded-xl bg-card p-1.5 text-foreground shadow-lg backdrop:bg-foreground/35"
    @cancel.prevent="emit('close')"
  >
    <div class="flex flex-col gap-4 rounded-lg bg-linear-to-b from-muted from-0% to-card to-72% px-5 pt-5 pb-4">
      <span class="flex size-9 items-center justify-center rounded-sm bg-card text-status-danger shadow-xs" aria-hidden="true">
        <Ban class="size-4.5" />
      </span>
      <div class="flex flex-col gap-1.5">
        <h2 id="mark-title" class="font-semibold">Mark {{ workspaces.length === 1 ? workspaces[0]!.id : plural(workspaces.length, 'workspace') }} as abuse</h2>
        <p class="text-muted-foreground">
          Stops {{ plural(pods, 'pod') }} in {{ plural(environments, 'environment') }} and blocks new deploys. You can undo it afterwards.
        </p>
      </div>

      <ul class="max-h-56 overflow-y-auto border-t border-hairline">
        <li
          v-for="workspace in workspaces"
          :key="workspace.id"
          class="grid min-h-11 grid-cols-[24px_minmax(0,1fr)_auto_64px] items-center gap-x-2.5 border-b border-hairline"
        >
          <span class="flex size-6 items-center justify-center rounded-xs font-semibold" :style="avatarStyle(workspace.id)" aria-hidden="true">{{ initial(workspace.id) }}</span>
          <span class="truncate font-medium">{{ workspace.id }}</span>
          <span
            v-if="firstSignal(workspace)?.strong"
            class="inline-flex max-w-56 items-center gap-1 truncate rounded-md bg-status-danger-background px-1.5 py-0.5 text-status-danger"
          >
            <TriangleAlert class="size-3 shrink-0" aria-hidden="true" />
            <span class="truncate">{{ signalLabel(firstSignal(workspace)!) }}</span>
          </span>
          <span v-else-if="firstSignal(workspace)" class="max-w-56 truncate text-status-warn">{{ signalLabel(firstSignal(workspace)!) }}</span>
          <span v-else class="text-muted-foreground">No signal</span>
          <span class="text-right text-muted-foreground tabular-nums">{{ plural(workspace.runningPods + workspace.pendingPods, 'pod') }}</span>
        </li>
      </ul>

      <label class="flex flex-col gap-1.5 font-medium">
        Note
        <textarea
          v-model="note"
          rows="2"
          maxlength="500"
          placeholder="Panel wave, shared commit"
          class="resize-none rounded-sm border bg-card px-2.5 py-2 font-normal"
        />
      </label>
    </div>

    <div class="mt-1.5 flex justify-end gap-2 border-t border-hairline px-3.5 pt-3 pb-2">
      <button type="button" class="h-9 rounded-sm border px-3.5 hover:bg-muted" :disabled="running" @click="emit('close')">{{ workspaces.length === 1 ? 'Keep it running' : 'Keep them running' }}</button>
      <button
        type="button"
        class="h-9 rounded-sm bg-destructive px-3.5 font-medium text-white disabled:opacity-50"
        :disabled="running"
        @click="emit('confirm', workspaces.map((workspace) => workspace.id), note.trim())"
      >
        {{ running ? 'Marking…' : 'Mark as abuse' }}
      </button>
    </div>
  </dialog>
</template>
