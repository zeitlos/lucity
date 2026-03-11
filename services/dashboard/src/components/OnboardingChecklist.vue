<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink } from 'vue-router';
import { Github, FolderPlus, Rocket, Check, X } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';

const props = defineProps<{
  githubConnected: boolean;
  hasProjects: boolean;
  hasDeployments: boolean;
  firstProjectId?: string;
}>();

defineEmits<{
  (e: 'dismiss'): void;
  (e: 'create-project'): void;
}>();

const completedCount = computed(() =>
  [props.githubConnected, props.hasProjects, props.hasDeployments].filter(Boolean).length,
);

const allComplete = computed(() => completedCount.value === 3);
</script>

<template>
  <div class="rounded-lg border border-border bg-card p-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-sm font-medium text-foreground">Getting started</span>
        <span class="text-xs text-muted-foreground">{{ completedCount }} of 3</span>
      </div>
      <button
        class="text-muted-foreground transition-colors hover:text-foreground"
        @click="$emit('dismiss')"
      >
        <X :size="14" />
      </button>
    </div>

    <div class="mt-3 grid gap-2 sm:grid-cols-3">
      <!-- Step 1: Connect GitHub -->
      <div
        class="flex items-center gap-3 rounded-md border px-3 py-2"
        :class="githubConnected ? 'border-primary/20 bg-primary/5' : ''"
      >
        <div
          class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full"
          :class="githubConnected ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'"
        >
          <Check v-if="githubConnected" :size="12" />
          <Github v-else :size="12" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-xs font-medium"
            :class="githubConnected ? 'text-muted-foreground line-through' : 'text-foreground'"
          >
            Connect GitHub
          </p>
        </div>
        <a
          v-if="!githubConnected"
          href="/auth/github/connect"
          class="shrink-0"
        >
          <Button variant="ghost" size="sm" class="h-6 px-2 text-xs">
            Connect
          </Button>
        </a>
      </div>

      <!-- Step 2: Create project -->
      <div
        class="flex items-center gap-3 rounded-md border px-3 py-2"
        :class="hasProjects ? 'border-primary/20 bg-primary/5' : ''"
      >
        <div
          class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full"
          :class="hasProjects ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'"
        >
          <Check v-if="hasProjects" :size="12" />
          <FolderPlus v-else :size="12" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-xs font-medium"
            :class="hasProjects ? 'text-muted-foreground line-through' : 'text-foreground'"
          >
            Create a project
          </p>
        </div>
        <Button
          v-if="!hasProjects"
          variant="ghost"
          size="sm"
          class="h-6 shrink-0 px-2 text-xs"
          @click="$emit('create-project')"
        >
          Create
        </Button>
      </div>

      <!-- Step 3: Deploy -->
      <div
        class="flex items-center gap-3 rounded-md border px-3 py-2"
        :class="hasDeployments ? 'border-primary/20 bg-primary/5' : ''"
      >
        <div
          class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full"
          :class="hasDeployments ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'"
        >
          <Check v-if="hasDeployments" :size="12" />
          <Rocket v-else :size="12" />
        </div>
        <div class="min-w-0 flex-1">
          <p
            class="text-xs font-medium"
            :class="hasDeployments ? 'text-muted-foreground line-through' : 'text-foreground'"
          >
            Deploy to dev
          </p>
        </div>
        <RouterLink
          v-if="!hasDeployments && firstProjectId"
          :to="{ name: 'project', params: { id: firstProjectId } }"
        >
          <Button variant="ghost" size="sm" class="h-6 px-2 text-xs">
            View
          </Button>
        </RouterLink>
      </div>
    </div>

    <p
      v-if="allComplete"
      class="mt-2 text-xs text-muted-foreground"
    >
      You're all set! <button class="underline hover:text-foreground" @click="$emit('dismiss')">Dismiss</button>
    </p>
  </div>
</template>
