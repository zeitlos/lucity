<script setup lang="ts">
import { computed, ref } from 'vue';
import { Download } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { type BrandAsset, brandSvgUrl, downloadPng, downloadSvg } from '@/lib/brand';
import { cn } from '@/lib/utils';

const props = defineProps<{
  asset: BrandAsset;
  name: string;
  note?: string;
  background: string;
  pngWidth: number;
  imageClass?: string;
}>();

const src = computed(() => brandSvgUrl(props.asset));
const rendering = ref(false);

async function png() {
  rendering.value = true;
  try {
    await downloadPng(props.asset, props.pngWidth);
  } finally {
    rendering.value = false;
  }
}
</script>

<template>
  <div class="flex flex-col bg-card">
    <div
      class="flex h-40 items-center justify-center p-8"
      :style="{ background }"
    >
      <img
        :src="src"
        :alt="name"
        :class="cn('max-h-full max-w-full', imageClass)"
      >
    </div>
    <div class="flex flex-wrap items-center gap-x-2 gap-y-3 px-4 py-3">
      <div class="min-w-0 flex-1 basis-28">
        <p class="truncate text-sm font-medium text-foreground">
          {{ name }}
        </p>
        <p
          v-if="note"
          class="truncate text-xs text-muted-foreground"
        >
          {{ note }}
        </p>
      </div>
      <div class="flex gap-2">
        <Button
          variant="outline"
          size="xs"
          @click="downloadSvg(asset)"
        >
          <Download />
          SVG
        </Button>
        <Button
          variant="outline"
          size="xs"
          :pending="rendering"
          @click="png"
        >
          <Download />
          PNG
        </Button>
      </div>
    </div>
  </div>
</template>
