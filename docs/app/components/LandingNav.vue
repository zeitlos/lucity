<script setup lang="ts">
type MenuKey = 'features' | 'solutions' | 'resources';
type Section = MenuKey | 'pricing';

const appUrl = '/app';
const signInUrl = `${appUrl}/login`;
const signUpUrl = `${appUrl}/login`;
const githubUrl = 'https://github.com/zeitlos/lucity';

const logoMenu = [[{
  label: 'Download logos',
  icon: 'i-lucide-download',
  to: `${appUrl}/brand`,
  external: true,
}]];

const menus: { key: MenuKey; label: string }[] = [
  { key: 'features', label: 'Features' },
  { key: 'solutions', label: 'Solutions' },
  { key: 'resources', label: 'Resources' },
];

const sections: Record<Section, string[]> = {
  features: ['/features', '/security', '/no-lock-in'],
  solutions: ['/use-cases', '/comparisons', '/agents'],
  resources: ['/blog', '/docs'],
  pricing: ['/pricing'],
};

const route = useRoute();
const overDark = useHeaderOverDark();
const { open: searchOpen } = useContentSearch();

const current = computed(() => {
  const match = Object.entries(sections).find(([, bases]) =>
    bases.some(base => route.path === base || route.path.startsWith(`${base}/`)),
  );
  return match ? match[0] as Section : null;
});

const header = ref<HTMLElement>();
const burger = ref<HTMLElement>();

const tucked = ref(false);
const open = ref<MenuKey | null>(null);
const pinned = ref(false);
const sheetOpen = ref(false);
const sheetTop = ref(0);

const floating = computed(() => tucked.value && !sheetOpen.value);

let closeTimer: ReturnType<typeof setTimeout> | undefined;
let quietFocus = false;
let desktop: MediaQueryList | undefined;

function show(key: MenuKey, pin: boolean) {
  clearTimeout(closeTimer);
  open.value = key;
  pinned.value = pin;
}

function hide() {
  clearTimeout(closeTimer);
  open.value = null;
  pinned.value = false;
}

function hideSoon() {
  clearTimeout(closeTimer);
  closeTimer = setTimeout(hide, 140);
}

function onTriggerEnter(event: PointerEvent, key: MenuKey) {
  if (event.pointerType !== 'mouse') return;
  if (open.value === key) clearTimeout(closeTimer);
  else show(key, false);
}

function onTriggerClick(key: MenuKey) {
  if (open.value === key && pinned.value) hide();
  else show(key, true);
}

function onTriggerFocus(event: FocusEvent, key: MenuKey) {
  if (quietFocus || open.value === key) return;
  if ((event.target as HTMLElement).matches(':focus-visible')) show(key, false);
}

function onTriggerDown(event: KeyboardEvent, key: MenuKey) {
  event.preventDefault();
  show(key, true);
  nextTick(() => panel(key)?.querySelector<HTMLElement>('a, button')?.focus());
}

function onPlainEnter(event: PointerEvent) {
  if (event.pointerType === 'mouse' && !pinned.value) hide();
}

function onHeaderEnter() {
  clearTimeout(closeTimer);
}

function onHeaderLeave(event: PointerEvent) {
  if (event.pointerType === 'mouse' && !pinned.value) hideSoon();
}

function onFocusOut(event: FocusEvent) {
  const next = event.relatedTarget as Node | null;
  if (!next || !open.value) return;
  if (!trigger(open.value)?.contains(next) && !panel(open.value)?.contains(next)) hide();
}

function onLinkClick(event: MouseEvent) {
  if (!(event.target as HTMLElement).closest('a')) return;
  hide();
  sheetOpen.value = false;
}

function openSearch() {
  hide();
  searchOpen.value = true;
}

function panel(key: MenuKey) {
  return header.value?.querySelector<HTMLElement>(`#menu-${key}`);
}

function trigger(key: MenuKey) {
  return header.value?.querySelector<HTMLElement>(`[aria-controls="menu-${key}"]`);
}

function focusQuietly(element?: HTMLElement | null) {
  quietFocus = true;
  element?.focus();
  quietFocus = false;
}

function trapFocus(event: KeyboardEvent) {
  const stops = [...header.value?.querySelectorAll<HTMLElement>('a[href], button') ?? []]
    .filter(element => element.checkVisibility({ visibilityProperty: true }));
  const first = stops[0];
  const last = stops[stops.length - 1];
  if (!first || !last) return;
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault();
    first.focus();
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Tab' && sheetOpen.value) {
    trapFocus(event);
    return;
  }
  if (event.key !== 'Escape') return;
  if (open.value) {
    const key = open.value;
    const inside = panel(key)?.contains(document.activeElement);
    hide();
    if (inside) focusQuietly(trigger(key));
  } else if (sheetOpen.value) {
    sheetOpen.value = false;
    burger.value?.focus();
  }
}

function onPointerDown(event: PointerEvent) {
  if (open.value && !header.value?.contains(event.target as Node)) hide();
}

function onScroll() {
  tucked.value = tucked.value ? window.scrollY > 24 : window.scrollY > 96;
}

function measureSheet() {
  sheetTop.value = Math.max(0, header.value?.getBoundingClientRect().bottom ?? 0);
}

function lockScroll(locked: boolean) {
  const root = document.documentElement;
  const gap = window.innerWidth - root.clientWidth;
  root.style.overflow = locked ? 'hidden' : '';
  root.style.paddingRight = locked && gap > 0 ? `${gap}px` : '';
}

function onDesktop(event: MediaQueryListEvent) {
  if (event.matches) sheetOpen.value = false;
}

watch(sheetOpen, (value) => {
  if (value) {
    hide();
    nextTick(measureSheet);
  }
  lockScroll(value);
});

watch(() => route.fullPath, () => {
  hide();
  sheetOpen.value = false;
});

onMounted(() => {
  desktop = window.matchMedia('(min-width: 1024px)');
  desktop.addEventListener('change', onDesktop);
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', measureSheet);
  window.addEventListener('keydown', onKeydown);
  document.addEventListener('pointerdown', onPointerDown);
  onScroll();
});

onBeforeUnmount(() => {
  clearTimeout(closeTimer);
  desktop?.removeEventListener('change', onDesktop);
  window.removeEventListener('scroll', onScroll);
  window.removeEventListener('resize', measureSheet);
  window.removeEventListener('keydown', onKeydown);
  document.removeEventListener('pointerdown', onPointerDown);
  lockScroll(false);
});

const item = 'relative inline-flex min-h-11 cursor-pointer items-center gap-1.5 rounded-[10px] px-3.5 text-lg whitespace-nowrap transition-colors hover:text-stone-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:hover:text-stone-100';
const idle = 'text-stone-700 dark:text-stone-300';
const lit = 'text-stone-900 dark:text-stone-100';
</script>

<template>
  <header
    ref="header"
    class="site-nav sticky top-0 z-50 px-6 pt-5 pb-3"
    :class="{
      'dark': overDark,
      'is-floating': floating,
      'bg-stone-100 dark:bg-stone-950': sheetOpen,
    }"
    @pointerenter="onHeaderEnter"
    @pointerleave="onHeaderLeave"
    @focusout="onFocusOut"
  >
    <div v-show="!sheetOpen" class="nav-band" aria-hidden="true">
      <div class="nav-band__blur" />
      <div class="nav-band__blur" />
      <div class="nav-band__tint" />
    </div>

    <div class="nav-bar relative isolate z-10 mx-auto grid min-h-13 grid-cols-[minmax(0,1fr)_auto] items-center gap-4 lg:grid-cols-[auto_minmax(0,1fr)_auto] xl:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]">
      <div class="nav-bar__surface" aria-hidden="true" />

      <div class="flex justify-start">
        <UContextMenu :items="logoMenu">
          <NuxtLink
            to="/"
            aria-label="Lucity home"
            class="flex min-h-11 items-center rounded-lg focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-(--nav-accent)"
          >
            <LandingLogo :tucked="floating" class="text-2xl text-neutral-900 lg:text-[28px] dark:text-neutral-100" />
          </NuxtLink>
        </UContextMenu>
      </div>

      <nav aria-label="Main" class="hidden justify-center gap-1 lg:flex">
        <template v-for="menu in menus" :key="menu.key">
          <button
            type="button"
            :aria-expanded="open === menu.key"
            :aria-controls="`menu-${menu.key}`"
            :class="[
              item,
              open === menu.key || current === menu.key ? lit : idle,
              open === menu.key ? 'bg-stone-900/7 dark:bg-white/8' : '',
            ]"
            @pointerenter="onTriggerEnter($event, menu.key)"
            @focus="onTriggerFocus($event, menu.key)"
            @click="onTriggerClick(menu.key)"
            @keydown.down="onTriggerDown($event, menu.key)"
          >
            {{ menu.label }}
            <UIcon
              name="i-lucide-chevron-down"
              class="-mr-1 size-4 transition-transform duration-200 motion-reduce:transition-none"
              :class="{ 'rotate-180': open === menu.key }"
            />
            <NavSquiggle v-if="current === menu.key" />
          </button>
          <div
            :id="`menu-${menu.key}`"
            role="region"
            :aria-label="menu.label"
            class="nav-panel"
            :class="{ 'is-open': open === menu.key }"
            @click="onLinkClick"
          >
            <NavFeaturesPanel v-if="menu.key === 'features'" />
            <NavSolutionsPanel v-else-if="menu.key === 'solutions'" />
            <NavResourcesPanel v-else @search="openSearch" />
          </div>
        </template>

        <NuxtLink
          to="/pricing"
          :class="[item, current === 'pricing' ? lit : idle]"
          @pointerenter="onPlainEnter"
        >
          Pricing
          <NavSquiggle v-if="current === 'pricing'" />
        </NuxtLink>
      </nav>

      <div class="flex items-center justify-end gap-1.5">
        <a
          :href="githubUrl"
          target="_blank"
          rel="noopener"
          aria-label="Lucity on GitHub"
          class="hidden size-11 items-center justify-center rounded-[10px] text-stone-700 transition-colors hover:text-stone-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) lg:flex dark:text-stone-300 dark:hover:text-stone-100"
        >
          <UIcon name="i-simple-icons-github" class="size-5.5" />
        </a>
        <a
          :href="signInUrl"
          class="hidden min-h-11 items-center rounded-[10px] pr-3.5 pl-2 text-lg whitespace-nowrap text-stone-700 transition-colors hover:text-stone-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) lg:inline-flex dark:text-stone-300 dark:hover:text-stone-100"
        >
          Sign in
        </a>
        <a
          :href="signUpUrl"
          class="inline-flex min-h-11 items-center rounded-lg bg-primary px-3.5 text-base font-semibold whitespace-nowrap text-white max-[349px]:hidden min-[390px]:px-4.5 min-[390px]:text-lg transition-colors hover:bg-primary-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent)"
        >
          Get started
        </a>
        <button
          ref="burger"
          type="button"
          :aria-label="sheetOpen ? 'Close menu' : 'Open menu'"
          :aria-expanded="sheetOpen"
          aria-controls="menu-sheet"
          class="-mr-2.5 flex size-11 cursor-pointer items-center justify-center rounded-lg text-stone-900 focus-visible:outline-2 focus-visible:outline-(--nav-accent) lg:hidden dark:text-stone-100"
          @click="sheetOpen = !sheetOpen"
        >
          <UIcon :name="sheetOpen ? 'i-lucide-x' : 'i-lucide-menu'" class="size-6" />
        </button>
      </div>
    </div>

    <Transition name="nav-sheet">
      <div
        v-show="sheetOpen"
        id="menu-sheet"
        class="nav-sheet absolute inset-x-0 top-full overflow-y-auto overscroll-contain border-t border-stone-300/70 bg-stone-100 lg:hidden dark:border-stone-800 dark:bg-stone-950"
        :style="{ height: `calc(100dvh - ${sheetTop}px)` }"
        @click="onLinkClick"
      >
        <NavSheet :sign-in-url="signInUrl" :sign-up-url="signUpUrl" :github-url="githubUrl" />
      </div>
    </Transition>
  </header>
</template>

<style scoped>
.site-nav {
  --nav-accent: oklch(0.55 0.15 300);
  --nav-paper: oklch(0.99 0.006 80);
  --nav-surface: oklch(0.98 0.01 80 / .86);
  --nav-ring: oklch(0.25 0.04 55 / .10);
  --nav-shade: oklch(0.25 0.04 55 / .5);
  --nav-ease: cubic-bezier(0.2, 0.8, 0.2, 1);
}

.dark .site-nav,
.site-nav.dark {
  --nav-accent: oklch(0.76 0.13 300);
  --nav-paper: oklch(0.24 0.03 55);
  --nav-surface: oklch(0.25 0.03 55 / .86);
  --nav-ring: oklch(1 0 0 / .10);
  --nav-shade: oklch(0 0 0 / .7);
}

.nav-bar {
  max-width: 100%;
  padding-inline: 0;
  transition:
    max-width 0.5s var(--nav-ease),
    padding 0.5s var(--nav-ease);
}

.is-floating .nav-bar {
  max-width: 1120px;
  padding-inline: 20px 10px;
}

.nav-bar__surface {
  position: absolute;
  inset: -4px 0;
  z-index: -1;
  border-radius: 16px;
  background: var(--nav-surface);
  -webkit-backdrop-filter: blur(20px);
  backdrop-filter: blur(20px);
  box-shadow:
    0 0 0 1px var(--nav-ring),
    0 18px 36px -22px var(--nav-shade);
  opacity: 0;
  visibility: hidden;
  transition:
    opacity 0.3s ease,
    visibility 0s 0.3s;
}

.is-floating .nav-bar__surface {
  opacity: 1;
  visibility: visible;
  transition:
    opacity 0.3s ease,
    visibility 0s;
}

.nav-panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 50%;
  box-sizing: border-box;
  width: min(1000px, 100%);
  padding: 10px;
  border-radius: 22px;
  background: var(--nav-paper);
  box-shadow:
    0 40px 80px -32px var(--nav-shade),
    0 0 0 1px var(--nav-ring);
  translate: -50% -6px;
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition:
    opacity 0.12s ease,
    translate 0.12s ease,
    visibility 0s 0.12s;
}

.nav-panel.is-open {
  translate: -50% 0;
  opacity: 1;
  visibility: visible;
  pointer-events: auto;
  transition:
    opacity 0.18s ease-out,
    translate 0.18s ease-out,
    visibility 0s;
}

.is-floating .nav-panel {
  top: calc(100% + 10px);
}

.nav-sheet-enter-active,
.nav-sheet-leave-active {
  transition: opacity 0.16s ease;
}

.nav-sheet-enter-from,
.nav-sheet-leave-to {
  opacity: 0;
}

.nav-band {
  position: absolute;
  inset: 0 0 -2.75rem 0;
  z-index: 0;
  pointer-events: none;
}

.nav-band__blur,
.nav-band__tint {
  position: absolute;
  inset: 0;
}

.nav-band__blur:nth-child(1) {
  -webkit-backdrop-filter: blur(3px);
  backdrop-filter: blur(3px);
  -webkit-mask-image: linear-gradient(to bottom, #000 36%, rgb(0 0 0 / .957) 44%, rgb(0 0 0 / .844) 52%, rgb(0 0 0 / .684) 60%, rgb(0 0 0 / .5) 68%, rgb(0 0 0 / .316) 76%, rgb(0 0 0 / .156) 84%, rgb(0 0 0 / .043) 92%, transparent 100%);
  mask-image: linear-gradient(to bottom, #000 36%, rgb(0 0 0 / .957) 44%, rgb(0 0 0 / .844) 52%, rgb(0 0 0 / .684) 60%, rgb(0 0 0 / .5) 68%, rgb(0 0 0 / .316) 76%, rgb(0 0 0 / .156) 84%, rgb(0 0 0 / .043) 92%, transparent 100%);
}

.nav-band__blur:nth-child(2) {
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
  -webkit-mask-image: linear-gradient(to bottom, #000 4%, rgb(0 0 0 / .957) 11.75%, rgb(0 0 0 / .844) 19.5%, rgb(0 0 0 / .684) 27.25%, rgb(0 0 0 / .5) 35%, rgb(0 0 0 / .316) 42.75%, rgb(0 0 0 / .156) 50.5%, rgb(0 0 0 / .043) 58.25%, transparent 66%);
  mask-image: linear-gradient(to bottom, #000 4%, rgb(0 0 0 / .957) 11.75%, rgb(0 0 0 / .844) 19.5%, rgb(0 0 0 / .684) 27.25%, rgb(0 0 0 / .5) 35%, rgb(0 0 0 / .316) 42.75%, rgb(0 0 0 / .156) 50.5%, rgb(0 0 0 / .043) 58.25%, transparent 66%);
}

.nav-band__tint {
  background: var(--ui-bg);
  opacity: 0.9;
  transition: opacity 0.3s ease;
  -webkit-mask-image: linear-gradient(to bottom, #000 0%, rgb(0 0 0 / .957) 12.5%, rgb(0 0 0 / .844) 25%, rgb(0 0 0 / .684) 37.5%, rgb(0 0 0 / .5) 50%, rgb(0 0 0 / .316) 62.5%, rgb(0 0 0 / .156) 75%, rgb(0 0 0 / .043) 87.5%, transparent 100%);
  mask-image: linear-gradient(to bottom, #000 0%, rgb(0 0 0 / .957) 12.5%, rgb(0 0 0 / .844) 25%, rgb(0 0 0 / .684) 37.5%, rgb(0 0 0 / .5) 50%, rgb(0 0 0 / .316) 62.5%, rgb(0 0 0 / .156) 75%, rgb(0 0 0 / .043) 87.5%, transparent 100%);
}

.is-floating .nav-band__tint {
  opacity: 0.5;
}

@media (prefers-reduced-motion: reduce) {
  .nav-band__tint,
  .nav-bar,
  .nav-bar__surface,
  .is-floating .nav-bar__surface,
  .nav-panel,
  .nav-panel.is-open,
  .nav-sheet-enter-active,
  .nav-sheet-leave-active {
    transition: none;
  }
}
</style>
