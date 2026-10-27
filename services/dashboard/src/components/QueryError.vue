<script setup lang="ts">
import { computed } from 'vue';
import { CircleAlert, RotateCw, WifiOff } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { describeError } from '@/lib/errors';
import { openBugReport } from '@/composables/useReportBug';

const props = defineProps<{ title: string; error: unknown }>();
const emit = defineEmits<{ (e: 'retry'): void }>();

const described = computed(() => describeError(props.error));
const kind = computed(() => described.value.kind);
const message = computed(() => described.value.message);

function report() {
  openBugReport({ title: props.title, error: described.value.detail });
}
</script>

<template>
  <div
    role="alert"
    class="relative isolate overflow-hidden rounded-xl border border-status-danger/25 bg-card p-4 shadow-sm before:pointer-events-none before:absolute before:inset-1.5 before:-z-10 before:rounded-md before:bg-gradient-to-b before:from-status-danger/10 before:to-transparent before:to-70%"
  >
    <div class="flex items-start gap-3">
      <div class="flex size-9 shrink-0 items-center justify-center rounded-md border border-status-danger/20 bg-card text-status-danger shadow-sm">
        <WifiOff v-if="kind === 'network'" :size="16" />
        <CircleAlert v-else :size="16" />
      </div>
      <div class="flex min-w-0 flex-1 flex-wrap items-center gap-x-4 gap-y-3">
        <div class="min-w-48 flex-1">
          <p class="text-sm font-medium">
            {{ title }}
          </p>
          <p class="mt-0.5 text-sm text-muted-foreground">
            {{ message }}
          </p>
        </div>
        <div class="ml-auto flex shrink-0 gap-2">
          <Button
            v-if="kind === 'unexpected'"
            variant="ghost"
            size="xs"
            @click="report"
          >
            Report
          </Button>
          <Button
            variant="outline"
            size="xs"
            @click="emit('retry')"
          >
            <RotateCw />Try again
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
