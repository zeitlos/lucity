<script setup lang="ts">
import { computed, type Component } from 'vue';
import { CircleAlert, CircleCheck, Info, X } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import CopyButton from '@/components/CopyButton.vue';
import IconButton from '@/components/IconButton.vue';
import { cn } from '@/lib/utils';

const props = defineProps<{
  tone: 'ok' | 'danger' | 'neutral';
  title: string;
  description?: string;
  action?: { label: string; onClick: () => void };
  copy?: boolean;
  icon?: Component;
  isPaused?: boolean;
}>();
const emit = defineEmits<{ (e: 'closeToast'): void }>();

const tones = {
  ok: { icon: CircleCheck, glow: 'from-status-ok/10', mark: 'border-status-ok/20 text-status-ok' },
  danger: { icon: CircleAlert, glow: 'from-status-danger/10', mark: 'border-status-danger/20 text-status-danger' },
  neutral: { icon: Info, glow: 'from-muted', mark: 'text-muted-foreground' },
};

const tone = computed(() => tones[props.tone]);
const copyText = computed(() => (props.description ? `${props.title}: ${props.description}` : props.title));

function runAction() {
  props.action?.onClick();
  emit('closeToast');
}
</script>

<template>
  <div :class="cn('rounded-lg bg-gradient-to-b to-transparent to-70% p-3', tone.glow)">
    <div class="flex items-start gap-3">
      <div :class="cn('flex size-8 shrink-0 items-center justify-center rounded-md border bg-card shadow-sm', tone.mark)">
        <component :is="icon ?? tone.icon" :size="16" />
      </div>
      <div class="min-w-0 flex-1 pt-1.5">
        <p class="text-sm leading-5 font-medium">
          {{ title }}
        </p>
        <p v-if="description" class="mt-0.5 text-sm text-muted-foreground">
          {{ description }}
        </p>
        <Button
          v-if="action"
          variant="outline"
          size="xs"
          class="mt-3"
          @click="runAction"
        >
          {{ action.label }}
        </Button>
      </div>
      <div class="flex shrink-0 items-center pt-1">
        <CopyButton v-if="copy" :value="copyText" label="Copy message" />
        <IconButton label="Dismiss" size="icon-xs" @click="emit('closeToast')">
          <X />
        </IconButton>
      </div>
    </div>
  </div>
</template>
