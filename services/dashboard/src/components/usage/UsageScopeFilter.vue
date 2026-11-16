<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Check, ChevronDown, LayoutGrid, Search, X } from '@lucide/vue';
import IconButton from '@/components/IconButton.vue';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { formatCents, formatDay, sum, type UsageProject, type UsageScope } from '@/lib/usage';
import { cn } from '@/lib/utils';

const props = defineProps<{
  projects: UsageProject[];
  colors: Map<string, string>;
}>();

const scope = defineModel<UsageScope | null>({ required: true });

const open = ref(false);
const query = ref('');

watch(open, () => {
  query.value = '';
});

interface Option {
  key: string;
  label: string;
  note: string;
  amountCents: number;
  level: number;
  color?: string;
  scope: UsageScope | null;
}

const options = computed<Option[]>(() => {
  const q = query.value.trim().toLowerCase();
  const list: Option[] = [];

  if (!q) {
    list.push({ key: 'all', label: 'All projects', note: '', amountCents: sum(props.projects.map(project => project.amountCents)), level: 0, scope: null });
  }

  for (const project of props.projects) {
    const projectHit = !q || project.name.toLowerCase().includes(q);
    const environments = project.environments.filter(environment => projectHit || environment.name.toLowerCase().includes(q));
    if (!projectHit && environments.length === 0) continue;

    list.push({
      key: project.projectId,
      label: project.name,
      note: project.deletedAt ? `deleted ${formatDay(new Date(project.deletedAt))}` : '',
      amountCents: project.amountCents,
      level: 0,
      color: props.colors.get(project.projectId),
      scope: { project: project.projectId },
    });

    for (const environment of environments) {
      list.push({
        key: environment.environmentId,
        label: environment.name,
        note: environment.deletedAt ? `deleted ${formatDay(new Date(environment.deletedAt))}` : '',
        amountCents: environment.amountCents,
        level: 1,
        color: props.colors.get(environment.environmentId),
        scope: { project: project.projectId, environment: environment.environmentId },
      });
    }
  }

  return list;
});

const selectedKey = computed(() => scope.value ? scope.value.environment ?? scope.value.project : 'all');

const selectedProject = computed(() => props.projects.find(project => project.projectId === scope.value?.project));
const selectedEnvironment = computed(() => selectedProject.value?.environments.find(environment => environment.environmentId === scope.value?.environment));

function pick(option: Option) {
  scope.value = option.scope;
  open.value = false;
}
</script>

<template>
  <div class="flex items-center gap-1">
    <Popover v-model:open="open">
      <PopoverTrigger as-child>
        <button
          type="button"
          class="inline-flex h-9 items-center gap-2 rounded-md border bg-card pr-2.5 pl-3 text-sm text-foreground transition-colors hover:border-muted-foreground focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none"
        >
          <span
            v-if="scope"
            class="size-2 shrink-0 rounded-xs"
            :style="{ background: colors.get(scope.environment ?? scope.project) }"
          />
          <LayoutGrid v-else :size="16" class="text-muted-foreground" />
          <span class="font-medium whitespace-nowrap">{{ selectedProject?.name ?? 'All projects' }}</span>
          <template v-if="selectedEnvironment">
            <span class="text-muted-foreground">/</span>
            <span class="font-medium whitespace-nowrap">{{ selectedEnvironment.name }}</span>
          </template>
          <ChevronDown :size="14" class="text-muted-foreground" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" class="w-90 overflow-hidden p-0">
        <label class="flex h-11 items-center gap-2 border-b px-3">
          <Search :size="16" class="shrink-0 text-muted-foreground" />
          <input
            v-model="query"
            type="text"
            placeholder="Find a project or environment"
            aria-label="Find a project or environment"
            class="min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
          >
        </label>
        <div role="listbox" aria-label="Projects and environments" class="flex max-h-95 flex-col overflow-y-auto p-1">
          <button
            v-for="option in options"
            :key="option.key"
            type="button"
            role="option"
            :aria-selected="option.key === selectedKey"
            :class="cn(
              'flex min-h-9 w-full shrink-0 items-center gap-2 rounded-sm pr-2.5 text-left text-sm hover:bg-muted',
              option.level ? 'pl-8.5' : 'pl-2.5',
              option.key === selectedKey && 'bg-muted',
            )"
            @click="pick(option)"
          >
            <span class="flex size-4 shrink-0">
              <Check v-if="option.key === selectedKey" :size="16" />
            </span>
            <span class="size-2 shrink-0 rounded-xs" :style="{ background: option.color }" />
            <span :class="cn('truncate', option.level === 0 && 'font-medium')">{{ option.label }}</span>
            <span v-if="option.note" class="text-xs whitespace-nowrap text-muted-foreground">{{ option.note }}</span>
            <span class="ml-auto pl-3 text-muted-foreground tabular-nums">{{ formatCents(option.amountCents) }}</span>
          </button>
          <p v-if="options.length === 0" class="px-3 py-2.5 text-sm text-muted-foreground">
            No project or environment matches
          </p>
        </div>
      </PopoverContent>
    </Popover>
    <IconButton v-if="scope" label="Show all projects" variant="outline" size="icon" @click="scope = null">
      <X />
    </IconButton>
  </div>
</template>
