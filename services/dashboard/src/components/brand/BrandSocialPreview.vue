<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { Check, Copy, Download } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import { failureToast } from '@/components/ui/sonner';
import { Textarea } from '@/components/ui/textarea';
import SegmentedControl from '@/components/SegmentedControl.vue';
import { CREAM, ESPRESSO, INK, LILAC, LOCKUP_HEIGHT, LOCKUP_WIDTH, PAPER, PURPLE, brandSvgUrl, lockupAsset } from '@/lib/brand';

interface Theme {
  text: string;
  soft: string;
  mark: string;
  wordmark: string;
  kicker: string;
  kickerText: string;
}

interface Background {
  name: string;
  theme: Theme;
  src?: string;
  thumb?: string;
  color?: string;
}

const MARGIN = 72;
const LOCKUP_TOP = 64;
const LOCKUP_SIZE = 48;
const HEADLINE_SIZE = 92;
const TAGLINE_SIZE = 30;
const KICKER_SIZE = 22;
const GAP = 24;

const formats = [
  { value: 'wide' as const, label: 'Wide', width: 1280, height: 640, scale: 1, preview: 'w-full', file: 'lucity-social-preview', hint: '1280 × 640, for GitHub releases and link previews' },
  { value: 'square' as const, label: 'Square', width: 1080, height: 1080, scale: 1.1, preview: 'w-1/2', file: 'lucity-social-preview-square', hint: '1080 × 1080, for social posts' },
  { value: 'portrait' as const, label: 'Portrait', width: 1080, height: 1350, scale: 1.2, preview: 'w-2/5', file: 'lucity-social-preview-portrait', hint: '1080 × 1350, for social posts on phones' },
];

type FormatKey = (typeof formats)[number]['value'];

const photoTheme: Theme = { text: CREAM, soft: 'rgba(246, 237, 224, 0.8)', mark: LILAC, wordmark: CREAM, kicker: LILAC, kickerText: ESPRESSO };

const photos = import.meta.glob<string>('../../../assets/img/*.webp', { eager: true, import: 'default' });
const thumbs = import.meta.glob<string>('../../../assets/img/thumbs/*.webp', { eager: true, import: 'default' });

function wallpaper(name: string, file: string): Background {
  return { name, theme: photoTheme, src: photos[`../../../assets/img/${file}.webp`], thumb: thumbs[`../../../assets/img/thumbs/${file}.webp`] };
}

const backgrounds: Background[] = [
  wallpaper('Alps harbor', 'alps_harbor'),
  wallpaper('Mountain city', 'mountain_city'),
  wallpaper('Night city', 'mountain_city_night'),
  wallpaper('Container harbor', 'container_harbor'),
  wallpaper('Mountain ship', 'mountain_ship'),
  wallpaper('Mountain plane', 'mountain_plane'),
  wallpaper('Branching river', 'branching_river'),
  wallpaper('Lake', 'lake'),
  wallpaper('Cablecar', 'cablecar'),
  wallpaper('Hot air balloon', 'hotair_baloon'),
  wallpaper('Parked plane', 'plane_parked'),
  wallpaper('Octopus river', 'octopus_river'),
  wallpaper('Gopher ship', 'gopher_ship'),
  { name: 'Cream', color: CREAM, theme: { text: INK, soft: '#7a6250', mark: PURPLE, wordmark: INK, kicker: PURPLE, kickerText: CREAM } },
  { name: 'Espresso', color: ESPRESSO, theme: { text: CREAM, soft: 'rgba(246, 237, 224, 0.75)', mark: LILAC, wordmark: PAPER, kicker: LILAC, kickerText: ESPRESSO } },
  { name: 'Purple', color: PURPLE, theme: { text: CREAM, soft: 'rgba(246, 237, 224, 0.85)', mark: CREAM, wordmark: CREAM, kicker: CREAM, kickerText: PURPLE } },
];

const canvasRef = ref<HTMLCanvasElement | null>(null);
const formatKey = ref<FormatKey>('wide');
const selected = ref(8);
const kicker = ref('');
const headline = ref('The PaaS you can leave.');
const tagline = ref('Made in Switzerland · Hosted in the EU · Open source');
const shade = ref(1);
const copied = ref(false);
const images = new Map<string, HTMLImageElement>();
let copiedTimer: ReturnType<typeof setTimeout> | undefined;

const format = computed(() => formats.find(item => item.value === formatKey.value)!);
const background = computed(() => backgrounds[selected.value]!);

function image(src: string) {
  const cached = images.get(src);
  if (cached) return cached.complete && cached.naturalWidth > 0 ? cached : undefined;
  const img = new Image();
  img.onload = () => draw();
  img.src = src;
  images.set(src, img);
  return undefined;
}

function lockupUrl(theme: Theme, scale: number) {
  const height = LOCKUP_SIZE * scale;
  return brandSvgUrl(lockupAsset('lockup', theme.mark, theme.wordmark), height * LOCKUP_WIDTH / LOCKUP_HEIGHT, height);
}

function drawCover(ctx: CanvasRenderingContext2D, img: HTMLImageElement, width: number, height: number) {
  const scale = Math.max(width / img.width, height / img.height);
  const sourceWidth = width / scale;
  const sourceHeight = height / scale;
  ctx.drawImage(img, (img.width - sourceWidth) / 2, (img.height - sourceHeight) / 2, sourceWidth, sourceHeight, 0, 0, width, height);
}

function drawShade(ctx: CanvasRenderingContext2D, width: number, height: number) {
  const gradient = ctx.createLinearGradient(0, 0, 0, height);
  for (const [stop, alpha] of [[0, 0.55], [0.28, 0.12], [0.55, 0.3], [1, 0.85]] as const) {
    gradient.addColorStop(stop, `rgba(33, 18, 8, ${Math.min(1, alpha * shade.value)})`);
  }
  ctx.fillStyle = gradient;
  ctx.fillRect(0, 0, width, height);
}

function fitFont(ctx: CanvasRenderingContext2D, lines: string[], size: number, font: (size: number) => string, maxWidth: number) {
  ctx.font = font(size);
  const widest = Math.max(...lines.map(line => ctx.measureText(line).width));
  const fitted = widest > maxWidth ? size * maxWidth / widest : size;
  ctx.font = font(fitted);
  return fitted;
}

function tileClass(item: Background, index: number) {
  if (selected.value === index) return 'border-primary';
  return item.color ? 'border-border' : 'border-transparent hover:border-border';
}

function draw() {
  const ctx = canvasRef.value?.getContext('2d');
  if (!ctx) return;
  const { theme, src, color } = background.value;
  const { scale } = format.value;
  const photo = src ? image(src) : undefined;
  const lockup = image(lockupUrl(theme, scale));
  if ((src && !photo) || !lockup) return;

  const width = format.value.width / scale;
  const height = format.value.height / scale;
  ctx.setTransform(scale, 0, 0, scale, 0, 0);

  if (photo) {
    drawCover(ctx, photo, width, height);
    drawShade(ctx, width, height);
  } else {
    ctx.fillStyle = color!;
    ctx.fillRect(0, 0, width, height);
  }

  ctx.drawImage(lockup, MARGIN, LOCKUP_TOP, LOCKUP_SIZE * LOCKUP_WIDTH / LOCKUP_HEIGHT, LOCKUP_SIZE);

  const maxWidth = width - 2 * MARGIN;
  ctx.textBaseline = 'top';
  ctx.textAlign = 'left';
  let bottom = height - MARGIN;

  if (tagline.value) {
    const size = fitFont(ctx, [tagline.value], TAGLINE_SIZE, s => `300 ${s}px Sora`, maxWidth);
    bottom -= size * 1.3;
    ctx.fillStyle = theme.soft;
    ctx.fillText(tagline.value, MARGIN, bottom + size * 0.15);
    bottom -= GAP;
  }

  const kickerHeight = kicker.value ? KICKER_SIZE * 1.2 + 16 + GAP : 0;
  const lines = headline.value.split('\n');
  const room = bottom - kickerHeight - (LOCKUP_TOP + LOCKUP_SIZE + 2 * GAP);
  const headlineFont = (s: number) => `${s}px "Redaction 20"`;
  const size = Math.min(fitFont(ctx, lines, HEADLINE_SIZE, headlineFont, maxWidth), room / (lines.length * 1.02));
  ctx.font = headlineFont(size);
  ctx.fillStyle = theme.text;
  bottom -= lines.length * size * 1.02;
  lines.forEach((line, i) => ctx.fillText(line, MARGIN, bottom + i * size * 1.02));

  if (kicker.value) {
    ctx.font = `600 ${KICKER_SIZE}px Sora`;
    const pillWidth = Math.min(ctx.measureText(kicker.value).width + 36, maxWidth);
    const pillHeight = KICKER_SIZE * 1.2 + 16;
    const top = bottom - GAP - pillHeight;
    ctx.fillStyle = theme.kicker;
    ctx.beginPath();
    ctx.roundRect(MARGIN, top, pillWidth, pillHeight, pillHeight / 2);
    ctx.fill();
    ctx.fillStyle = theme.kickerText;
    ctx.fillText(kicker.value, MARGIN + 18, top + 8 + KICKER_SIZE * 0.1, pillWidth - 36);
  }
}

function download(type: 'png' | 'jpg') {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const link = document.createElement('a');
  link.download = `${format.value.file}.${type}`;
  link.href = canvas.toDataURL(type === 'png' ? 'image/png' : 'image/jpeg', 0.95);
  link.click();
}

async function copyImage() {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const png = new Promise<Blob>((resolve, reject) => {
    canvas.toBlob(blob => (blob ? resolve(blob) : reject(new Error('The image could not be rendered.'))), 'image/png');
  });
  try {
    await navigator.clipboard.write([new ClipboardItem({ 'image/png': png })]);
    copied.value = true;
    clearTimeout(copiedTimer);
    copiedTimer = setTimeout(() => {
      copied.value = false;
    }, 1500);
  } catch (error) {
    failureToast('Failed to copy image', error);
  }
}

watch([formatKey, selected, kicker, headline, tagline, shade], () => draw(), { flush: 'post' });

onMounted(async () => {
  await Promise.all([
    document.fonts.load(`${HEADLINE_SIZE}px "Redaction 20"`),
    document.fonts.load(`300 ${TAGLINE_SIZE}px Sora`),
    document.fonts.load(`600 ${KICKER_SIZE}px Sora`),
  ]);
  draw();
});
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-border bg-card">
    <div class="bg-muted">
      <canvas
        ref="canvasRef"
        :width="format.width"
        :height="format.height"
        class="mx-auto block"
        :class="format.preview"
      />
    </div>
    <div class="grid gap-6 border-t border-border p-4 sm:p-6 md:grid-cols-[1fr_auto]">
      <div class="space-y-6">
        <div>
          <p class="mb-3 text-sm font-medium text-foreground">
            Background
          </p>
          <div class="grid grid-cols-4 gap-2 sm:grid-cols-8">
            <button
              v-for="(item, i) in backgrounds"
              :key="item.name"
              type="button"
              class="aspect-video overflow-hidden rounded-sm border-2 transition-colors"
              :class="tileClass(item, i)"
              :style="item.color ? { background: item.color } : undefined"
              :aria-label="item.name"
              :aria-pressed="selected === i"
              @click="selected = i"
            >
              <img
                v-if="item.thumb"
                :src="item.thumb"
                alt=""
                class="size-full object-cover"
              >
            </button>
          </div>
        </div>
        <div class="space-y-3">
          <Label>Shade</Label>
          <Slider
            :model-value="[shade]"
            :min="0.4"
            :max="1.4"
            :step="0.05"
            :disabled="!background.src"
            @update:model-value="shade = $event?.[0] ?? shade"
          />
        </div>
      </div>
      <div class="space-y-4 md:w-72">
        <div class="flex flex-col items-start gap-2">
          <Label>Format</Label>
          <SegmentedControl v-model="formatKey" label="Format" :options="formats" />
          <p class="text-xs text-muted-foreground">
            {{ format.hint }}
          </p>
        </div>
        <div class="space-y-2">
          <Label for="social-kicker">Kicker</Label>
          <Input
            id="social-kicker"
            v-model="kicker"
            placeholder="Optional, like a release version"
          />
        </div>
        <div class="space-y-2">
          <Label for="social-headline">Headline</Label>
          <Textarea
            id="social-headline"
            v-model="headline"
            rows="2"
            class="resize-none text-sm"
          />
        </div>
        <div class="space-y-2">
          <Label for="social-tagline">Tagline</Label>
          <Input id="social-tagline" v-model="tagline" />
        </div>
        <div class="grid grid-cols-2 gap-2">
          <Button
            variant="outline"
            class="col-span-2"
            @click="copyImage"
          >
            <Check v-if="copied" class="text-status-ok" />
            <Copy v-else />
            {{ copied ? 'Copied' : 'Copy image' }}
          </Button>
          <Button
            variant="outline"
            @click="download('png')"
          >
            <Download />
            PNG
          </Button>
          <Button
            variant="outline"
            @click="download('jpg')"
          >
            <Download />
            JPG
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
