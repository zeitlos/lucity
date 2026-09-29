<script setup lang="ts">
</script>

<template>
  <UContentSearchButton
    :collapsed="false"
    variant="outline"
    class="-mx-1 mb-4 w-[calc(100%+0.5rem)] justify-start py-2.5 bg-elevated hover:bg-accented"
    :ui="{ leadingIcon: 'size-4' }"
  />

  <DocsNavTree />
</template>
