<script setup lang="ts">
import { inject, type Ref } from 'vue';

const id = inject<Ref<string>>('form-item-id');
</script>

<template>
  <div>
    <slot :id="id" />
  </div>
</template>
