<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { Eye, EyeOff } from '@lucide/vue';
import IconButton from '@/components/IconButton.vue';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';

defineOptions({ inheritAttrs: false });

const value = defineModel<string>({ required: true });
const revealed = defineModel<boolean>('revealed', { default: false });

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <div class="relative">
    <Input
      v-bind="$attrs"
      v-model="value"
      type="text"
      autocomplete="off"
      autocapitalize="off"
      spellcheck="false"
      data-1p-ignore
      data-lpignore="true"
      data-bwignore
      :class="cn('pr-9 font-mono text-sm', !revealed && value && '[-webkit-text-security:disc]', props.class)"
    />
    <IconButton
      :label="revealed ? 'Hide value' : 'Show value'"
      :aria-pressed="revealed"
      size="icon"
      tabindex="-1"
      class="absolute right-0 top-0 text-muted-foreground hover:bg-transparent"
      @click="revealed = !revealed"
    >
      <EyeOff v-if="revealed" />
      <Eye v-else />
    </IconButton>
  </div>
</template>
