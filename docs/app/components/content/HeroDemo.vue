<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';

const video = ref<HTMLVideoElement | null>(null);
const currentTime = ref(0);

const segments = [
  {
    // icon: 'i-lucide-rocket',
    label: 'Build & Deploy',
    description: 'From GitHub repo to running app with zero config.',
    color: 'oklch(0.75 0.18 160)',
    start: 0,
    end: 22,
  },
  {
    // icon: 'i-lucide-globe',
    label: 'Go Live',
    description: 'Free public domain or bring custom domains. Natrually with autoamtic TLS.',
    color: 'oklch(0.70 0.22 0)',
    start: 22,
    end: 26,
  },
  {
    // icon: 'i-lucide-database',
    label: 'Batteries included',
    description: 'Deploy PostgreSQL, Redis, Object Storage in one click. Automatically wired to your app.',
    color: 'oklch(0.85 0.15 95)',
    start: 26,
    end: 44,
  },
  {
    // icon: 'i-lucide-table',
    label: 'Clever integrations',
    description: 'Browse datbases, configure dynamic variables and scale effotlessly.',
    color: 'oklch(0.72 0.14 300)',
    start: 44,
    end: 60.33,
  },
];

const activeIndex = computed(() => {
  const t = currentTime.value;
  for (let i = segments.length - 1; i >= 0; i--) {
    if (t >= segments[i].start) return i;
  }
  return 0;
});

function segmentProgress(index: number) {
  const seg = segments[index];
  const duration = seg.end - seg.start;
  if (currentTime.value < seg.start) return 0;
  if (currentTime.value >= seg.end) return 100;
  return ((currentTime.value - seg.start) / duration) * 100;
}

function seek(index: number) {
  if (!video.value) return;
  video.value.currentTime = segments[index].start;
  video.value.play();
}

function onTimeUpdate() {
  if (video.value) {
    currentTime.value = video.value.currentTime;
  }
}

onMounted(() => {
  video.value?.addEventListener('timeupdate', onTimeUpdate);
});

onUnmounted(() => {
  video.value?.removeEventListener('timeupdate', onTimeUpdate);
});
</script>

<template>
  <div class="flex justify-center">
    <div class="w-full">
      <div class="shadow-2xl rounded-3xl overflow-hidden">
        <video
          ref="video"
          src="/video/demo.mp4"
          autoplay
          muted
          loop
          playsinline
          preload="metadata"
          class="block h-auto w-full"
        />
      </div>
      <div class="mt-3 flex flex-col items-stretch gap-1.5 min-[651px]:flex-row min-[651px]:flex-wrap min-[651px]:gap-0.5 min-[1001px]:flex-nowrap">
        <button
          v-for="(seg, i) in segments"
          :key="seg.label"
          class="flex w-full cursor-pointer flex-col items-start gap-1.5 rounded-lg px-4 py-3 text-left transition-colors duration-200 hover:bg-[var(--ui-bg-elevated)] min-[651px]:w-[calc(50%-1px)] min-[1001px]:w-auto min-[1001px]:flex-1"
          :class="activeIndex === i ? 'bg-[var(--ui-bg-elevated)]' : ''"
          @click="seek(i)"
        >
          <div class="text-muted text-xl font-bold" :class="{ 'text-neutral-950 dark:text-neutral-50': activeIndex === i }">
            {{ seg.label }}
          </div>
          <span class="font-sans-condensed">
            {{ seg.description }}
          </span>
          <div class="mt-auto h-1.5 w-full overflow-hidden rounded-sm bg-[var(--ui-border)] max-[650px]:mt-2">
            <div
              class="h-full rounded transition-[width] duration-[250ms] ease-linear"
              :style="{
                width: `${segmentProgress(i)}%`,
                background: seg.color,
              }"
            />
          </div>
        </button>
      </div>
    </div>
  </div>
</template>
