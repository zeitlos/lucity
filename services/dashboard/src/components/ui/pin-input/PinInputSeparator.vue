<script setup lang="ts">
import { Dot } from '@lucide/vue';
</script>

<template>
  <div role="separator">
    <slot>
      <Dot />
    </slot>
  </div>
</template>
