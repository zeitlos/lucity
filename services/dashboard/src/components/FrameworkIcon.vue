<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Container } from '@lucide/vue';
import { frameworkIconUrl } from '@/lib/frameworkIcon';

const props = withDefaults(
  defineProps<{
    framework?: string;
    language?: string;
    size?: number;
  }>(),
  { size: 20 },
);

const failed = ref(false);
const src = computed(() => frameworkIconUrl(props.framework ?? '', props.language ?? ''));

watch(src, () => {
  failed.value = false;
});
</script>

<template>
  <img
    v-if="src && !failed"
    :src="src"
    :width="size"
    :height="size"
    class="shrink-0"
    alt=""
    @error="failed = true"
  />
  <Container v-else :size="size" class="shrink-0 text-muted-foreground" />
</template>
