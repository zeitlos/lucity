<script setup lang="ts">
import { Bug, BookOpen, CircleHelp, ExternalLink, Lightbulb } from '@lucide/vue';
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
        <Bug />
        <span>Report a bug</span>
        <ExternalLink class="ml-auto size-3! text-muted-foreground" />
      </DropdownMenuItem>
      <DropdownMenuItem @select="requestFeature()">
        <Lightbulb />
        <span>Request a feature</span>
        <ExternalLink class="ml-auto size-3! text-muted-foreground" />
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem as-child>
        <a href="https://lucity.cloud/getting-started/concepts" target="_blank" rel="noopener">
          <BookOpen />
          <span>Documentation</span>
          <ExternalLink class="ml-auto size-3! text-muted-foreground" />
        </a>
      </DropdownMenuItem>
      <DropdownMenuItem as-child>
        <a href="https://github.com/zeitlos/lucity" target="_blank" rel="noopener">
          <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="M12 .5C5.37.5 0 5.78 0 12.292c0 5.211 3.438 9.63 8.205 11.188.6.111.82-.254.82-.567 0-.28-.01-1.022-.015-2.005-3.338.711-4.042-1.582-4.042-1.582-.546-1.361-1.335-1.725-1.335-1.725-1.087-.731.084-.716.084-.716 1.205.082 1.838 1.215 1.838 1.215 1.07 1.803 2.809 1.282 3.495.981.108-.763.417-1.282.76-1.577-2.665-.295-5.466-1.309-5.466-5.827 0-1.287.465-2.339 1.235-3.164-.135-.298-.54-1.497.105-3.121 0 0 1.005-.316 3.3 1.209.96-.262 1.98-.392 3-.398 1.02.006 2.04.136 3 .398 2.28-1.525 3.285-1.209 3.285-1.209.645 1.624.24 2.823.12 3.121.765.825 1.23 1.877 1.23 3.164 0 4.53-2.805 5.527-5.475 5.817.42.354.81 1.077.81 2.182 0 1.578-.015 2.846-.015 3.229 0 .315.21.689.825.573C20.565 21.917 24 17.498 24 12.292 24 5.78 18.627.5 12 .5z" />
          </svg>
          <span>GitHub</span>
          <ExternalLink class="ml-auto size-3! text-muted-foreground" />
        </a>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
