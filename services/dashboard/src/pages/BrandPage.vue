<script setup lang="ts">
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
import { AlertCircle, Check, Rocket, Zap } from 'lucide-vue-next';

const colors = [
  { name: 'Primary', var: '--primary', fg: '--primary-foreground', desc: 'Brand purple' },
  { name: 'Secondary', var: '--secondary', fg: '--secondary-foreground', desc: 'Light purple' },
  { name: 'Accent', var: '--accent', fg: '--accent-foreground', desc: 'Coral highlight' },
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

        <div class="grid grid-cols-2 gap-8">
          <Card>
            <CardHeader>
              <CardTitle>Construction</CardTitle>
              <CardDescription>
                Grid coordinates and geometry
              </CardDescription>
            </CardHeader>
            <CardContent class="flex justify-center">
              <BaseLogo :size="280" debug />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Clean</CardTitle>
              <CardDescription>
                Final mark without construction lines
              </CardDescription>
            </CardHeader>
            <CardContent class="flex justify-center">
              <BaseLogo :size="280" />
            </CardContent>
          </Card>
        </div>

        <div class="mt-8">
          <h3 class="mb-4 text-sm font-medium text-muted-foreground">
            Optical sizes
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
          Oxanium for headings and body. Fira Code for monospace.
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
    </div>
  </div>
</template>
