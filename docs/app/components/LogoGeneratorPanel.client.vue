<script setup lang="ts">
const { state, saved, panelOpen, ready, fontTick, persist, restore, reset } = useLogoGenerator();

const PANEL_KEY = 'lucity-logo-generator-panel-v1';

const pos = ref({ x: 24, y: 96 });
const collapsed = ref(false);
const status = ref('');
const panel = useTemplateRef<HTMLElement>('panel');

let offset = { x: 0, y: 0 };

const previewLockup = computed(() => {
  fontTick.value;
  return ready.value ? buildLogoLockup(state.value, +state.value.previewZoom, 'preview') : null;
});

function clampPos(x: number, y: number) {
  const width = panel.value?.offsetWidth ?? 320;
  const height = panel.value?.offsetHeight ?? 200;
  return {
    x: Math.min(Math.max(8, x), window.innerWidth - width - 8),
    y: Math.min(Math.max(8, y), window.innerHeight - Math.min(height, 120) - 8),
  };
}

function onDrag(event: PointerEvent) {
  pos.value = clampPos(event.clientX - offset.x, event.clientY - offset.y);
}

function endDrag() {
  window.removeEventListener('pointermove', onDrag);
  window.removeEventListener('pointerup', endDrag);
  savePanel();
}

function startDrag(event: PointerEvent) {
  offset = { x: event.clientX - pos.value.x, y: event.clientY - pos.value.y };
  window.addEventListener('pointermove', onDrag);
  window.addEventListener('pointerup', endDrag);
}

function savePanel() {
  try {
    localStorage.setItem(PANEL_KEY, JSON.stringify({
      pos: pos.value,
      collapsed: collapsed.value,
      open: panelOpen.value,
    }));
  } catch {}
}

function flash(message: string) {
  status.value = message;
  setTimeout(() => { if (status.value === message) status.value = ''; }, 2600);
}

function saveState() {
  const name = prompt('Name this state:', `state ${Object.keys(saved.value).length + 1}`);
  if (!name) return;
  saved.value = { ...saved.value, [name]: { ...state.value } };
  persist();
  flash(`Saved "${name}"`);
}

function loadState(name: string) {
  state.value = { ...state.value, ...saved.value[name] };
  flash(`Loaded "${name}"`);
}

function deleteState(name: string) {
  const next = { ...saved.value };
  delete next[name];
  saved.value = next;
  persist();
}

watch(state, persist, { deep: true });
watch([collapsed, panelOpen], savePanel);

onMounted(() => {
  restore();

  pos.value = { x: window.innerWidth - 344, y: 88 };

  try {
    const raw = localStorage.getItem(PANEL_KEY);
    if (raw) {
      const data = JSON.parse(raw);
      if (data.pos) pos.value = data.pos;
      if (typeof data.collapsed === 'boolean') collapsed.value = data.collapsed;
      if (typeof data.open === 'boolean') panelOpen.value = data.open;
    }
  } catch {}

  pos.value = clampPos(pos.value.x, pos.value.y);
  ready.value = true;

  document.fonts?.ready.then(() => { fontTick.value++; });
});
</script>

<template>
  <Teleport to="body">
    <section
      v-if="panelOpen"
      ref="panel"
      class="lugen"
      :style="{ left: `${pos.x}px`, top: `${pos.y}px` }"
    >
      <header class="lugen-bar" @pointerdown.prevent="startDrag">
        <span class="lugen-grip" />
        <span class="lugen-name">Logo generator</span>
        <button type="button" class="lugen-icon" title="Reset" @pointerdown.stop @click="reset">&#8634;</button>
        <button type="button" class="lugen-icon" :title="collapsed ? 'Expand' : 'Collapse'" @pointerdown.stop @click="collapsed = !collapsed">{{ collapsed ? '&#9634;' : '&#8212;' }}</button>
        <button type="button" class="lugen-icon" title="Close" @pointerdown.stop @click="panelOpen = false">&#10005;</button>
      </header>

      <div v-show="!collapsed" class="lugen-body">
        <div
          class="lugen-stage"
          :style="state.bgOn === 'Color' ? { background: state.bg } : {}"
        >
          <div
            v-if="previewLockup"
            class="lugen-art"
            :style="previewLockup.size"
          >
            <span class="lugen-mark" v-html="previewLockup.mark" />
            <span
              v-if="state.text"
              class="lugen-word"
              :style="previewLockup.text"
            >{{ state.text }}</span>
          </div>
        </div>

        <div v-for="[group, keys] in LOGO_GROUPS" :key="group" class="lugen-group">
          <div class="lugen-gl">{{ group }}</div>
          <div v-for="k in keys" :key="k" class="lugen-ctl">
            <label>
              <span>{{ LOGO_PARAMS[k].label }}</span>
              <span class="lugen-val">{{ state[k] }}</span>
            </label>

            <input v-if="LOGO_PARAMS[k].type === 'text'" v-model="state[k]" type="text">

            <select v-else-if="LOGO_PARAMS[k].type === 'select'" v-model="state[k]">
              <option v-for="opt in LOGO_PARAMS[k].options" :key="opt">{{ opt }}</option>
            </select>

            <div v-else-if="LOGO_PARAMS[k].type === 'swatch'" class="lugen-swrow">
              <button
                v-for="hex in LOGO_SWATCHES"
                :key="hex"
                type="button"
                class="lugen-sw"
                :class="{ sel: String(state[k]).toLowerCase() === hex.toLowerCase() }"
                :style="{ background: hex }"
                :title="hex"
                @click="state[k] = hex"
              />
              <input v-model="state[k]" type="color" title="Custom color">
            </div>

            <input
              v-else
              v-model.number="state[k]"
              type="range"
              :min="LOGO_PARAMS[k].min"
              :max="LOGO_PARAMS[k].max"
              :step="LOGO_PARAMS[k].step"
            >
          </div>
        </div>

        <div class="lugen-group">
          <div class="lugen-gl">Saved states</div>
          <button type="button" class="lugen-btn full" @click="saveState">Save current</button>
          <div v-for="(_, name) in saved" :key="name" class="lugen-staterow">
            <span class="lugen-sname" @click="loadState(name)">{{ name }}</span>
            <button type="button" class="lugen-icon" @click="deleteState(name)">&#10005;</button>
          </div>
        </div>

        <div class="lugen-status">{{ status }}</div>
      </div>
    </section>
  </Teleport>
</template>

<style scoped>
.lugen {
  position: fixed;
  z-index: 120;
  width: 320px;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--ui-border-accented);
  border-radius: 0.75rem;
  background: var(--ui-bg);
  box-shadow: 0 18px 48px oklch(0 0 0 / 0.22);
  font-size: 12px;
  color: var(--ui-text);
  overflow: hidden;
}

.lugen-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 8px;
  background: var(--ui-bg-elevated);
  border-bottom: 1px solid var(--ui-border);
  cursor: grab;
  touch-action: none;
  user-select: none;
}

.lugen-bar:active { cursor: grabbing; }

.lugen-grip {
  width: 10px;
  height: 10px;
  border-radius: 2px;
  background:
    radial-gradient(circle, var(--ui-text-dimmed) 1px, transparent 1px) 0 0 / 5px 5px;
}

.lugen-name {
  flex: 1;
  font-weight: 600;
  color: var(--ui-text-highlighted);
}

.lugen-icon {
  border: 0;
  background: transparent;
  color: var(--ui-text-dimmed);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
  line-height: 1;
}

.lugen-icon:hover {
  color: var(--ui-text-highlighted);
  background: var(--ui-bg-accented);
}

.lugen-body {
  padding: 10px;
  overflow-y: auto;
  max-height: min(72vh, 720px);
}

.lugen-tabs { display: flex; gap: 4px; margin-bottom: 8px; }

.lugen-tab {
  flex: 1;
  padding: 4px 8px;
  border: 1px solid var(--ui-border);
  border-radius: 6px;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
}

.lugen-tab.sel {
  border-color: var(--ui-primary);
  color: var(--ui-text-highlighted);
  font-weight: 600;
}

.lugen-stage {
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  min-height: 120px;
  max-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px;
  overflow: auto;
}

.lugen-art { flex: none; position: relative; }
.lugen-mark { position: absolute; inset: 0; }
.lugen-mark :deep(svg) { display: block; width: 100%; height: 100%; }
.lugen-word { position: absolute; line-height: 1; white-space: pre; }

.lugen-actions { display: flex; gap: 6px; margin: 8px 0 12px; }

.lugen-btn {
  flex: 1;
  font: inherit;
  font-weight: 600;
  padding: 5px 10px;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid var(--ui-border-accented);
  background: transparent;
  color: inherit;
}

.lugen-btn:hover { border-color: var(--ui-primary); color: var(--ui-text-highlighted); }
.lugen-btn.full { width: 100%; }

.lugen-group { margin-bottom: 14px; }

.lugen-gl {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--ui-text-dimmed);
  border-bottom: 1px solid var(--ui-border);
  padding-bottom: 4px;
  margin-bottom: 8px;
}

.lugen-ctl { margin-bottom: 8px; }
.lugen-ctl label { display: flex; justify-content: space-between; gap: 8px; margin-bottom: 2px; }

.lugen-val {
  font-variant-numeric: tabular-nums;
  color: var(--ui-text-dimmed);
  max-width: 55%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lugen-ctl input[type='range'] { width: 100%; accent-color: var(--ui-primary); }

.lugen-ctl input[type='text'],
.lugen-ctl select {
  width: 100%;
  padding: 4px 6px;
  border: 1px solid var(--ui-border);
  border-radius: 6px;
  background: var(--ui-bg);
  color: inherit;
  font: inherit;
}

.lugen-swrow { display: flex; flex-wrap: wrap; gap: 4px; align-items: center; }

.lugen-sw {
  width: 18px;
  height: 18px;
  border-radius: 5px;
  padding: 0;
  border: 1px solid var(--ui-border-accented);
  cursor: pointer;
}

.lugen-sw.sel { outline: 2px solid var(--ui-primary); outline-offset: 1px; }

.lugen-swrow input[type='color'] {
  width: 18px;
  height: 18px;
  padding: 0;
  border: 1px solid var(--ui-border-accented);
  border-radius: 50%;
  background: none;
  cursor: pointer;
}

.lugen-staterow { display: flex; align-items: center; gap: 6px; margin-top: 6px; }

.lugen-sname {
  flex: 1;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lugen-sname:hover { color: var(--ui-primary); }

.lugen-status { color: var(--ui-text-dimmed); min-height: 14px; }

</style>
