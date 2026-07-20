<script setup lang="ts">
import { computed } from 'vue';
import Anser from 'anser';

const props = defineProps<{
  text: string;
}>();

type Segment = ReturnType<typeof Anser.ansiToJson>[number] & {
  isInverted?: boolean;
};

const segments = computed(() =>
  Anser.ansiToJson(props.text, { use_classes: false, remove_empty: true }),
);

function styleFor(segment: Segment) {
  const style: Record<string, string> = {};

  const foreground = segment.fg_truecolor ?? segment.fg;
  const background = segment.bg_truecolor ?? segment.bg;
  const [color, backgroundColor] = segment.isInverted
    ? [background, foreground]
    : [foreground, background];

  if (color) style.color = `rgb(${color})`;
  if (backgroundColor) style.backgroundColor = `rgb(${backgroundColor})`;

  const decorations = segment.decorations ?? [];
  if (decorations.includes('bold')) style.fontWeight = '600';
  if (decorations.includes('dim')) style.opacity = '0.6';
  if (decorations.includes('italic')) style.fontStyle = 'italic';
  if (decorations.includes('hidden')) style.opacity = '0';

  const textDecoration: string[] = [];
  if (decorations.includes('underline')) textDecoration.push('underline');
  if (decorations.includes('strikethrough')) textDecoration.push('line-through');
  if (textDecoration.length) style.textDecoration = textDecoration.join(' ');

  return style;
}
</script>

<template>
  <span class="whitespace-pre-wrap break-all">
    <span
      v-for="(segment, idx) in segments"
      :key="idx"
      :style="styleFor(segment)"
      >{{ segment.content }}</span
    >
  </span>
</template>
