<script setup lang="ts">
const props = defineProps<{
  href?: string;
  target?: string;
}>();

const external = computed(() => /^https?:\/\//.test(props.href ?? ''));
</script>

<template>
  <NuxtLink :to="href" :target="external ? (target ?? '_blank') : target">
    <slot /><span v-if="external" class="sr-only"> (opens in a new tab)</span>
  </NuxtLink>
</template>
