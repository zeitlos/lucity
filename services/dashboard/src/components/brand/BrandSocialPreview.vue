<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue';
import { Download } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { LILAC } from '@/lib/brand';
import { LOGO_HEIGHT, LOGO_PATH, LOGO_WIDTH } from '@/lib/logo';

import alpsHarborImg from '../../../assets/img/alps_harbor.webp';
import mountainCityImg from '../../../assets/img/mountain_city.webp';
import mountainCityNightImg from '../../../assets/img/mountain_city_night.webp';
import branchingRiverImg from '../../../assets/img/branching_river.webp';
import containerHarborImg from '../../../assets/img/container_harbor.webp';
import mountainShipImg from '../../../assets/img/mountain_ship.webp';
import mountainPlaneImg from '../../../assets/img/mountain_plane.webp';
import cablecarImg from '../../../assets/img/cablecar.webp';
import lakeImg from '../../../assets/img/lake.webp';
import hotairBaloonImg from '../../../assets/img/hotair_baloon.webp';
import planeParkedImg from '../../../assets/img/plane_parked.webp';
import octopusRiverImg from '../../../assets/img/octopus_river.webp';
import gopherShipImg from '../../../assets/img/gopher_ship.webp';

const canvasRef = ref<HTMLCanvasElement | null>(null);
const selectedWallpaper = ref(0);
const headline = ref('The PaaS you can leave.');
const tagline = ref('Open-source · Kubernetes-native · Ejectable');
const loadedImages = new Map<string, HTMLImageElement>();

const wallpapers = [
  { name: 'Alps Harbor', src: alpsHarborImg },
  { name: 'Mountain City', src: mountainCityImg },
  { name: 'Night City', src: mountainCityNightImg },
  { name: 'Container Harbor', src: containerHarborImg },
  { name: 'Mountain Ship', src: mountainShipImg },
  { name: 'Mountain Plane', src: mountainPlaneImg },
  { name: 'Branching River', src: branchingRiverImg },
  { name: 'Lake', src: lakeImg },
  { name: 'Cablecar', src: cablecarImg },
  { name: 'Hot Air Balloon', src: hotairBaloonImg },
  { name: 'Parked Plane', src: planeParkedImg },
  { name: 'Octopus River', src: octopusRiverImg },
  { name: 'Gopher Ship', src: gopherShipImg },
];

function preloadWallpapers() {
  wallpapers.forEach((wp) => {
    const img = new Image();
    img.onload = () => {
      loadedImages.set(wp.src, img);
      if (wallpapers[selectedWallpaper.value]?.src === wp.src) drawCanvas();
    };
    img.src = wp.src;
  });
}

function drawLogo(ctx: CanvasRenderingContext2D, cx: number, cy: number, logoHeight: number) {
  const scale = logoHeight / LOGO_HEIGHT;
  ctx.save();
  ctx.translate(cx - (LOGO_WIDTH * scale) / 2, cy - logoHeight / 2);
  ctx.scale(scale, scale);
  ctx.fillStyle = LILAC;
  ctx.fill(new Path2D(LOGO_PATH));
  ctx.restore();
}

function drawCanvas() {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  const W = 1280;
  const H = 640;
  ctx.clearRect(0, 0, W, H);

  const wp = wallpapers[selectedWallpaper.value]!;
  const img = loadedImages.get(wp.src);

  if (img) {
    const imgRatio = img.width / img.height;
    const canvasRatio = W / H;
    let sx = 0;
    let sy = 0;
    let sw = img.width;
    let sh = img.height;
    if (imgRatio > canvasRatio) {
      sw = img.height * canvasRatio;
      sx = (img.width - sw) / 2;
    } else {
      sh = img.width / canvasRatio;
      sy = (img.height - sh) / 2;
    }
    ctx.drawImage(img, sx, sy, sw, sh, 0, 0, W, H);
  } else {
    const grad = ctx.createLinearGradient(0, 0, W, H);
    grad.addColorStop(0, '#1a3a2f');
    grad.addColorStop(1, '#0d1b16');
    ctx.fillStyle = grad;
    ctx.fillRect(0, 0, W, H);
  }

  const vig = ctx.createRadialGradient(W / 2, H / 2, H * 0.1, W / 2, H / 2, W * 0.72);
  vig.addColorStop(0, 'rgba(0,0,0,0.3)');
  vig.addColorStop(1, 'rgba(0,0,0,0.7)');
  ctx.fillStyle = vig;
  ctx.fillRect(0, 0, W, H);

  drawLogo(ctx, W / 2, H / 2 - 105, 84);

  ctx.fillStyle = 'white';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';
  ctx.font = '88px "Redaction 20", Georgia, serif';
  ctx.fillText(headline.value, W / 2, H / 2 + 5);

  ctx.font = '300 32px "Sora", system-ui, sans-serif';
  ctx.fillStyle = 'rgba(255,255,255,0.7)';
  ctx.fillText(tagline.value, W / 2, H / 2 + 110);
}

function download(format: 'png' | 'jpg') {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const mimeType = format === 'png' ? 'image/png' : 'image/jpeg';
  const link = document.createElement('a');
  link.download = `lucity-social-preview.${format}`;
  link.href = canvas.toDataURL(mimeType, 0.95);
  link.click();
}

watch([selectedWallpaper, headline, tagline], () => drawCanvas());

onMounted(async () => {
  preloadWallpapers();
  await document.fonts.load('88px "Redaction 20"');
  await document.fonts.ready;
  await nextTick();
  drawCanvas();
});
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-border bg-card">
    <canvas
      ref="canvasRef"
      width="1280"
      height="640"
      class="block w-full"
    />
    <div class="grid gap-6 border-t border-border p-4 sm:p-6 md:grid-cols-[1fr_auto]">
      <div>
        <p class="mb-3 text-sm font-medium text-foreground">
          Background
        </p>
        <div class="grid grid-cols-4 gap-2 sm:grid-cols-7">
          <button
            v-for="(wp, i) in wallpapers"
            :key="wp.name"
            class="group relative overflow-hidden rounded-sm border-2 transition-all"
            :class="selectedWallpaper === i ? 'border-primary' : 'border-transparent hover:border-border'"
            :aria-label="wp.name"
            @click="selectedWallpaper = i"
          >
            <img
              :src="wp.src"
              :alt="wp.name"
              class="aspect-video w-full object-cover"
            >
          </button>
        </div>
      </div>
      <div class="space-y-4 md:w-64">
        <div class="space-y-2">
          <Label for="social-headline">Headline</Label>
          <Input id="social-headline" v-model="headline" />
        </div>
        <div class="space-y-2">
          <Label for="social-tagline">Tagline</Label>
          <Input id="social-tagline" v-model="tagline" />
        </div>
        <div class="flex gap-2">
          <Button
            variant="outline"
            class="flex-1"
            @click="download('png')"
          >
            <Download />
            PNG
          </Button>
          <Button
            variant="outline"
            class="flex-1"
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
