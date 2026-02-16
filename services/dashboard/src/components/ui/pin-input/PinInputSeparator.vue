<script setup lang="ts">
import { Dot } from 'lucide-vue-next';
</script>

<template>
  <div role="separator">
    <slot>
      <Dot />
    </slot>
  </div>
</template>
