<script setup lang="ts">
import { computed } from 'vue';

import { LogLevel, type PlatformLogsQuery } from '@/gql/graphql';
import { ago, change, count, hour, oneDecimal, shortDate } from '@/lib/platform';

type Logs = PlatformLogsQuery['platformLogs'];
type Message = Logs['messages'][number];

const props = defineProps<{
  logs: Logs | null;
  error: string | null;
}>();

const DAY = 24 * 60 * 60 * 1000;
const columns = 'grid-cols-[72px_minmax(0,1fr)_112px_88px_120px_120px]';

const bars = computed(() => {
  const hours = props.logs?.hours ?? [];
  const top = Math.max(1, ...hours.map((item) => item.errors + item.warnings));

  return hours.map((item) => ({
    start: item.start,
    errors: `${(item.errors / top) * 100}%`,
    warnings: `${(item.warnings / top) * 100}%`,
    label: `${hour(item.start)}, ${count(item.errors)} errors, ${count(item.warnings)} warnings`,
  }));
});

const ticks = computed(() => {
  const hours = props.logs?.hours ?? [];
  return [0, 6, 12, 18].filter((index) => hours[index]).map((index) => hour(hours[index]!.start));
});

function versusAverage(day: number, week: number): string {
  const average = week / 7;

  if (average === 0) return day === 0 ? 'none in 7 days' : 'none before';
  return `${change(day / average - 1)} vs 7-day avg`;
}

const errorsRising = computed(() => {
  const logs = props.logs;
  return logs != null && logs.errors24h > logs.errors7d / 7;
});

function rising(message: Message): boolean {
  return message.count24h > 0 && message.count24h > (message.count7d / 7) * 1.5;
}

function isNew(message: Message): boolean {
  return Date.now() - Date.parse(message.firstSeenAt) < DAY;
}
</script>

<template>
  <section aria-labelledby="logs-title" class="mt-4 border-t">
    <div class="flex min-h-11 items-center px-4">
      <h2 id="logs-title" class="font-semibold">Platform logs</h2>
      <span class="ml-auto text-muted-foreground">24 hours</span>
    </div>

    <p v-if="error" class="border-t border-hairline px-4 py-3 text-status-danger">{{ error }}</p>
    <div v-else-if="!logs" class="h-40 animate-pulse border-t border-hairline bg-sidebar" />

    <template v-else>
      <div class="flex flex-wrap items-end gap-x-12 gap-y-4 px-4 pt-1 pb-4.5 tabular-nums">
        <div class="flex min-w-0 flex-[999_1_640px] flex-col gap-1.5">
          <div role="img" aria-label="Errors and warnings per hour over the last 24 hours" class="flex h-30 items-end gap-1 border-b">
            <span v-for="bar in bars" :key="bar.start" :aria-label="bar.label" class="flex h-full flex-1 flex-col justify-end gap-px">
              <span class="rounded-t-[3px] bg-status-warn/42" :style="{ height: bar.warnings }" />
              <span class="rounded-[2px] bg-linear-to-b from-[oklch(0.58_0.19_25)] to-[oklch(0.48_0.19_25)]" :style="{ height: bar.errors }" />
            </span>
          </div>
          <div class="flex justify-between text-muted-foreground">
            <span v-for="tick in ticks" :key="tick">{{ tick }}</span>
            <span>now</span>
          </div>
        </div>
        <div class="flex flex-[1_1_260px] flex-col gap-1 pb-7">
          <span class="text-muted-foreground">Errors, 24 hours</span>
          <span class="flex items-baseline gap-2.5">
            <span class="text-3xl font-semibold" :class="logs.errors24h ? 'text-status-danger' : ''">{{ count(logs.errors24h) }}</span>
            <span :class="errorsRising ? 'font-medium text-status-danger' : 'text-muted-foreground'">{{ versusAverage(logs.errors24h, logs.errors7d) }}</span>
          </span>
          <span class="text-muted-foreground">{{ count(logs.warnings24h) }} warnings, {{ versusAverage(logs.warnings24h, logs.warnings7d) }}</span>
        </div>
      </div>

      <p v-if="logs.messages.length === 0" class="border-t border-hairline px-4 py-3 text-muted-foreground">No errors or warnings in the last 7 days.</p>
      <div v-else class="overflow-x-auto">
        <div class="min-w-[980px] tabular-nums">
          <div class="grid h-9 items-center gap-3 border-y border-hairline bg-sidebar px-4 text-muted-foreground" :class="columns">
            <span>Level</span>
            <span>Message</span>
            <span>Service</span>
            <span class="text-right">24h</span>
            <span class="text-right">Daily avg, 7d</span>
            <span class="text-right">First seen</span>
          </div>
          <div
            v-for="message in logs.messages"
            :key="`${message.service}-${message.level}-${message.message}`"
            class="grid min-h-10 items-center gap-3 border-b border-hairline px-4"
            :class="columns"
          >
            <span class="font-medium" :class="message.level === LogLevel.Error ? 'text-status-danger' : 'text-status-warn'">
              {{ message.level === LogLevel.Error ? 'error' : 'warn' }}
            </span>
            <span class="truncate font-mono">{{ message.message }}</span>
            <span class="text-muted-foreground">{{ message.service }}</span>
            <span class="text-right font-medium" :class="rising(message) ? 'text-status-danger' : ''">{{ count(message.count24h) }}</span>
            <span class="text-right text-muted-foreground">{{ oneDecimal(message.count7d / 7) }}</span>
            <span class="text-right" :class="isNew(message) ? 'font-semibold text-status-progress' : 'text-muted-foreground'">
              {{ isNew(message) ? `New, ${ago(message.firstSeenAt)}` : shortDate(message.firstSeenAt) }}
            </span>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>
