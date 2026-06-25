<script setup lang="ts">
import { Braces } from '@lucide/vue';
import BucketIcon from '@/components/BucketIcon.vue';

withDefaults(
  defineProps<{
    typename?: 'DatabaseSource' | 'KeyValueStoreSource' | 'BucketSource' | 'SharedSource' | null;
    size?: number;
  }>(),
  { size: 14 },
);
</script>

<template>
  <img
    v-if="typename === 'DatabaseSource'"
    src="https://devicons.railway.com/i/postgresql.svg"
    :width="size"
    :height="size"
    alt=""
    class="shrink-0"
  />
  <img
    v-else-if="typename === 'KeyValueStoreSource'"
    src="https://devicons.railway.com/i/redis.svg"
    :width="size"
    :height="size"
    alt=""
    class="shrink-0"
  />
  <BucketIcon v-else-if="typename === 'BucketSource'" :size="size" class="shrink-0" />
  <Braces v-else-if="typename === 'SharedSource'" :size="size" class="shrink-0" />
</template>
