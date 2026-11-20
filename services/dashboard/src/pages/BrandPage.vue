<script setup lang="ts">
import { computed, ref } from 'vue';
import { Download } from '@lucide/vue';
import BaseLogo from '@/components/BaseLogo.vue';
import CopyButton from '@/components/CopyButton.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import BrandAssetTile from '@/components/brand/BrandAssetTile.vue';
import BrandSocialPreview from '@/components/brand/BrandSocialPreview.vue';
import { Button } from '@/components/ui/button';
import { failureToast } from '@/components/ui/sonner';
import {
  CREAM,
  ESPRESSO,
  GREEN,
  INK,
  LILAC,
  LOCKUP_HEIGHT,
  LOCKUP_WIDTH,
  LOCKUP_WORDMARK_TRANSFORM,
  PAPER,
  PURPLE,
  appIconAsset,
  downloadKit,
  flatIconAsset,
  iconThemes,
  lockupAsset,
  markAsset,
  wordmarkAsset,
} from '@/lib/brand';
import { LOGO_HEIGHT, LOGO_PATH, WORDMARK_CAP, WORDMARK_PATH } from '@/lib/logo';
import { WORDMARK_X_HEIGHT, wordmarkGlyphs } from '@/lib/wordmarkGlyphs';
import { usePageTitle } from '@/composables/usePageTitle';

usePageTitle(['Brand']);

const lockups = [
  { name: 'Light', note: 'Purple and ink', asset: lockupAsset('lucity-logo-light', PURPLE, INK), background: CREAM },
  { name: 'Dark', note: 'Lilac and paper', asset: lockupAsset('lucity-logo-dark', LILAC, PAPER), background: ESPRESSO },
  { name: 'Ink', note: 'One color', asset: lockupAsset('lucity-logo-ink', INK, INK), background: CREAM },
  { name: 'Cream', note: 'One color, reversed', asset: lockupAsset('lucity-logo-cream', CREAM, CREAM), background: PURPLE },
];

const wordmarks = [
  { name: 'Ink', note: INK, asset: wordmarkAsset('lucity-wordmark-ink', INK), background: CREAM },
  { name: 'Paper', note: PAPER, asset: wordmarkAsset('lucity-wordmark-paper', PAPER), background: ESPRESSO },
];

type IconStyle = 'macos' | 'flat';

const iconStyle = ref<IconStyle>('macos');
const iconStyles: { value: IconStyle; label: string }[] = [
  { value: 'macos', label: 'macOS' },
  { value: 'flat', label: 'Flat' },
];

const icons = computed(() => iconThemes.map((theme) => {
  if (iconStyle.value === 'macos') {
    return {
      name: theme.name,
      note: theme.file === 'espresso' ? 'The cockpit app' : undefined,
      asset: appIconAsset(theme),
    };
  }
  return {
    name: theme.name,
    note: theme.file === 'purple' ? 'The web app icon' : undefined,
    asset: flatIconAsset(theme),
  };
}));

const marks = [
  { name: 'Purple', color: PURPLE, background: CREAM },
  { name: 'Lilac', color: LILAC, background: ESPRESSO },
  { name: 'Ink', color: INK, background: CREAM },
  { name: 'Cream', color: CREAM, background: PURPLE },
  { name: 'White', color: '#ffffff', background: INK },
  { name: 'Black', color: '#000000', background: '#ffffff' },
].map(mark => ({ ...mark, asset: markAsset(`lucity-mark-${mark.name.toLowerCase()}`, mark.color) }));

const kit = [
  ...[...lockups, ...wordmarks].map(item => ({ asset: item.asset, pngWidth: 2048 })),
  ...iconThemes.flatMap(theme => [appIconAsset(theme), flatIconAsset(theme)]).map(asset => ({ asset, pngWidth: 1024 })),
  ...marks.map(mark => ({ asset: mark.asset, pngWidth: 1024 })),
];

const preparing = ref(false);

async function downloadAll() {
  preparing.value = true;
  try {
    await downloadKit(kit);
  } catch (error) {
    failureToast('Failed to download the brand kit', error);
  } finally {
    preparing.value = false;
  }
}

const specimenPad = 20;
const specimenBox = `${-specimenPad} ${-specimenPad} ${LOCKUP_WIDTH + specimenPad * 2} ${LOCKUP_HEIGHT + specimenPad * 2}`;
const lockupUnit = LOGO_HEIGHT / WORDMARK_CAP;
const guides = [
  { label: 'Cap line', y: 0 },
  { label: 'x-height', y: LOGO_HEIGHT - WORDMARK_X_HEIGHT * lockupUnit },
  { label: 'Baseline', y: LOGO_HEIGHT },
  { label: 'Step', y: LOCKUP_HEIGHT },
].map(guide => ({ ...guide, top: ((guide.y + specimenPad) / (LOCKUP_HEIGHT + specimenPad * 2)) * 100 }));

const settings = [
  { label: 'Typeface', value: 'Mona Sans, Display' },
  { label: 'Width', value: '105' },
  { label: 'Weight', value: '600' },
  { label: 'Optical size', value: '100' },
  { label: 'Tracking', value: '−8' },
  { label: 'Kerning', value: 't y −6, c i +4' },
  { label: 'Mark', value: 'Cap height, on the baseline' },
  { label: 'Gap', value: '0.61 × mark height' },
];

const glyphNotes: Record<string, { title: string; note: string }> = {
  L: { title: 'Shorter L', note: 'The arm is 48 units shorter, so the u moves in closer.' },
  i: { title: 'Square dot', note: 'A square 0.934 times the stem, its top on the cap line.' },
  t: { title: 'Flat t', note: 'The slanted top is cut flat and raised to the cap line.' },
  y: { title: 'Stepped y', note: 'Built from the u. The tail steps 1.08 stems left like the mark, on a foot 1.41 stems wide, 200 units below the baseline.' },
};

const glyphTop = -WORDMARK_CAP - 80;
const glyphHeight = WORDMARK_CAP + 200 + 160;
const glyphGuides = [-WORDMARK_CAP, -WORDMARK_X_HEIGHT, 0, 200];

const glyphs = wordmarkGlyphs.map(glyph => ({
  ...glyph,
  ...glyphNotes[glyph.letter]!,
  viewBox: `${glyph.left - 80} ${glyphTop} ${glyph.right - glyph.left + 160} ${glyphHeight}`,
  from: glyph.left - 80,
  to: glyph.right + 80,
}));

const typefaces = [
  { name: 'Mona Sans', role: 'Wordmark, docs and landing page', family: '\'Mona Sans\', system-ui, sans-serif', sample: 'Deploy anything. Own your stack.' },
  { name: 'Redaction 20', role: 'Headlines', family: 'var(--font-serif)', sample: 'The PaaS you can leave.' },
  { name: 'Sora', role: 'Dashboard', family: 'var(--font-sans)', sample: 'Projects, environments and services' },
  { name: 'Fira Code', role: 'Code', family: 'var(--font-mono)', sample: 'git push origin main' },
];

const colors = [
  { name: 'Purple', value: PURPLE, use: 'The mark on light backgrounds' },
  { name: 'Lilac', value: LILAC, use: 'The mark on dark backgrounds' },
  { name: 'Ink', value: INK, use: 'Wordmark and text on light' },
  { name: 'Paper', value: PAPER, use: 'Wordmark and text on dark' },
  { name: 'Cream', value: CREAM, use: 'Light background' },
  { name: 'Espresso', value: ESPRESSO, use: 'Dark background' },
  { name: 'Green', value: GREEN, use: 'Actions' },
];
</script>

<template>
  <div class="min-h-screen bg-background">
    <div class="mx-auto max-w-5xl px-4 py-12 sm:px-8">
      <header class="mb-16 flex flex-wrap items-center gap-4">
        <div class="flex items-center gap-4">
          <BaseLogo :size="36" />
          <div>
            <h1 class="font-serif text-4xl text-foreground">
              Brand
            </h1>
            <p class="text-muted-foreground">
              Logos, icons, type and colors for Lucity
            </p>
          </div>
        </div>
        <Button
          variant="outline"
          class="ml-auto"
          :pending="preparing"
          @click="downloadAll"
        >
          <Download />
          {{ preparing ? 'Preparing…' : 'Download all' }}
        </Button>
      </header>

      <section class="mb-20">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Logo
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          A folded sheet with two bends next to a customized Mona Sans wordmark. Purple on light, lilac on dark.
        </p>

        <h3 class="mb-3 text-sm font-medium text-foreground">
          Lockup
        </h3>
        <div class="mb-10 grid gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-2">
          <BrandAssetTile
            v-for="lockup in lockups"
            :key="lockup.asset.file"
            :asset="lockup.asset"
            :name="lockup.name"
            :note="lockup.note"
            :background="lockup.background"
            :png-width="2048"
            image-class="w-56"
          />
        </div>

        <h3 class="mb-3 text-sm font-medium text-foreground">
          Wordmark
        </h3>
        <div class="grid gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-2">
          <BrandAssetTile
            v-for="wordmark in wordmarks"
            :key="wordmark.asset.file"
            :asset="wordmark.asset"
            :name="wordmark.name"
            :note="wordmark.note"
            :background="wordmark.background"
            :png-width="2048"
            image-class="w-40"
          />
        </div>
      </section>

      <section class="mb-20">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Icons
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          The mark on a tile for app icons and avatars, and on its own for everything else.
        </p>

        <div class="mb-3 flex flex-wrap items-end justify-between gap-x-4 gap-y-3">
          <div>
            <h3 class="text-sm font-medium text-foreground">
              App icon
            </h3>
            <p class="text-sm text-muted-foreground">
              {{ iconStyle === 'macos' ? 'Rounded tile with a lit rim and a shaded bend, on the macOS grid at 1024 px.' : 'Square tile, the mark at 72% of its width.' }}
            </p>
          </div>
          <SegmentedControl
            v-model="iconStyle"
            label="Icon style"
            :options="iconStyles"
          />
        </div>
        <div class="mb-10 grid grid-cols-1 gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-2 lg:grid-cols-4">
          <BrandAssetTile
            v-for="icon in icons"
            :key="icon.asset.file"
            :asset="icon.asset"
            :name="icon.name"
            :note="icon.note"
            background="var(--muted)"
            :png-width="1024"
            :image-class="iconStyle === 'macos' ? 'size-28' : 'size-20'"
          />
        </div>

        <h3 class="text-sm font-medium text-foreground">
          Mark
        </h3>
        <p class="mb-3 text-sm text-muted-foreground">
          Transparent background, cropped to the mark.
        </p>
        <div class="grid gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-2 lg:grid-cols-3">
          <BrandAssetTile
            v-for="mark in marks"
            :key="mark.asset.file"
            :asset="mark.asset"
            :name="mark.name"
            :note="mark.color"
            :background="mark.background"
            :png-width="1024"
            image-class="w-20"
          />
        </div>
      </section>

      <section class="mb-20">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Typeface
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          The wordmark is drawn from Mona Sans Display and changes four letters, so the word shares the mark's flat top and its bend. Units are thousandths of an em.
        </p>

        <div class="mb-10 overflow-hidden rounded-lg border border-border bg-card">
          <div class="flex gap-4 p-6 sm:p-10">
            <div class="relative hidden w-16 shrink-0 text-right text-xs text-muted-foreground sm:block">
              <span
                v-for="guide in guides"
                :key="guide.label"
                class="absolute right-0 -translate-y-1/2 whitespace-nowrap"
                :style="{ top: `${guide.top}%` }"
              >
                {{ guide.label }}
              </span>
            </div>
            <svg
              :viewBox="specimenBox"
              class="min-w-0 flex-1"
              role="img"
              aria-label="Lucity lockup with type guides"
            >
              <line
                v-for="guide in guides"
                :key="guide.label"
                :x1="-specimenPad"
                :x2="LOCKUP_WIDTH + specimenPad"
                :y1="guide.y"
                :y2="guide.y"
                class="stroke-muted-foreground/40"
                vector-effect="non-scaling-stroke"
              />
              <path :d="LOGO_PATH" class="fill-(--brand-mark)" />
              <path
                :transform="LOCKUP_WORDMARK_TRANSFORM"
                :d="WORDMARK_PATH"
                class="fill-foreground"
              />
            </svg>
          </div>
          <dl class="grid grid-cols-2 gap-px border-t border-border bg-border sm:grid-cols-4">
            <div
              v-for="setting in settings"
              :key="setting.label"
              class="bg-card px-4 py-3"
            >
              <dt class="text-xs text-muted-foreground">
                {{ setting.label }}
              </dt>
              <dd class="text-sm text-foreground">
                {{ setting.value }}
              </dd>
            </div>
          </dl>
        </div>

        <h3 class="text-sm font-medium text-foreground">
          Letters
        </h3>
        <p class="mb-3 text-sm text-muted-foreground">
          Each edit over the stock letter, dashed.
        </p>
        <div class="mb-10 grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-border bg-border lg:grid-cols-4">
          <div
            v-for="glyph in glyphs"
            :key="glyph.letter"
            class="bg-card"
          >
            <svg
              :viewBox="glyph.viewBox"
              class="h-56 w-full px-4 pt-4"
              role="img"
              :aria-label="glyph.title"
            >
              <line
                v-for="y in glyphGuides"
                :key="y"
                :x1="glyph.from"
                :x2="glyph.to"
                :y1="y"
                :y2="y"
                class="stroke-muted-foreground/30"
                vector-effect="non-scaling-stroke"
              />
              <path :d="glyph.custom" class="fill-foreground" />
              <path
                :d="glyph.stock"
                class="fill-none stroke-(--brand-mark)"
                stroke-width="1.5"
                stroke-dasharray="4 3"
                vector-effect="non-scaling-stroke"
              />
            </svg>
            <div class="px-4 pt-2 pb-4">
              <p class="text-sm font-medium text-foreground">
                {{ glyph.title }}
              </p>
              <p class="text-sm text-muted-foreground">
                {{ glyph.note }}
              </p>
            </div>
          </div>
        </div>

        <h3 class="mb-3 text-sm font-medium text-foreground">
          Typefaces
        </h3>
        <div class="divide-y divide-border overflow-hidden rounded-lg border border-border bg-card">
          <div
            v-for="typeface in typefaces"
            :key="typeface.name"
            class="flex flex-col gap-2 px-4 py-4 sm:flex-row sm:items-center sm:gap-6"
          >
            <div class="sm:w-56 sm:shrink-0">
              <p class="text-sm font-medium text-foreground">
                {{ typeface.name }}
              </p>
              <p class="text-sm text-muted-foreground">
                {{ typeface.role }}
              </p>
            </div>
            <p
              class="min-w-0 text-2xl text-foreground"
              :style="{ fontFamily: typeface.family }"
            >
              {{ typeface.sample }}
            </p>
          </div>
        </div>
      </section>

      <section class="mb-20">
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Colors
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          The logo colors, the two backgrounds they sit on, and the green for actions.
        </p>

        <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div
            v-for="color in colors"
            :key="color.name"
            class="overflow-hidden rounded-lg border border-border bg-card"
          >
            <div
              class="h-20 border-b border-border"
              :style="{ background: color.value }"
            />
            <div class="flex items-start gap-2 px-3 py-2.5">
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-foreground">
                  {{ color.name }}
                </p>
                <p class="font-mono text-xs text-muted-foreground">
                  {{ color.value }}
                </p>
                <p class="mt-1 text-xs text-muted-foreground">
                  {{ color.use }}
                </p>
              </div>
              <CopyButton
                :value="color.value"
                :label="`Copy ${color.name.toLowerCase()}`"
              />
            </div>
          </div>
        </div>
      </section>

      <section>
        <h2 class="mb-2 font-serif text-2xl text-foreground">
          Social image
        </h2>
        <p class="mb-8 text-sm text-muted-foreground">
          Cards for releases, link previews and social posts.
        </p>
        <BrandSocialPreview />
      </section>
    </div>
  </div>
</template>
