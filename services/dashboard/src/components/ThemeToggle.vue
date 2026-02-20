<script setup lang="ts">
import { computed } from 'vue';
import { useTheme } from '@/composables/useTheme';

const { theme, toggleTheme } = useTheme();
const isDark = computed(() => theme.value === 'dark');
</script>

<template>
  <button
    class="theme-toggle"
    :class="{ 'is-dark': isDark }"
    :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    @click="toggleTheme"
  >
    <span class="theme-toggle__track">
      <!-- Sky / stars decoration -->
      <span class="theme-toggle__stars">
        <span class="star star--1" />
        <span class="star star--2" />
        <span class="star star--3" />
      </span>

      <!-- Cloud puffs (light mode) -->
      <span class="theme-toggle__clouds">
        <span class="cloud cloud--1" />
        <span class="cloud cloud--2" />
      </span>

      <!-- The celestial body (sun/moon) -->
      <span class="theme-toggle__thumb">
        <!-- Sun rays -->
        <span class="thumb__rays" />
        <!-- Moon craters -->
        <span class="thumb__crater thumb__crater--1" />
        <span class="thumb__crater thumb__crater--2" />
        <span class="thumb__crater thumb__crater--3" />
      </span>
    </span>
  </button>
</template>

<style scoped>
.theme-toggle {
  --toggle-w: 52px;
  --toggle-h: 26px;
  --thumb-size: 20px;
  --thumb-offset: 3px;
  --travel: calc(var(--toggle-w) - var(--thumb-size) - var(--thumb-offset) * 2);

  position: relative;
  width: var(--toggle-w);
  height: var(--toggle-h);
  padding: 0;
  border: none;
  background: none;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  outline: none;
}

.theme-toggle:focus-visible .theme-toggle__track {
  box-shadow: 0 0 0 2px var(--background), 0 0 0 4px var(--ring);
}

/* ── Track ── */
.theme-toggle__track {
  display: flex;
  align-items: center;
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 999px;
  overflow: hidden;
  background: linear-gradient(135deg, oklch(0.72 0.14 220), oklch(0.60 0.16 250));
  transition: background 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.is-dark .theme-toggle__track {
  background: linear-gradient(135deg, oklch(0.22 0.06 270), oklch(0.16 0.08 280));
}

/* ── Stars ── */
.theme-toggle__stars {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.star {
  position: absolute;
  border-radius: 50%;
  background: white;
  opacity: 0;
  transform: scale(0);
  transition: opacity 0.4s ease, transform 0.4s ease;
}

.is-dark .star {
  opacity: 1;
  transform: scale(1);
}

.star--1 { width: 2px; height: 2px; top: 6px;  left: 10px; transition-delay: 0.15s; }
.star--2 { width: 2.5px; height: 2.5px; top: 14px; left: 16px; transition-delay: 0.25s; }
.star--3 { width: 1.5px; height: 1.5px; top: 9px;  left: 22px; transition-delay: 0.35s; }

.is-dark .star--1 {
  animation: twinkle 3s ease-in-out 0.5s infinite;
}
.is-dark .star--2 {
  animation: twinkle 3s ease-in-out 1.2s infinite;
}
.is-dark .star--3 {
  animation: twinkle 3s ease-in-out 2s infinite;
}

@keyframes twinkle {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* ── Clouds ── */
.theme-toggle__clouds {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.cloud {
  position: absolute;
  border-radius: 999px;
  background: white;
  opacity: 0.6;
  transition: opacity 0.4s ease, transform 0.4s ease;
}

.is-dark .cloud {
  opacity: 0;
  transform: translateX(4px);
}

.cloud--1 { width: 12px; height: 5px; bottom: 5px; right: 6px; }
.cloud--2 { width: 8px; height: 4px; bottom: 9px; right: 14px; opacity: 0.4; }

/* ── Thumb (celestial body) ── */
.theme-toggle__thumb {
  position: absolute;
  left: var(--thumb-offset);
  width: var(--thumb-size);
  height: var(--thumb-size);
  border-radius: 50%;
  background: oklch(0.95 0.08 90);
  box-shadow:
    0 1px 3px oklch(0 0 0 / 0.2),
    inset 0 -1px 2px oklch(0 0 0 / 0.05);
  transition:
    transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1),
    background 0.4s ease,
    box-shadow 0.4s ease;
  z-index: 1;
}

.is-dark .theme-toggle__thumb {
  transform: translateX(var(--travel));
  background: oklch(0.88 0.04 80);
  box-shadow:
    0 1px 3px oklch(0 0 0 / 0.3),
    inset 0 -1px 2px oklch(0 0 0 / 0.1);
}

/* ── Sun rays ── */
.thumb__rays {
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  background: transparent;
  box-shadow: 0 0 8px 2px oklch(0.90 0.12 90 / 0.5);
  opacity: 1;
  transition: opacity 0.3s ease;
}

.is-dark .thumb__rays {
  opacity: 0;
}

/* ── Moon craters ── */
.thumb__crater {
  position: absolute;
  border-radius: 50%;
  background: oklch(0.78 0.03 80);
  opacity: 0;
  transform: scale(0);
  transition: opacity 0.3s ease 0.1s, transform 0.3s ease 0.1s;
}

.is-dark .thumb__crater {
  opacity: 1;
  transform: scale(1);
}

.thumb__crater--1 { width: 5px; height: 5px; top: 4px;  left: 4px; }
.thumb__crater--2 { width: 3px; height: 3px; top: 11px; left: 8px; }
.thumb__crater--3 { width: 3.5px; height: 3.5px; top: 5px; left: 11px; }

/* ── Hover ── */
.theme-toggle:hover .theme-toggle__thumb {
  box-shadow:
    0 2px 8px oklch(0 0 0 / 0.2),
    inset 0 -1px 2px oklch(0 0 0 / 0.05);
}

.theme-toggle:hover .thumb__rays {
  box-shadow: 0 0 12px 3px oklch(0.90 0.12 90 / 0.6);
}

.is-dark.theme-toggle:hover .theme-toggle__thumb {
  box-shadow:
    0 2px 8px oklch(0 0 0 / 0.3),
    0 0 10px 2px oklch(0.80 0.08 250 / 0.2),
    inset 0 -1px 2px oklch(0 0 0 / 0.1);
}

/* ── Active press ── */
.theme-toggle:active .theme-toggle__thumb {
  transform: translateX(0) scale(0.9);
}

.is-dark .theme-toggle:active .theme-toggle__thumb {
  transform: translateX(var(--travel)) scale(0.9);
}
</style>
