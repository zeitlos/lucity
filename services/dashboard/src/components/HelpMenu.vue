<script setup lang="ts">
import { Bug, BookOpen, CircleHelp, Github, Lightbulb } from 'lucide-vue-next';
import { useReportBug } from '@/composables/useReportBug';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const { report, requestFeature } = useReportBug();
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <button class="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
        <CircleHelp :size="16" />
      </button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end" class="min-w-48">
      <DropdownMenuItem @select="report()">
        <Bug :size="14" class="mr-2" />
        Report a bug
      </DropdownMenuItem>
      <DropdownMenuItem @select="requestFeature()">
        <Lightbulb :size="14" class="mr-2" />
        Request a feature
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem as-child>
        <a href="https://lucity.cloud/getting-started/quick-start" target="_blank" rel="noopener">
          <BookOpen :size="14" class="mr-2" />
          Documentation
        </a>
      </DropdownMenuItem>
      <DropdownMenuItem as-child>
        <a href="https://github.com/zeitlos/lucity" target="_blank" rel="noopener">
          <Github :size="14" class="mr-2" />
          GitHub
        </a>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
