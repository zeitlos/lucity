<script setup lang="ts">
import { type HTMLAttributes, provide, ref, toRefs, watchEffect } from 'vue';
import useEmblaCarousel from 'embla-carousel-vue';
import type { EmblaCarouselType, EmblaPluginType } from 'embla-carousel';
import { cn } from '@/lib/utils';

type CarouselApi = EmblaCarouselType | undefined;
type CarouselOrientation = 'horizontal' | 'vertical';

interface CarouselProps {
  opts?: Parameters<typeof useEmblaCarousel>[0];
  plugins?: EmblaPluginType[];
  orientation?: CarouselOrientation;
  class?: HTMLAttributes['class'];
}

const props = withDefaults(defineProps<CarouselProps>(), {
  orientation: 'horizontal',
});

const emits = defineEmits<{
  (e: 'init-api', api: CarouselApi): void;
}>();

const carouselOptions = ref({ ...props.opts, axis: props.orientation === 'horizontal' ? 'x' : 'y' });
const [emblaNode, emblaApi] = useEmblaCarousel(carouselOptions as Parameters<typeof useEmblaCarousel>[0], props.plugins);
const canScrollPrev = ref(false);
const canScrollNext = ref(false);

function onSelect(api: EmblaCarouselType) {
  canScrollPrev.value = api.canScrollPrev();
  canScrollNext.value = api.canScrollNext();
}

function scrollPrev() {
  emblaApi.value?.scrollPrev();
}

function scrollNext() {
  emblaApi.value?.scrollNext();
}

watchEffect(() => {
  const api = emblaApi.value;
  if (!api) return;

  onSelect(api);
  api.on('reInit', onSelect);
  api.on('select', onSelect);
  emits('init-api', api);
});

provide('carousel', {
  ...toRefs(props),
  emblaApi,
  canScrollPrev,
  canScrollNext,
  scrollPrev,
  scrollNext,
});
</script>

<template>
  <div :class="cn('relative', props.class)" role="region" aria-roledescription="carousel">
    <slot :can-scroll-prev="canScrollPrev" :can-scroll-next="canScrollNext" />
  </div>
</template>
