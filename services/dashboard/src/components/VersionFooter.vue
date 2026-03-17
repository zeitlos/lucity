<script setup lang="ts">
import { ref } from 'vue';
import { ChevronUp, Circle } from 'lucide-vue-next';

const dashboardVersion = typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : 'dev';

const expanded = ref(false);
const loading = ref(false);
const gatewayVersion = ref<string | null>(null);
const components = ref<{ name: string; status: string }[]>([]);
const fetchError = ref(false);

async function toggle() {
  if (expanded.value) {
    expanded.value = false;
    return;
  }

  expanded.value = true;

  if (gatewayVersion.value !== null) return;

  loading.value = true;
  fetchError.value = false;
  try {
    const res = await fetch('/version');
    if (!res.ok) throw new Error('Failed to fetch');
    const data = await res.json();
    gatewayVersion.value = data.version;
    components.value = data.components ?? [];
  } catch {
    fetchError.value = true;
    gatewayVersion.value = null;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="fixed bottom-3 left-3 z-50">
    <!-- Expanded panel -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      leave-active-class="transition-all duration-150 ease-in"
      enter-from-class="opacity-0 translate-y-2 scale-95"
      leave-to-class="opacity-0 translate-y-2 scale-95"
    >
      <div
        v-if="expanded"
        class="mb-1.5 min-w-48 rounded-lg border bg-card/95 p-3 shadow-lg backdrop-blur-sm"
      >
        <p class="mb-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          Components
        </p>

        <div v-if="loading" class="py-1 text-xs text-muted-foreground">
          Checking...
        </div>

        <div v-else-if="fetchError" class="py-1 text-xs text-muted-foreground">
          Could not reach gateway
        </div>

        <template v-else>
          <!-- Gateway -->
          <div
            v-if="gatewayVersion"
            class="flex items-center justify-between gap-4 py-1"
          >
            <div class="flex items-center gap-1.5">
              <Circle :size="6" class="fill-emerald-500 text-emerald-500" />
              <span class="text-xs text-foreground">gateway</span>
            </div>
            <span class="font-mono text-[11px] text-muted-foreground">{{ gatewayVersion }}</span>
          </div>

          <!-- Backend services -->
          <div
            v-for="c in components"
            :key="c.name"
            class="flex items-center justify-between gap-4 py-1"
          >
            <div class="flex items-center gap-1.5">
              <Circle
                :size="6"
                :class="c.status === 'UP'
                  ? 'fill-emerald-500 text-emerald-500'
                  : 'fill-red-500 text-red-500'"
              />
              <span class="text-xs text-foreground">{{ c.name }}</span>
            </div>
            <span class="text-[11px] text-muted-foreground">{{ c.status === 'UP' ? 'healthy' : 'unreachable' }}</span>
          </div>

          <!-- Dashboard -->
          <div class="flex items-center justify-between gap-4 py-1">
            <div class="flex items-center gap-1.5">
              <Circle :size="6" class="fill-emerald-500 text-emerald-500" />
              <span class="text-xs text-foreground">dashboard</span>
            </div>
            <span class="font-mono text-[11px] text-muted-foreground">{{ dashboardVersion }}</span>
          </div>
        </template>
      </div>
    </Transition>

    <!-- Version pill -->
    <button
      class="flex items-center gap-1.5 rounded-full border bg-card/80 px-2.5 py-1 text-[11px] text-muted-foreground shadow-sm backdrop-blur-sm transition-colors hover:text-foreground"
      @click="toggle"
    >
      <span class="font-mono">{{ dashboardVersion }}</span>
      <ChevronUp
        :size="12"
        class="transition-transform duration-200"
        :class="expanded ? '' : 'rotate-180'"
      />
    </button>
  </div>
</template>
