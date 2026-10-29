<script setup lang="ts">
import FilterPill from '@/components/FilterPill.vue';
import RangeFilterPanel from '@/components/RangeFilterPanel.vue';
import {
  choiceFilters,
  rangeFilters,
  type ChoiceFilter,
  type FilterOption,
  type Filters,
  type Range,
  type RangeFilter,
  type Workspace,
} from '@/lib/workspaces';

const props = defineProps<{
  workspaces: Workspace[];
}>();

const filters = defineModel<Filters>({ required: true });

const open: Range = { min: null, max: null };

function count(option: FilterOption): number {
  return props.workspaces.filter(option.test).length;
}

function chosen(filter: ChoiceFilter): string[] {
  return filters.value.choices[filter.key] ?? [];
}

function choiceSummary(filter: ChoiceFilter): string {
  return filter.options.filter((option) => chosen(filter).includes(option.value)).map((option) => option.label).join(', ');
}

function rangeSummary(filter: RangeFilter): string {
  const { min, max } = filters.value.ranges[filter.key] ?? open;

  if (min != null && max != null) return `${filter.format(min)}–${filter.format(max)}`;
  if (min != null) return `≥ ${filter.format(min)}`;
  if (max != null) return `≤ ${filter.format(max)}`;
  return '';
}

function toggle(filter: ChoiceFilter, value: string) {
  const values = chosen(filter);
  const next = values.includes(value) ? values.filter((candidate) => candidate !== value) : [...values, value];

  filters.value = { ...filters.value, choices: { ...filters.value.choices, [filter.key]: next } };
}

function clearChoices(filter: ChoiceFilter) {
  filters.value = { ...filters.value, choices: { ...filters.value.choices, [filter.key]: [] } };
}

function setRange(filter: RangeFilter, range: Range) {
  filters.value = { ...filters.value, ranges: { ...filters.value.ranges, [filter.key]: range } };
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <FilterPill
      v-for="filter in choiceFilters"
      :id="`filter-${filter.key}`"
      :key="filter.key"
      :label="filter.label"
      :summary="choiceSummary(filter)"
      @clear="clearChoices(filter)"
    >
      <div class="w-56">
        <label
          v-for="option in filter.options"
          :key="option.value"
          class="flex h-8 cursor-pointer items-center gap-2 rounded-sm px-2 hover:bg-muted"
        >
          <input
            type="checkbox"
            class="size-4 accent-foreground"
            :checked="chosen(filter).includes(option.value)"
            @change="toggle(filter, option.value)"
          >
          <component :is="option.icon" v-if="option.icon" class="size-4 text-muted-foreground" aria-hidden="true" />
          <span class="flex-1">{{ option.label }}</span>
          <span class="text-muted-foreground">{{ count(option) }}</span>
        </label>
      </div>
    </FilterPill>

    <FilterPill
      v-for="filter in rangeFilters"
      :id="`filter-${filter.key}`"
      :key="filter.key"
      :label="filter.label"
      :summary="rangeSummary(filter)"
      @clear="setRange(filter, open)"
    >
      <RangeFilterPanel
        :filter="filter"
        :workspaces="workspaces"
        :model-value="filters.ranges[filter.key] ?? open"
        @update:model-value="setRange(filter, $event)"
      />
    </FilterPill>
  </div>
</template>
