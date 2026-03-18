<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { Circle } from 'lucide-vue-next';

const version = typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : 'dev';

const expanded = ref(false);
const loading = ref(false);
const fetched = ref(false);
const gatewayUp = ref(false);
const components = ref<{ name: string; status: string }[]>([]);

async function toggle() {
  expanded.value = !expanded.value;
  if (!expanded.value || fetched.value) return;

  loading.value = true;
  try {
    const res = await fetch('/version');
    if (!res.ok) throw new Error();
    const data = await res.json();
    gatewayUp.value = true;
    components.value = data.components ?? [];
  } catch {
    gatewayUp.value = false;
  } finally {
    loading.value = false;
    fetched.value = true;
  }
}

const root = ref<HTMLElement>();

function onClickOutside(e: MouseEvent) {
  if (expanded.value && root.value && !root.value.contains(e.target as Node)) {
    expanded.value = false;
  }
}

onMounted(() => document.addEventListener('click', onClickOutside));
onUnmounted(() => document.removeEventListener('click', onClickOutside));
</script>

<template>
  <div ref="root" class="fixed bottom-0 left-0 z-50 pb-px pl-4">
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      leave-active-class="transition-all duration-150 ease-in"
      enter-from-class="opacity-0 translate-y-2 scale-95"
      leave-to-class="opacity-0 translate-y-2 scale-95"
    >
      <div
        v-if="expanded"
        class="mb-1.5 min-w-40 rounded-lg border bg-card/95 p-3 shadow-lg backdrop-blur-sm"
      >
        <p class="mb-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          Components
        </p>

        <div v-if="loading" class="py-1 text-xs text-muted-foreground">
          Checking...
        </div>

        <template v-else>
          <div class="flex items-center gap-1.5 py-1">
            <Circle
              :size="6"
              :class="gatewayUp
                ? 'fill-emerald-500 text-emerald-500'
                : 'fill-red-500 text-red-500'"
            />
            <span class="text-xs text-foreground">gateway</span>
          </div>

          <div
            v-for="c in components"
            :key="c.name"
            class="flex items-center gap-1.5 py-1"
          >
            <Circle
              :size="6"
              :class="c.status === 'UP'
                ? 'fill-emerald-500 text-emerald-500'
                : 'fill-red-500 text-red-500'"
            />
            <span class="text-xs text-foreground">{{ c.name }}</span>
          </div>

          <div class="flex items-center gap-1.5 py-1">
            <Circle :size="6" class="fill-emerald-500 text-emerald-500" />
            <span class="text-xs text-foreground">dashboard</span>
          </div>
        </template>
      </div>
    </Transition>

    <button
      class="font-mono text-[10px] text-muted-foreground/50 transition-colors hover:text-muted-foreground"
      @click="toggle"
    >
      {{ version }}
    </button>
  </div>
</template>
