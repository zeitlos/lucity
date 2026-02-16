<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { computed, onMounted, onUnmounted, provide, ref, watch } from 'vue';
import { useMediaQuery } from '@vueuse/core';
import { cn } from '@/lib/utils';
import { SIDEBAR_COOKIE_MAX_AGE, SIDEBAR_COOKIE_NAME, SIDEBAR_KEYBOARD_SHORTCUT, SIDEBAR_WIDTH, SIDEBAR_WIDTH_ICON, type SidebarContext, SidebarSymbol } from './utils';
import { TooltipProvider } from '@/components/ui/tooltip';

const props = withDefaults(defineProps<{
  class?: HTMLAttributes['class'];
  defaultOpen?: boolean;
  open?: boolean;
}>(), {
  defaultOpen: true,
});

const emits = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const isMobile = useMediaQuery('(max-width: 768px)');
const openMobile = ref(false);
const _open = ref(props.defaultOpen);

const open = computed({
  get: () => props.open ?? _open.value,
  set: (value) => {
    _open.value = value;
    emits('update:open', value);
    document.cookie = `${SIDEBAR_COOKIE_NAME}=${value}; path=/; max-age=${SIDEBAR_COOKIE_MAX_AGE}`;
  },
});

const state = computed(() => open.value ? 'expanded' : 'collapsed');

function setOpen(value: boolean) {
  open.value = value;
}

function setOpenMobile(value: boolean) {
  openMobile.value = value;
}

function toggleSidebar() {
  if (isMobile.value) {
    openMobile.value = !openMobile.value;
  } else {
    open.value = !open.value;
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === SIDEBAR_KEYBOARD_SHORTCUT && (event.metaKey || event.ctrlKey)) {
    event.preventDefault();
    toggleSidebar();
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
});

const context: SidebarContext = {
  state,
  open,
  setOpen,
  openMobile,
  setOpenMobile,
  isMobile,
  toggleSidebar,
};

provide(SidebarSymbol, context);
</script>

<template>
  <TooltipProvider :delay-duration="0">
    <div
      :class="cn('group/sidebar-wrapper flex min-h-svh w-full has-[[data-variant=inset]]:bg-sidebar', props.class)"
      :style="{
        '--sidebar-width': SIDEBAR_WIDTH,
        '--sidebar-width-icon': SIDEBAR_WIDTH_ICON,
      } as any"
    >
      <slot />
    </div>
  </TooltipProvider>
</template>
