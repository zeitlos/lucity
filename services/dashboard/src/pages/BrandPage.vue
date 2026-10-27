<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue';
import BaseLogo from '@/components/BaseLogo.vue';
import { LOGO_HEIGHT, LOGO_PATH, LOGO_WIDTH } from '@/lib/logo';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { AlertCircle, Check, Download, Rocket, Zap } from '@lucide/vue';

import alpsHarborImg from '../../assets/img/alps_harbor.webp';
import mountainCityImg from '../../assets/img/mountain_city.webp';
import mountainCityNightImg from '../../assets/img/mountain_city_night.webp';
import branchingRiverImg from '../../assets/img/branching_river.webp';
import containerHarborImg from '../../assets/img/container_harbor.webp';
import mountainShipImg from '../../assets/img/mountain_ship.webp';
import mountainPlaneImg from '../../assets/img/mountain_plane.webp';
import cablecarImg from '../../assets/img/cablecar.webp';
import lakeImg from '../../assets/img/lake.webp';
import hotairBaloonImg from '../../assets/img/hotair_baloon.webp';
import planeParkedImg from '../../assets/img/plane_parked.webp';
import octopusRiverImg from '../../assets/img/octopus_river.webp';
import gopherShipImg from '../../assets/img/gopher_ship.webp';

const colors = [
  { name: 'Mark', var: '--brand-mark', fg: '--primary-foreground', desc: 'Logo purple' },
  { name: 'Primary', var: '--primary', fg: '--primary-foreground', desc: 'Brand teal' },
  { name: 'Secondary', var: '--secondary', fg: '--secondary-foreground', desc: 'Warm beige' },
  { name: 'Accent', var: '--accent', fg: '--accent-foreground', desc: 'Soft warm' },
  { name: 'Destructive', var: '--destructive', fg: '--destructive-foreground', desc: 'Danger red' },
  { name: 'Muted', var: '--muted', fg: '--muted-foreground', desc: 'Subdued elements' },
];

const semanticColors = [
  { name: 'Background', var: '--background' },
  { name: 'Foreground', var: '--foreground' },
  { name: 'Card', var: '--card' },
  { name: 'Popover', var: '--popover' },
  { name: 'Border', var: '--border' },
  { name: 'Input', var: '--input' },
  { name: 'Ring', var: '--ring' },
];

const sidebarColors = [
  { name: 'Sidebar', var: '--sidebar' },
  { name: 'Sidebar Foreground', var: '--sidebar-foreground' },
  { name: 'Sidebar Primary', var: '--sidebar-primary' },
  { name: 'Sidebar Accent', var: '--sidebar-accent' },
  { name: 'Sidebar Border', var: '--sidebar-border' },
];

const chartColors = [
  { name: 'Chart 1', var: '--chart-1' },
  { name: 'Chart 2', var: '--chart-2' },
  { name: 'Chart 3', var: '--chart-3' },
  { name: 'Chart 4', var: '--chart-4' },
  { name: 'Chart 5', var: '--chart-5' },
];

// --- Social Preview Canvas ---

const canvasRef = ref<HTMLCanvasElement | null>(null);
const selectedWallpaper = ref(0);
const previewHeadline = ref('The PaaS you can leave.');
const previewTagline = ref('Open-source · Kubernetes-native · Ejectable');
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
  ctx.fillStyle = '#a274ee';
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

  // Background wallpaper
  const wp = wallpapers[selectedWallpaper.value]!;
  const img = loadedImages.get(wp.src);

  if (img) {
    // Cover fit
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

  // Vignette overlay — stronger for text legibility
  const vig = ctx.createRadialGradient(W / 2, H / 2, H * 0.1, W / 2, H / 2, W * 0.72);
  vig.addColorStop(0, 'rgba(0,0,0,0.3)');
  vig.addColorStop(1, 'rgba(0,0,0,0.7)');
  ctx.fillStyle = vig;
  ctx.fillRect(0, 0, W, H);

  drawLogo(ctx, W / 2, H / 2 - 105, 84);

  // Headline — large enough to read at card size (~400px wide display)
  ctx.fillStyle = 'white';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';
  ctx.font = '88px "Redaction 20", Georgia, serif';
  ctx.fillText(previewHeadline.value, W / 2, H / 2 + 5);

  // Tagline
  ctx.font = '300 32px "Sora", system-ui, sans-serif';
  ctx.fillStyle = 'rgba(255,255,255,0.7)';
  ctx.fillText(previewTagline.value, W / 2, H / 2 + 110);
}

function downloadPreview(format: 'png' | 'jpg') {
  const canvas = canvasRef.value;
  if (!canvas) return;
  const mimeType = format === 'png' ? 'image/png' : 'image/jpeg';
  const link = document.createElement('a');
  link.download = `lucity-social-preview.${format}`;
  link.href = canvas.toDataURL(mimeType, 0.95);
  link.click();
}

watch([selectedWallpaper, previewHeadline, previewTagline], () => drawCanvas());

onMounted(async () => {
  preloadWallpapers();
  await document.fonts.load('88px "Redaction 20"');
  await document.fonts.ready;
  await nextTick();
  drawCanvas();
});
</script>

<template>
  <div class="min-h-screen bg-background">
    <div class="mx-auto max-w-5xl px-8 py-12">
      <div class="mb-4 flex items-center gap-4">
        <BaseLogo :size="36" />
        <div>
          <h1 class="font-serif text-4xl text-foreground">
            Brand
          </h1>
          <p class="text-muted-foreground">
            Logo, colors, typography, and components
          </p>
        </div>
      </div>

      <Separator class="mb-12" />

      <!-- Logo -->
      <section class="mb-16">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Logo
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          A folded sheet with two bends, extruded unevenly so top and bottom never mirror each other. Purple on light, lilac on dark.
        </p>

        <div class="grid grid-cols-3 gap-8">
          <Card>
            <CardHeader>
              <CardTitle>Mark</CardTitle>
              <CardDescription>
                Standalone, wherever space is tight
              </CardDescription>
            </CardHeader>
            <CardContent class="flex h-40 items-center justify-center">
              <BaseLogo :size="96" />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Lockup</CardTitle>
              <CardDescription>
                Mark and wordmark in Redaction 35
              </CardDescription>
            </CardHeader>
            <CardContent class="flex h-40 items-center justify-center gap-[30px]">
              <BaseLogo :size="35" />
              <span class="font-wordmark text-4xl leading-none text-foreground">Lucity</span>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>App icon</CardTitle>
              <CardDescription>
                Cream mark on a purple tile
              </CardDescription>
            </CardHeader>
            <CardContent class="flex h-40 items-center justify-center">
              <div class="app-icon flex size-28 items-center justify-center rounded-[22%]">
                <BaseLogo :size="58" />
              </div>
            </CardContent>
          </Card>
        </div>

        <div class="mt-8">
          <h3 class="mb-4 text-sm font-medium text-muted-foreground">
            Sizes
          </h3>
          <div class="flex items-end gap-8">
            <div
              v-for="size in [12, 16, 20, 24, 32, 48, 64]"
              :key="size"
              class="flex flex-col items-center gap-2"
            >
              <BaseLogo :size="size" />
              <span class="font-mono text-xs text-muted-foreground">{{ size }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Colors -->
      <section class="mb-16">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Colors
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          OKLCH color space. All tokens adapt to light and dark mode via CSS custom properties.
        </p>

        <div class="space-y-8">
          <!-- Brand colors -->
          <div>
            <h3 class="mb-4 text-sm font-medium text-muted-foreground">
              Brand
            </h3>
            <div class="grid grid-cols-6 gap-4">
              <div
                v-for="color in colors"
                :key="color.var"
                class="space-y-2"
              >
                <div
                  class="flex h-20 items-end rounded-lg p-3"
                  :style="{ background: `var(${color.var})` }"
                >
                  <span
                    class="font-mono text-xs"
                    :style="{ color: `var(${color.fg})` }"
                  >
                    {{ color.var }}
                  </span>
                </div>
                <div>
                  <p class="text-sm font-medium text-foreground">{{ color.name }}</p>
                  <p class="text-xs text-muted-foreground">{{ color.desc }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Semantic colors -->
          <div>
            <h3 class="mb-4 text-sm font-medium text-muted-foreground">
              Semantic
            </h3>
            <div class="grid grid-cols-7 gap-3">
              <div
                v-for="color in semanticColors"
                :key="color.var"
                class="space-y-1.5"
              >
                <div
                  class="h-12 rounded-lg border border-border"
                  :style="{ background: `var(${color.var})` }"
                />
                <p class="truncate text-xs text-muted-foreground">{{ color.name }}</p>
              </div>
            </div>
          </div>

          <!-- Sidebar colors -->
          <div>
            <h3 class="mb-4 text-sm font-medium text-muted-foreground">
              Sidebar
            </h3>
            <div class="grid grid-cols-5 gap-3">
              <div
                v-for="color in sidebarColors"
                :key="color.var"
                class="space-y-1.5"
              >
                <div
                  class="h-12 rounded-lg border border-border"
                  :style="{ background: `var(${color.var})` }"
                />
                <p class="truncate text-xs text-muted-foreground">{{ color.name }}</p>
              </div>
            </div>
          </div>

          <!-- Chart colors -->
          <div>
            <h3 class="mb-4 text-sm font-medium text-muted-foreground">
              Chart
            </h3>
            <div class="flex gap-0 overflow-hidden rounded-lg">
              <div
                v-for="color in chartColors"
                :key="color.var"
                class="h-10 flex-1"
                :style="{ background: `var(${color.var})` }"
              />
            </div>
          </div>
        </div>
      </section>

      <!-- Typography -->
      <section class="mb-16">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Typography
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          Sora for headings and body. Redaction for display. Fira Code for monospace.
        </p>

        <Card>
          <CardContent class="space-y-6 pt-6">
            <div class="space-y-3">
              <p class="font-serif text-5xl text-foreground">The PaaS you can leave.</p>
              <p class="text-3xl font-bold text-foreground">The quick brown fox</p>
              <p class="text-2xl font-semibold text-foreground">The quick brown fox</p>
              <p class="text-xl font-medium text-foreground">The quick brown fox</p>
              <p class="text-base text-foreground">The quick brown fox jumps over the lazy dog.</p>
              <p class="text-sm text-muted-foreground">The quick brown fox jumps over the lazy dog.</p>
              <p class="text-xs text-muted-foreground">The quick brown fox jumps over the lazy dog.</p>
            </div>
            <Separator />
            <div>
              <p class="mb-1 text-xs text-muted-foreground">Monospace</p>
              <p class="font-mono text-sm text-foreground">
                const deploy = (env: string) =&gt; argocd.sync(env);
              </p>
            </div>
          </CardContent>
        </Card>
      </section>

      <!-- Components -->
      <section class="mb-16">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Components
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          shadcn-vue primitives styled with the brand tokens.
        </p>

        <Tabs default-value="buttons">
          <TabsList>
            <TabsTrigger value="buttons">
              Buttons
            </TabsTrigger>
            <TabsTrigger value="inputs">
              Inputs
            </TabsTrigger>
            <TabsTrigger value="feedback">
              Feedback
            </TabsTrigger>
            <TabsTrigger value="cards">
              Cards
            </TabsTrigger>
          </TabsList>

          <TabsContent
            value="buttons"
            class="mt-6"
          >
            <Card>
              <CardContent class="space-y-6 pt-6">
                <div>
                  <p class="mb-3 text-sm font-medium text-muted-foreground">Variants</p>
                  <div class="flex flex-wrap gap-3">
                    <Button>Default</Button>
                    <Button variant="secondary">
                      Secondary
                    </Button>
                    <Button variant="outline">
                      Outline
                    </Button>
                    <Button variant="ghost">
                      Ghost
                    </Button>
                    <Button variant="link">
                      Link
                    </Button>
                    <Button variant="destructive">
                      Destructive
                    </Button>
                  </div>
                </div>
                <Separator />
                <div>
                  <p class="mb-3 text-sm font-medium text-muted-foreground">Sizes</p>
                  <div class="flex items-center gap-3">
                    <Button>Default</Button>
                    <Button size="xs">
                      Extra small
                    </Button>
                    <Button size="icon" aria-label="Deploy">
                      <Rocket />
                    </Button>
                    <Button size="icon-sm" variant="ghost" aria-label="Deploy">
                      <Rocket />
                    </Button>
                    <Button size="icon-xs" variant="ghost" aria-label="Deploy">
                      <Rocket />
                    </Button>
                  </div>
                </div>
                <Separator />
                <div>
                  <p class="mb-3 text-sm font-medium text-muted-foreground">With icons</p>
                  <div class="flex gap-3">
                    <Button>
                      <Rocket :size="16" />
                      Deploy
                    </Button>
                    <Button variant="outline">
                      <Zap :size="16" />
                      Build
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent
            value="inputs"
            class="mt-6"
          >
            <Card>
              <CardContent class="space-y-6 pt-6">
                <div class="grid max-w-sm gap-4">
                  <div class="space-y-2">
                    <Label>Project name</Label>
                    <Input
                      placeholder="my-awesome-app"
                      model-value="lucity"
                    />
                  </div>
                  <div class="space-y-2">
                    <Label>Repository URL</Label>
                    <Input
                      placeholder="https://github.com/..."
                      model-value="https://github.com/zeitlos/lucity"
                    />
                  </div>
                  <div class="flex items-center gap-3">
                    <Switch />
                    <Label>Auto-deploy on push</Label>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent
            value="feedback"
            class="mt-6"
          >
            <Card>
              <CardContent class="space-y-6 pt-6">
                <div>
                  <p class="mb-3 text-sm font-medium text-muted-foreground">Badges</p>
                  <div class="flex flex-wrap gap-2">
                    <Badge>Default</Badge>
                    <Badge variant="secondary">
                      Secondary
                    </Badge>
                    <Badge variant="outline">
                      Outline
                    </Badge>
                    <Badge variant="destructive">
                      Destructive
                    </Badge>
                  </div>
                </div>
                <Separator />
                <div>
                  <p class="mb-3 text-sm font-medium text-muted-foreground">Progress</p>
                  <div class="max-w-sm space-y-3">
                    <Progress :model-value="72" />
                    <Progress :model-value="33" />
                  </div>
                </div>
                <Separator />
                <div>
                  <p class="mb-3 text-sm font-medium text-muted-foreground">Skeleton</p>
                  <div class="flex items-center gap-4">
                    <Skeleton class="h-12 w-12 rounded-full" />
                    <div class="space-y-2">
                      <Skeleton class="h-4 w-48" />
                      <Skeleton class="h-4 w-32" />
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent
            value="cards"
            class="mt-6"
          >
            <div class="grid grid-cols-3 gap-4">
              <Card>
                <CardHeader>
                  <div class="flex items-center gap-2">
                    <Check
                      :size="16"
                      class="text-chart-3"
                    />
                    <CardTitle class="text-base">
                      Healthy
                    </CardTitle>
                  </div>
                  <CardDescription>All systems operational</CardDescription>
                </CardHeader>
                <CardContent>
                  <p class="text-2xl font-bold text-foreground">
                    99.9%
                  </p>
                  <p class="text-xs text-muted-foreground">
                    Uptime last 30 days
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <div class="flex items-center gap-2">
                    <Rocket
                      :size="16"
                      class="text-primary"
                    />
                    <CardTitle class="text-base">
                      Deployments
                    </CardTitle>
                  </div>
                  <CardDescription>This week</CardDescription>
                </CardHeader>
                <CardContent>
                  <p class="text-2xl font-bold text-foreground">
                    42
                  </p>
                  <p class="text-xs text-muted-foreground">
                    Across 6 services
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <div class="flex items-center gap-2">
                    <AlertCircle
                      :size="16"
                      class="text-destructive"
                    />
                    <CardTitle class="text-base">
                      Issues
                    </CardTitle>
                  </div>
                  <CardDescription>Needs attention</CardDescription>
                </CardHeader>
                <CardContent>
                  <p class="text-2xl font-bold text-foreground">
                    2
                  </p>
                  <p class="text-xs text-muted-foreground">
                    Build failures
                  </p>
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </section>

      <!-- Social Preview -->
      <section class="mb-16">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Social Preview
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          Generate og:image for link sharing. 1280 &times; 640px.
        </p>

        <!-- Canvas preview -->
        <Card>
          <CardContent class="pt-6">
            <canvas
              ref="canvasRef"
              width="1280"
              height="640"
              class="w-full rounded-lg"
            />
          </CardContent>
        </Card>

        <!-- Controls -->
        <div class="mt-6 grid grid-cols-[1fr_auto] gap-8">
          <!-- Wallpaper selector -->
          <div>
            <p class="mb-3 text-sm font-medium text-muted-foreground">
              Background
            </p>
            <div class="grid grid-cols-7 gap-2">
              <button
                v-for="(wp, i) in wallpapers"
                :key="wp.name"
                class="group relative overflow-hidden rounded-lg border-2 transition-all"
                :class="selectedWallpaper === i ? 'border-primary' : 'border-transparent hover:border-border'"
                @click="selectedWallpaper = i"
              >
                <img
                  :src="wp.src"
                  :alt="wp.name"
                  class="aspect-video w-full object-cover"
                >
                <span
                  class="absolute inset-x-0 bottom-0 bg-black/50 px-1 py-0.5 text-[10px] text-white opacity-0 transition-opacity group-hover:opacity-100"
                >
                  {{ wp.name }}
                </span>
              </button>
            </div>
          </div>

          <!-- Text controls + download -->
          <div class="w-64 space-y-4">
            <div class="space-y-2">
              <Label>Headline</Label>
              <Input v-model="previewHeadline" />
            </div>
            <div class="space-y-2">
              <Label>Tagline</Label>
              <Input v-model="previewTagline" />
            </div>
            <Separator />
            <div class="flex gap-2">
              <Button
                variant="outline"
                class="flex-1 gap-1.5"
                @click="downloadPreview('png')"
              >
                <Download :size="14" />
                PNG
              </Button>
              <Button
                variant="outline"
                class="flex-1 gap-1.5"
                @click="downloadPreview('jpg')"
              >
                <Download :size="14" />
                JPG
              </Button>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.app-icon {
  --brand-mark: #f6ede0;
  background: #8843db;
}
</style>
