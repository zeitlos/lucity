<script setup lang="ts">
const props = defineProps<{
  slug?: string;
}>();

const router = useRouter();

const shown = ref(props.slug);
watch(() => props.slug, (slug) => {
  if (slug) {
    shown.value = slug;
  }
});

const index = computed(() => agentSheets.findIndex(sheet => sheet.slug === shown.value));
const sheet = computed(() => agentSheets[index.value]);
const previous = computed(() => agentSheets[index.value - 1]);
const next = computed(() => agentSheets[index.value + 1]);

const body = ref<HTMLElement>();
watch(shown, () => {
  body.value?.scrollTo({ top: 0 });
});

function focusBody(event: Event) {
  event.preventDefault();
  body.value?.focus({ preventScroll: true });
}

const panel = ref<HTMLElement>();
const settle = '0.3s cubic-bezier(0.32, 0.72, 0, 1)';

interface Drag {
  startY: number;
  lastY: number;
  lastTime: number;
  velocity: number;
  offset: number;
  canDrag: boolean;
  mode: 'pending' | 'drag' | 'scroll';
}

let drag: Drag | undefined;
let dismissed = false;

function layers() {
  const content = panel.value?.closest<HTMLElement>('[data-slot="content"]') ?? undefined;
  const overlay = content?.parentElement?.querySelector<HTMLElement>('[data-slot="overlay"]') ?? undefined;
  return { content, overlay };
}

function resetLayers() {
  const { content, overlay } = layers();
  for (const element of [content, overlay]) {
    element?.style.removeProperty('transform');
    element?.style.removeProperty('transition');
    element?.style.removeProperty('animation');
    element?.style.removeProperty('opacity');
  }
}

function onTouchStart(event: TouchEvent) {
  const touch = event.touches[0];
  if (!touch || event.touches.length > 1 || !window.matchMedia('(max-width: 639px)').matches) {
    drag = undefined;
    return;
  }
  const inBody = !!body.value?.contains(event.target as Node);
  drag = {
    startY: touch.clientY,
    lastY: touch.clientY,
    lastTime: event.timeStamp,
    velocity: 0,
    offset: 0,
    canDrag: !inBody || (body.value?.scrollTop ?? 0) <= 0,
    mode: 'pending',
  };
}

function onTouchMove(event: TouchEvent) {
  const touch = event.touches[0];
  if (!drag || !touch || drag.mode === 'scroll') {
    return;
  }
  const delta = touch.clientY - drag.startY;
  if (drag.mode === 'pending') {
    if (delta === 0) {
      return;
    }
    drag.mode = delta > 0 && drag.canDrag ? 'drag' : 'scroll';
    if (drag.mode === 'scroll') {
      return;
    }
  }
  event.preventDefault();
  const elapsed = event.timeStamp - drag.lastTime;
  if (elapsed > 0) {
    drag.velocity = (touch.clientY - drag.lastY) / elapsed;
  }
  drag.lastY = touch.clientY;
  drag.lastTime = event.timeStamp;
  drag.offset = Math.max(0, delta);
  const { content, overlay } = layers();
  if (!content) {
    return;
  }
  content.style.transition = 'none';
  content.style.transform = `translateY(${drag.offset}px)`;
  if (overlay) {
    overlay.style.transition = 'none';
    overlay.style.opacity = String(Math.max(0, 1 - drag.offset / content.offsetHeight));
  }
}

function onTouchEnd(event: TouchEvent) {
  const ended = drag;
  drag = undefined;
  if (ended?.mode !== 'drag') {
    return;
  }
  const { content, overlay } = layers();
  if (!content) {
    return;
  }
  content.style.transition = `transform ${settle}`;
  if (overlay) {
    overlay.style.transition = `opacity ${settle}`;
  }
  const flicked = ended.velocity > 0.5 && ended.offset > 24 && event.timeStamp - ended.lastTime < 100;
  if (flicked || ended.offset > content.offsetHeight * 0.25) {
    dismissed = true;
    content.style.animation = 'none';
    content.style.transform = 'translateY(100%)';
    if (overlay) {
      overlay.style.animation = 'none';
      overlay.style.opacity = '0';
    }
    setTimeout(close, 300);
    return;
  }
  content.style.transform = 'translateY(0)';
  if (overlay) {
    overlay.style.opacity = '1';
  }
  setTimeout(resetLayers, 300);
}

watch(() => props.slug, (slug) => {
  if (!slug && dismissed) {
    dismissed = false;
    requestAnimationFrame(() => requestAnimationFrame(resetLayers));
  }
});

function close() {
  const back = router.options.history.state.back;
  if (typeof back === 'string' && back.replace(/[?#].*$/, '') === '/agents') {
    router.back();
    return;
  }
  router.push('/agents');
}

const copied = ref(false);
let copiedTimer: ReturnType<typeof setTimeout> | undefined;

async function copyLink() {
  await navigator.clipboard.writeText(`${window.location.origin}/agents/${shown.value}`);
  copied.value = true;
  clearTimeout(copiedTimer);
  copiedTimer = setTimeout(() => {
    copied.value = false;
  }, 1600);
}

const iconButton = 'flex size-11 cursor-pointer items-center justify-center rounded-[10px] text-stone-600 transition-colors hover:bg-stone-900/5 hover:text-stone-900 focus-visible:outline-2 focus-visible:outline-(--brand-mark) dark:text-stone-400 dark:hover:bg-white/8 dark:hover:text-stone-100';
</script>

<template>
  <USlideover
    :open="!!slug"
    :portal="false"
    :unmount-on-hide="false"
    :title="sheet?.title"
    :description="sheet?.intro"
    :content="{ onOpenAutoFocus: focusBody }"
    :ui="{
      overlay: 'z-60 bg-[oklch(0.20_0.03_55/.44)] dark:bg-black/60',
      content: 'z-60 inset-y-3 right-3 w-[calc(100%-1.5rem)] max-w-[640px] divide-y-0 overflow-hidden rounded-2xl bg-[oklch(0.99_0.006_80)] sm:shadow-[0_0_0_1px_oklch(0.20_0.03_55/.2),0_40px_80px_-30px_oklch(0.15_0.03_55/.7)] sm:ring-0 max-sm:inset-x-0 max-sm:top-10 max-sm:bottom-0 max-sm:w-full max-sm:max-w-none max-sm:rounded-t-[20px] max-sm:rounded-b-none max-sm:data-[state=open]:animate-[slide-in-from-bottom_350ms_cubic-bezier(0.32,0.72,0,1)] max-sm:data-[state=closed]:animate-[slide-out-to-bottom_250ms_var(--ease-out)] dark:bg-stone-900 dark:sm:shadow-[0_0_0_1px_oklch(1_0_0/.1)]',
    }"
    @update:open="(value) => !value && close()"
  >
    <template #content>
      <div
        v-if="sheet"
        ref="panel"
        class="flex h-full flex-col font-sans text-stone-900 dark:text-stone-100"
        @touchstart="onTouchStart"
        @touchmove="onTouchMove"
        @touchend="onTouchEnd"
        @touchcancel="onTouchEnd"
      >
        <div class="flex flex-none justify-center pt-2 max-sm:touch-none sm:hidden" aria-hidden="true">
          <span class="h-1.25 w-10 rounded-full bg-stone-300 dark:bg-stone-600" />
        </div>
        <header class="flex h-16 flex-none items-center justify-between gap-3 border-b border-stone-200 pr-3 pl-5 max-sm:h-13 max-sm:touch-none sm:pl-10 dark:border-stone-800">
          <span class="font-display text-xl italic text-stone-700 dark:text-stone-300">Agent hosting</span>
          <div class="flex gap-0.5">
            <button type="button" :class="iconButton" :aria-label="copied ? 'Link copied' : 'Copy link to this sheet'" @click="copyLink">
              <UIcon :name="copied ? 'i-lucide-check' : 'i-lucide-link'" class="size-[19px]" />
            </button>
            <button type="button" :class="iconButton" aria-label="Close" @click="close">
              <UIcon name="i-lucide-x" class="size-[21px]" />
            </button>
          </div>
        </header>

        <div ref="body" tabindex="-1" class="min-h-0 flex-1 overflow-y-auto overscroll-contain outline-none">
          <article class="px-5 pt-10 sm:px-10 sm:pt-12">
            <h1 class="font-display text-[44px]/[1.05] sm:text-[56px]/[1.05]">{{ sheet.title }}</h1>
            <p class="mt-5 text-lg/relaxed text-stone-700 sm:text-[19px]/[1.6] dark:text-stone-300">{{ sheet.intro }}</p>

            <figure class="mt-9">
              <div class="figure-frame flex justify-center rounded-xl px-4 py-10 sm:px-8">
                <AgentsSheetFigure :slug="sheet.slug" />
              </div>
              <figcaption class="mt-3 font-mono text-[13px] text-stone-500 dark:text-stone-400">{{ sheet.caption }}</figcaption>
            </figure>

            <section v-for="section in sheet.sections" :key="section.title" class="mt-11">
              <h2 class="font-display text-[28px]/[1.2]">{{ section.title }}</h2>
              <p class="mt-2.5 text-[17px]/[1.65] text-stone-700 dark:text-stone-300">{{ section.text }}</p>
            </section>

            <section class="mt-14">
              <h2 class="font-display text-[28px]/[1.2]">{{ sheet.table.title }}</h2>
              <div role="table" :aria-label="sheet.table.title" class="mt-4.5 border-t border-stone-300 dark:border-stone-700">
                <div role="row" class="flex flex-wrap gap-x-6 gap-y-1.5 border-b border-stone-200 py-3 text-sm text-stone-500 max-sm:hidden dark:border-stone-800 dark:text-stone-400">
                  <span role="columnheader" class="max-w-[210px] flex-[1_1_150px]">{{ sheet.table.head[0] }}</span>
                  <span role="columnheader" class="flex-[999_1_250px]">{{ sheet.table.head[1] }}</span>
                </div>
                <div v-for="row in sheet.table.rows" :key="row.name" role="row" class="flex flex-wrap gap-x-6 gap-y-1.5 border-b border-stone-200 py-3.75 dark:border-stone-800">
                  <div role="cell" class="flex max-w-[210px] flex-[1_1_150px] items-start gap-2.75 font-semibold">
                    <span class="mt-0.5" :class="{ 'text-(--brand-mark)': row.icon === 'lucity' }"><AgentsIcon :name="row.icon" :size="20" /></span>
                    <span>
                      {{ row.name }}
                      <span v-if="row.region" class="mt-0.5 block text-sm font-normal text-stone-500 dark:text-stone-400">{{ row.region }}</span>
                    </span>
                  </div>
                  <div role="cell" class="flex-[999_1_250px] text-base/[1.55] text-stone-700 dark:text-stone-300">{{ row.text }}</div>
                </div>
              </div>
            </section>
          </article>

          <footer class="px-5 pt-16 pb-10 sm:px-10">
            <nav aria-label="More spec sheets" class="flex justify-between gap-6 border-t border-stone-200 pt-5 dark:border-stone-800">
              <NuxtLink v-if="previous" :to="`/agents/${previous.slug}`" replace class="flex min-h-11 flex-col items-start gap-0.75">
                <span class="text-sm text-stone-500 dark:text-stone-400">Previous</span>
                <span class="flex items-center gap-2 text-[17px] font-semibold"><UIcon name="i-lucide-arrow-left" class="size-4.5" />{{ previous.title }}</span>
              </NuxtLink>
              <span v-else />
              <NuxtLink v-if="next" :to="`/agents/${next.slug}`" replace class="flex min-h-11 flex-col items-end gap-0.75">
                <span class="text-sm text-stone-500 dark:text-stone-400">Next</span>
                <span class="flex items-center gap-2 text-[17px] font-semibold">{{ next.title }}<UIcon name="i-lucide-arrow-right" class="size-4.5" /></span>
              </NuxtLink>
            </nav>
          </footer>
        </div>
      </div>
    </template>
  </USlideover>
</template>

<style scoped>
.figure-frame {
  background-color: oklch(0.97 0.01 80);
  background-image:
    linear-gradient(oklch(0.925 0.014 78) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.925 0.014 78) 1px, transparent 1px);
  background-size: 32px 32px;
  background-position: -1px -1px;
  box-shadow: inset 0 0 0 1px oklch(0.88 0.02 78);
}

.dark .figure-frame {
  background-color: oklch(0.20 0.015 55);
  background-image:
    linear-gradient(oklch(1 0 0 / .06) 1px, transparent 1px),
    linear-gradient(90deg, oklch(1 0 0 / .06) 1px, transparent 1px);
  box-shadow: inset 0 0 0 1px oklch(1 0 0 / .08);
}
</style>
