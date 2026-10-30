<script setup lang="ts">
import { ref, watch } from 'vue';

import { avatarStyle, initial } from '@/lib/workspaces';

const props = defineProps<{
  src: string | null | undefined;
  label: string;
}>();

const failed = ref(false);

watch(() => props.src, () => {
  failed.value = false;
});
</script>

<template>
  <img
    v-if="src && !failed"
    :src="src"
    alt=""
    class="shrink-0 rounded-full bg-muted object-cover"
    @error="failed = true"
  >
  <span
    v-else
    class="flex shrink-0 items-center justify-center rounded-full font-semibold"
    :style="avatarStyle(label)"
    aria-hidden="true"
  >{{ initial(label) }}</span>
</template>
