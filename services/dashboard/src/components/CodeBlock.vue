<script setup lang="ts">
import { computed, ref } from 'vue';
import { Check, Copy, Eye, EyeOff } from '@lucide/vue';
import { Button } from '@/components/ui/button';

const props = withDefaults(defineProps<{ code: string; secrets?: string[]; title?: string }>(), {
  secrets: () => [],
  title: undefined,
});

const revealed = ref(false);
const copied = ref(false);
let timer: ReturnType<typeof setTimeout> | undefined;

const shown = computed(() => {
  if (revealed.value) return props.code;
  return props.secrets
    .filter(Boolean)
    .reduce((code, secret) => code.split(secret).join('••••••••••••'), props.code);
});

const hasSecrets = computed(() => props.secrets.some(secret => secret && props.code.includes(secret)));

async function copy() {
  await navigator.clipboard.writeText(props.code);
  copied.value = true;
  clearTimeout(timer);
  timer = setTimeout(() => {
    copied.value = false;
  }, 1500);
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border bg-muted/40">
    <div class="flex min-h-10 items-center justify-between gap-2 border-b p-1.5">
      <div class="min-w-0">
        <slot name="header">
          <span v-if="title" class="truncate pl-1.5 font-mono text-xs text-muted-foreground">{{ title }}</span>
        </slot>
      </div>
      <div class="flex shrink-0 items-center gap-1">
        <Button
          v-if="hasSecrets"
          variant="ghost"
          size="xs"
          :aria-pressed="revealed"
          @click="revealed = !revealed"
        >
          <EyeOff v-if="revealed" />
          <Eye v-else />
          {{ revealed ? 'Hide secrets' : 'Show secrets' }}
        </Button>
        <Button variant="ghost" size="xs" @click="copy">
          <Check v-if="copied" class="text-status-ok" />
          <Copy v-else />
          {{ copied ? 'Copied' : 'Copy' }}
        </Button>
      </div>
    </div>
    <pre class="overflow-x-auto p-3 font-mono text-xs leading-relaxed">{{ shown }}</pre>
  </div>
</template>
