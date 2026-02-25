<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue';
import BaseLogo from '@/components/BaseLogo.vue';
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
import { AlertCircle, Check, Download, Rocket, Zap } from 'lucide-vue-next';

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

type PreviewPoint = [number, number];

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

function isoProject(x: number, y: number): PreviewPoint {
  const CELL = 20;
  return [(x - y) * CELL * 0.866, (x + y) * CELL * 0.5];
}

function drawLogo(ctx: CanvasRenderingContext2D, cx: number, cy: number, logoHeight: number) {
  const L_CELLS: PreviewPoint[] = [[0, 2], [1, 2], [2, 2], [2, 1]];
  const TRI_VERTS: PreviewPoint[] = [[0, 1], [1, 0], [1, 1]];
  const EXT = 0.4;

  // Grid lines in raw isometric space
  const gridLines: { x1: number; y1: number; x2: number; y2: number }[] = [];
  for (let r = 0; r <= 3; r++) {
    const [gx1, gy1] = isoProject(-EXT, r);
    const [gx2, gy2] = isoProject(3 + EXT, r);
    gridLines.push({ x1: gx1, y1: gy1, x2: gx2, y2: gy2 });
  }
  for (let c = 0; c <= 3; c++) {
    const [gx1, gy1] = isoProject(c, -EXT);
    const [gx2, gy2] = isoProject(c, 3 + EXT);
    gridLines.push({ x1: gx1, y1: gy1, x2: gx2, y2: gy2 });
  }

  // Bounding box includes grid lines (matches BaseLogo viewBox)
  const allPts: PreviewPoint[] = [];
  for (const [c, r] of L_CELLS) {
    allPts.push(isoProject(c, r), isoProject(c + 1, r), isoProject(c + 1, r + 1), isoProject(c, r + 1));
  }
  for (const [x, y] of TRI_VERTS) allPts.push(isoProject(x, y));
  for (const line of gridLines) {
    allPts.push([line.x1, line.y1], [line.x2, line.y2]);
  }

  const xs = allPts.map((p) => p[0]);
  const ys = allPts.map((p) => p[1]);
  const minX = Math.min(...xs);
  const maxX = Math.max(...xs);
  const minY = Math.min(...ys);
  const maxY = Math.max(...ys);
  const rawW = maxX - minX;
  const rawH = maxY - minY;
  const scale = logoHeight / rawH;

  function project(x: number, y: number): PreviewPoint {
    const [px, py] = isoProject(x, y);
    return [cx + (px - minX - rawW / 2) * scale, cy + (py - minY - rawH / 2) * scale];
  }

  function projectRaw(px: number, py: number): PreviewPoint {
    return [cx + (px - minX - rawW / 2) * scale, cy + (py - minY - rawH / 2) * scale];
  }

  // Grid lines with fade (drawn first, behind tiles)
  ctx.lineWidth = Math.max(0.3, 0.35 * scale);
  for (const line of gridLines) {
    const [sx1, sy1] = projectRaw(line.x1, line.y1);
    const [sx2, sy2] = projectRaw(line.x2, line.y2);
    const grad = ctx.createLinearGradient(sx1, sy1, sx2, sy2);
    grad.addColorStop(0, 'rgba(255,255,255,0)');
    grad.addColorStop(0.15, 'rgba(255,255,255,0.25)');
    grad.addColorStop(0.85, 'rgba(255,255,255,0.25)');
    grad.addColorStop(1, 'rgba(255,255,255,0)');
    ctx.strokeStyle = grad;
    ctx.beginPath();
    ctx.moveTo(sx1, sy1);
    ctx.lineTo(sx2, sy2);
    ctx.stroke();
  }

  // L-shape tiles — use exact brand OKLCH values
  ctx.fillStyle = 'oklch(0.75 0.18 160)';
  ctx.strokeStyle = 'oklch(0.75 0.18 160)';
  ctx.lineWidth = Math.max(0.5, 0.5 * scale);
  ctx.lineJoin = 'round';

  for (const [c, r] of L_CELLS) {
    const [x1, y1] = project(c, r);
    const [x2, y2] = project(c + 1, r);
    const [x3, y3] = project(c + 1, r + 1);
    const [x4, y4] = project(c, r + 1);
    ctx.beginPath();
    ctx.moveTo(x1, y1);
    ctx.lineTo(x2, y2);
    ctx.lineTo(x3, y3);
    ctx.lineTo(x4, y4);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
  }

  // Triangle accent — warm beige from brand palette
  ctx.fillStyle = 'oklch(0.92 0.02 75)';
  ctx.beginPath();
  TRI_VERTS.forEach((vert, i) => {
    const [px, py] = project(vert[0], vert[1]);
    if (i === 0) ctx.moveTo(px, py);
    else ctx.lineTo(px, py);
  });
  ctx.closePath();
  ctx.fill();
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

  // Logo — smaller, upper portion (brand mark, not the star)
  drawLogo(ctx, W / 2, H / 2 - 135, 100);

  // Headline — large enough to read at card size (~400px wide display)
  ctx.fillStyle = 'white';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';
  ctx.font = '88px "Instrument Serif", Georgia, serif';
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
  await document.fonts.ready;
  await nextTick();
  drawCanvas();
});
</script>

<template>
  <div class="min-h-screen bg-background">
    <div class="mx-auto max-w-5xl px-8 py-12">
      <div class="mb-4 flex items-center gap-4">
        <BaseLogo :size="48" />
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
          Isometric 3x3 grid with an L-shape in primary and an accent triangle.
        </p>

        <div class="grid grid-cols-3 gap-8">
          <Card>
            <CardHeader>
              <CardTitle>Construction</CardTitle>
              <CardDescription>
                Grid coordinates and geometry
              </CardDescription>
            </CardHeader>
            <CardContent class="flex justify-center">
              <BaseLogo :size="220" debug />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Clean</CardTitle>
              <CardDescription>
                Colored tiles and accent triangle
              </CardDescription>
            </CardHeader>
            <CardContent class="flex justify-center">
              <BaseLogo :size="220" />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Mark</CardTitle>
              <CardDescription>
                Knockout circle for compact use
              </CardDescription>
            </CardHeader>
            <CardContent class="flex items-center justify-center gap-6">
              <BaseLogo :size="120" variant="mark" />
              <BaseLogo :size="120" variant="mark" class="text-primary" />
              <BaseLogo :size="120" variant="mark" debug />
            </CardContent>
          </Card>
        </div>

        <div class="mt-8 space-y-8">
          <div>
            <h3 class="mb-4 text-sm font-medium text-muted-foreground">
              Optical sizes — Default
            </h3>
            <p class="mb-6 text-xs text-muted-foreground">
              Small sizes show only the mark with an enlarged triangle. Grid lines and gradient fills appear at 96px+.
            </p>
            <div class="flex items-end gap-8">
              <div
                v-for="size in [24, 32, 48, 64, 96, 128]"
                :key="size"
                class="flex flex-col items-center gap-2"
              >
                <BaseLogo :size="size" />
                <span class="font-mono text-xs text-muted-foreground">{{ size }}</span>
                <span class="text-[10px] text-muted-foreground/60">
                  {{ size >= 96 ? 'detail' : size <= 32 ? 'bold' : 'standard' }}
                </span>
              </div>
            </div>
          </div>

          <div>
            <h3 class="mb-4 text-sm font-medium text-muted-foreground">
              Optical sizes — Mark
            </h3>
            <div class="flex items-end gap-8">
              <div
                v-for="size in [16, 24, 32, 48, 64]"
                :key="'mark-' + size"
                class="flex flex-col items-center gap-2"
              >
                <BaseLogo :size="size" variant="mark" />
                <span class="font-mono text-xs text-muted-foreground">{{ size }}</span>
              </div>
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
            <div class="grid grid-cols-5 gap-4">
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
          Sora for headings and body. Instrument Serif for display. Fira Code for monospace.
        </p>

        <Card>
          <CardContent class="space-y-6 pt-6">
            <div class="space-y-3">
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
                    <Button size="sm">
                      Small
                    </Button>
                    <Button>Default</Button>
                    <Button size="lg">
                      Large
                    </Button>
                    <Button size="icon">
                      <Rocket :size="18" />
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
