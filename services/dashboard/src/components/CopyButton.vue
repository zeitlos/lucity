<script setup lang="ts">
import { ref } from 'vue';
import { Check, Copy } from '@lucide/vue';
import IconButton from '@/components/IconButton.vue';
import type { ButtonVariants } from '@/components/ui/button';

const props = withDefaults(defineProps<{
  value: string;
  label?: string;
  size?: ButtonVariants['size'];
  variant?: ButtonVariants['variant'];
}>(), {
  label: 'Copy',
  size: 'icon-xs',
  variant: 'ghost',
});

const copied = ref(false);
let timer: ReturnType<typeof setTimeout> | undefined;

async function copy() {
  await navigator.clipboard.writeText(props.value);
  copied.value = true;
  clearTimeout(timer);
  timer = setTimeout(() => {
    copied.value = false;
  }, 1500);
}
</script>

<template>
  <IconButton
    :label="copied ? 'Copied' : label"
    :size="size"
    :variant="variant"
    @click="copy"
  >
    <Check v-if="copied" class="text-status-ok" />
    <Copy v-else />
  </IconButton>
</template>
