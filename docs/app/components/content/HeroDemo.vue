<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';

const video = ref<HTMLVideoElement | null>(null);
const currentTime = ref(0);

const segments = [
  {
    icon: 'i-lucide-rocket',
    label: 'Build & Deploy',
    description: 'From GitHub repo to running app. Zero config.',
    color: 'oklch(0.75 0.18 160)',
    start: 0,
    end: 22,
  },
  {
    icon: 'i-lucide-globe',
    label: 'Go Live',
    description: 'Publicly reachable in seconds, with custom domains and automatic TLS',
    color: 'oklch(0.70 0.22 0)',
    start: 22,
    end: 26,
  },
  {
    icon: 'i-lucide-database',
    label: 'Managed Databases',
    description: 'PostgreSQL in one click, auto-wired',
    color: 'oklch(0.85 0.15 95)',
    start: 26,
    end: 44,
  },
  {
    icon: 'i-lucide-table',
    label: 'Database Explorer',
    description: 'Browse tables and run queries in the dashboard',
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
  <div class="hero-demo">
    <div class="video-container">
      <div class="video-wrapper">
        <video
          ref="video"
          src="/video/demo.mp4"
          autoplay
          muted
          loop
          playsinline
          preload="auto"
        />
      </div>
      <div class="tabs">
        <button
          v-for="(seg, i) in segments"
          :key="seg.label"
          class="tab"
          :class="{ 'tab-active': activeIndex === i }"
          @click="seek(i)"
        >
          <div class="tab-header">
            <span
              class="tab-icon"
              :style="activeIndex === i ? { color: seg.color } : undefined"
            >
              <UIcon
                :name="seg.icon"
                class="size-4"
              />
            </span>
            <span class="tab-label">{{ seg.label }}</span>
          </div>
          <span class="tab-description">{{ seg.description }}</span>
          <div class="progress-track">
            <div
              class="progress-fill"
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

<style scoped>
.hero-demo {
  display: flex;
  justify-content: center;
  padding: 2rem 1rem 0;
}

.video-container {
  width: 100%;
  max-width: 1100px;
}

.video-wrapper {
  border-radius: 12px;
  overflow: hidden;
  box-shadow:
    0 4px 24px oklch(0.3 0.02 55 / 0.08),
    0 1px 4px oklch(0.3 0.02 55 / 0.04);
  border: 1px solid var(--ui-border);
}

.video-wrapper video {
  display: block;
  width: 100%;
  height: auto;
}

.tabs {
  display: flex;
  gap: 2px;
  margin-top: 12px;
  align-items: stretch;
}

.tab {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  padding: 12px 16px;
  border-radius: 8px;
  border: none;
  background: transparent;
  cursor: pointer;
  transition: background 0.2s ease;
  text-align: left;
}

.tab:hover {
  background: var(--ui-bg-elevated);
}

.tab-active {
  background: var(--ui-bg-elevated);
}

.tab-header {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tab-icon {
  color: var(--ui-text-muted);
  transition: color 0.2s ease;
  display: flex;
  align-items: center;
}

.tab-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--ui-text-muted);
  transition: color 0.2s ease;
  white-space: nowrap;
}

.tab-active .tab-label {
  color: var(--ui-text);
}

.tab-description {
  font-size: 12px;
  color: var(--ui-text-dimmed);
  line-height: 1.4;
  transition: color 0.2s ease;
}

.tab-active .tab-description {
  color: var(--ui-text-muted);
}

.progress-track {
  width: 100%;
  height: 2px;
  background: var(--ui-border);
  border-radius: 1px;
  overflow: hidden;
  margin-top: auto;
}

.progress-fill {
  height: 100%;
  border-radius: 1px;
  transition: width 0.25s linear;
}

/* Dark mode shadow adjustment */
.dark .video-wrapper {
  box-shadow:
    0 4px 24px oklch(0 0 0 / 0.3),
    0 1px 4px oklch(0 0 0 / 0.2);
}

/* Responsive: hide descriptions on mobile */
@media (max-width: 640px) {
  .tab {
    padding: 10px 8px;
    gap: 4px;
  }

  .tab-label {
    font-size: 11px;
  }

  .tab-description {
    display: none;
  }

  .tab-icon .size-4 {
    width: 14px;
    height: 14px;
  }
}
</style>
