<script setup lang="ts">
import { computed, ref } from 'vue';
import { Eye, EyeOff } from '@lucide/vue';
import CopyButton from '@/components/CopyButton.vue';
import IconButton from '@/components/IconButton.vue';

const props = defineProps<{ label: string; value: string; sensitive?: boolean }>();

const revealed = ref(false);
const name = computed(() => (/^[A-Z][a-z]/.test(props.label) ? props.label.charAt(0).toLowerCase() + props.label.slice(1) : props.label));
</script>

<template>
  <div class="flex items-start gap-2 rounded-md bg-muted/40 px-3 py-2">
    <span class="w-28 shrink-0 pt-1 text-xs font-medium text-muted-foreground">{{ label }}</span>
    <span class="min-w-0 flex-1 pt-1 font-mono text-xs break-all text-foreground">
      {{ sensitive && !revealed ? '••••••••••••' : value }}
    </span>
    <IconButton
      v-if="sensitive"
      :label="revealed ? `Hide ${name}` : `Show ${name}`"
      :aria-pressed="revealed"
      size="icon-xs"
      class="shrink-0"
      @click="revealed = !revealed"
    >
      <EyeOff v-if="revealed" />
      <Eye v-else />
    </IconButton>
    <CopyButton :label="`Copy ${name}`" :value="value" class="shrink-0" />
  </div>
</template>
