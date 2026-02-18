<script setup lang="ts">
import { ChevronDown, Plus } from 'lucide-vue-next';
import { useEnvironment, type Environment } from '@/composables/useEnvironment';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
  DropdownMenuSeparator,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const { activeEnvironment, environments, setEnvironment } = useEnvironment();

function syncStatusVariant(status: string) {
  switch (status) {
    case 'SYNCED': return 'default';
    case 'PROGRESSING': return 'secondary';
    case 'OUT_OF_SYNC': return 'outline';
    case 'DEGRADED': return 'destructive';
    default: return 'outline';
  }
}

function handleSelect(env: Environment) {
  setEnvironment(env);
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger
      class="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-sm font-medium text-foreground transition-colors hover:bg-accent"
    >
      {{ activeEnvironment?.name ?? 'No environment' }}
      <ChevronDown :size="14" class="text-muted-foreground" />
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" class="w-56">
      <DropdownMenuCheckboxItem
        v-for="env in environments"
        :key="env.id"
        :checked="env.id === activeEnvironment?.id"
        @select="handleSelect(env)"
      >
        <div class="flex w-full items-center justify-between">
          <span>{{ env.name }}</span>
          <Badge
            :variant="syncStatusVariant(env.syncStatus)"
            class="ml-2 text-[10px]"
          >
            {{ env.syncStatus }}
          </Badge>
        </div>
      </DropdownMenuCheckboxItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem disabled class="text-muted-foreground">
        <Plus :size="14" class="mr-2" />
        New Environment
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
