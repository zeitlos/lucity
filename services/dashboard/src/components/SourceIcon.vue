<script setup lang="ts">
import { devicon } from '@/lib/devicons';
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
    :src="devicon('postgresql')"
    :width="size"
    :height="size"
    alt=""
    class="shrink-0"
  />
  <img
    v-else-if="typename === 'KeyValueStoreSource'"
    :src="devicon('redis')"
    :width="size"
    :height="size"
    alt=""
    class="shrink-0"
  />
  <BucketIcon v-else-if="typename === 'BucketSource'" :size="size" class="shrink-0" />
  <Braces v-else-if="typename === 'SharedSource'" :size="size" class="shrink-0" />
</template>
