<script setup lang="ts">
defineProps<{
  links: { id: string; text: string }[];
}>();

const open = ref(false);
const copied = ref(false);
let timer: ReturnType<typeof setTimeout> | undefined;

async function copy() {
  try {
    await navigator.clipboard.writeText(window.location.origin + window.location.pathname);
  } catch {
    return;
  }
  copied.value = true;
  clearTimeout(timer);
  timer = setTimeout(() => {
    copied.value = false;
  }, 2400);
}

onBeforeUnmount(() => clearTimeout(timer));
</script>

<template>
  <nav class="toc text-[15.5px]/[1.4]" aria-label="In this essay">
    <template v-if="links.length">
      <button
        type="button"
        class="toggle inline-flex min-h-11 cursor-pointer items-center gap-2 font-semibold text-stone-900 dark:text-stone-100"
        :aria-expanded="open"
        aria-controls="toc-list"
        @click="open = !open"
      >
        In this essay
        <UIcon name="i-lucide-chevron-down" class="size-4.5 shrink-0 transition-transform" :class="{ 'rotate-180': open }" />
      </button>
      <b class="title mb-1.5 font-semibold text-stone-900 dark:text-stone-100">In this essay</b>
      <ol id="toc-list" class="list" :class="{ open }">
        <li v-for="link in links" :key="link.id">
          <a
            :href="`#${link.id}`"
            class="block border-t border-stone-200 py-2.5 text-stone-700 hover:text-stone-900 dark:border-stone-800 dark:text-stone-400 dark:hover:text-stone-100"
            @click="open = false"
          >{{ link.text }}</a>
        </li>
      </ol>
    </template>
    <button
      type="button"
      class="copy inline-flex min-h-11 cursor-pointer items-center gap-2 font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100"
      @click="copy"
    >
      <UIcon :name="copied ? 'i-lucide-check' : 'i-lucide-link'" class="size-4.5 shrink-0" />
      <span aria-live="polite">{{ copied ? 'Link copied' : 'Copy link' }}</span>
    </button>
  </nav>
</template>

<style scoped>
.toc {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0 24px;
}

.title {
  display: none;
}

.list {
  display: none;
  order: 3;
  flex: 1 1 100%;
  margin-top: 6px;
}

.list.open {
  display: block;
}

.toggle:focus-visible,
.copy:focus-visible {
  outline: 2px solid oklch(0.55 0.15 300);
  outline-offset: 2px;
}

@media (min-width: 1100px) {
  .toc {
    display: block;
  }

  .toggle {
    display: none;
  }

  .title,
  .list {
    display: block;
    margin-top: 0;
  }

  .list + .copy {
    margin-top: 22px;
  }
}
</style>
