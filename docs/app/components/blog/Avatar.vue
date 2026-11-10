<script setup lang="ts">
import { authorOf, initialsOf } from './post';

const props = defineProps<{
  name?: string;
  large?: boolean;
}>();

const avatar = computed(() => authorOf(props.name).avatar);
const pixels = computed(() => props.large ? 56 : 44);
</script>

<template>
  <img
    v-if="avatar"
    :src="avatar"
    alt=""
    :width="pixels"
    :height="pixels"
    class="shrink-0 rounded-full"
    :class="large ? 'size-11 sm:size-14' : 'size-11'"
  >
  <span
    v-else
    class="inline-flex shrink-0 items-center justify-center rounded-full bg-[oklch(0.86_0.06_300)] font-semibold text-[oklch(0.30_0.10_300)]"
    :class="large ? 'size-11 text-base sm:size-14 sm:text-xl' : 'size-11 text-base'"
    aria-hidden="true"
  >{{ initialsOf(name) }}</span>
</template>
