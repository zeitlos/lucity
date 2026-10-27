<script setup lang="ts">
import { devicon } from '@/lib/devicons';
import { HardDrive } from '@lucide/vue';
import GithubIcon from '@/components/GithubIcon.vue';
import SourceIcon from '@/components/SourceIcon.vue';
import type { ResourceKind } from '@/lib/resources';

withDefaults(
  defineProps<{
    kind: ResourceKind;
    size?: number;
  }>(),
  { size: 14 },
);
</script>

<template>
  <GithubIcon v-if="kind === 'repositoryService'" :size="size" class="shrink-0" />
  <img
    v-else-if="kind === 'imageService'"
    :src="devicon('docker')"
    :width="size"
    :height="size"
    alt=""
    class="shrink-0"
  />
  <SourceIcon v-else-if="kind === 'database'" typename="DatabaseSource" :size="size" />
  <SourceIcon v-else-if="kind === 'keyValueStore'" typename="KeyValueStoreSource" :size="size" />
  <SourceIcon v-else-if="kind === 'bucket'" typename="BucketSource" :size="size" />
  <HardDrive v-else :size="size" class="shrink-0" />
</template>
