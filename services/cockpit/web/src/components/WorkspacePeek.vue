<script setup lang="ts">
import { ref, watch } from 'vue';
import { Ban, ChevronDown, ChevronUp, RotateCcw, X } from '@lucide/vue';

import EnvironmentLogs from '@/components/EnvironmentLogs.vue';
import { age, bytes, day, memory, relativeTime } from '@/lib/format';
import {
  avatarStyle,
  githubProfile,
  initial,
  openExternal,
  owner,
  plan,
  signalIcons,
  signalLabel,
  status,
  toneClass,
  youngAccount,
  type Workspace,
} from '@/lib/workspaces';

type Member = Workspace['members'][number];

const props = defineProps<{
  workspace: Workspace;
  position: number;
  total: number;
}>();

const emit = defineEmits<{
  close: [];
  previous: [];
  next: [];
  mark: [];
  clear: [];
}>();

const openLogs = ref(new Set<string>());

watch(() => props.workspace.id, () => {
  openLogs.value = new Set();
});

function toggleLogs(environmentID: string) {
  const next = new Set(openLogs.value);

  if (next.has(environmentID)) next.delete(environmentID);
  else next.add(environmentID);

  openLogs.value = next;
}

function short(commit: string): string {
  return commit.slice(0, 7);
}

const memberColumns = 'grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)_64px_96px]';
const serviceColumns = 'grid-cols-[minmax(0,1fr)_minmax(0,2.2fr)_64px]';

function youngOwner(member: Member): boolean {
  return member === owner(props.workspace) && youngAccount(props.workspace);
}

function isGitHub(url: string | null | undefined): url is string {
  return url?.startsWith('https://github.com/') ?? false;
}

function repositoryName(url: string): string {
  return url.replace(/^https:\/\//, '').replace(/\.git$/, '');
}

function commitURL(repository: string, commit: string): string {
  return `${repository.replace(/\.git$/, '')}/commit/${commit}`;
}
</script>

<template>
  <aside
    class="fixed inset-y-0 right-0 z-20 flex w-[480px] flex-col border-l bg-card shadow-lg"
    :aria-label="`Workspace ${workspace.id}`"
  >
    <header class="flex h-13 shrink-0 items-center gap-2.5 border-b bg-sidebar pr-2.5 pl-4">
      <span class="flex size-6 shrink-0 items-center justify-center rounded-xs font-semibold" :style="avatarStyle(workspace.id)" aria-hidden="true">
        {{ initial(workspace.id) }}
      </span>
      <h2 class="truncate font-semibold">{{ workspace.id }}</h2>
      <span class="shrink-0 text-muted-foreground" :title="day(Date.parse(workspace.createdAt))">Created {{ relativeTime(workspace.createdAt) }}</span>
      <button type="button" class="ml-auto flex size-8 shrink-0 items-center justify-center rounded-sm hover:bg-muted" aria-label="Close" @click="emit('close')">
        <X class="size-4" aria-hidden="true" />
      </button>
    </header>

    <div class="flex-1 overflow-y-auto">
      <dl class="grid grid-cols-3 divide-x divide-hairline border-b border-hairline">
        <div class="flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Status</dt>
          <dd :class="toneClass[status(workspace).tone]">{{ status(workspace).text }}</dd>
          <dd v-if="status(workspace).note" class="text-muted-foreground">{{ status(workspace).note }}</dd>
        </div>
        <div class="col-span-2 flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Plan</dt>
          <dd :class="toneClass[plan(workspace).tone]">{{ plan(workspace).text }}</dd>
          <dd v-if="workspace.billing" class="text-muted-foreground">
            {{ workspace.billing.hasPaymentMethod ? 'Payment method on file' : 'No payment method' }}
          </dd>
        </div>
      </dl>
      <dl class="grid grid-cols-3 divide-x divide-hairline border-b border-hairline">
        <div class="flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Pods</dt>
          <dd class="tabular-nums">
            {{ workspace.runningPods }}<span v-if="workspace.pendingPods" class="text-status-warn"> + {{ workspace.pendingPods }} pending</span>
          </dd>
        </div>
        <div class="flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Requests</dt>
          <dd class="tabular-nums">{{ memory(workspace.memoryRequested) }}</dd>
        </div>
        <div class="flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Egress 7d</dt>
          <dd class="tabular-nums">{{ bytes(workspace.egress7d) }}</dd>
        </div>
      </dl>

      <p v-if="workspace.abuseNote" class="mx-4 mt-4 flex gap-2 rounded-sm bg-status-danger-background px-3 py-2 text-status-danger">
        <Ban class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
        {{ workspace.abuseNote }}
      </p>

      <section class="border-b border-hairline py-3">
        <h3 class="flex items-baseline gap-1.5 px-4 py-1 font-medium">
          Signals <span v-if="workspace.signals.length" class="font-normal text-muted-foreground">{{ workspace.signals.length }}</span>
        </h3>
        <p v-if="!workspace.signals.length" class="px-4 py-1 text-muted-foreground">No signals</p>
        <ul v-else>
          <li v-for="signal in workspace.signals" :key="signal.kind + signal.detail" class="flex items-center gap-2.5 px-4 py-1">
            <component
              :is="signalIcons[signal.kind]"
              class="size-4 shrink-0"
              :class="signal.strong ? 'text-status-danger' : 'text-status-warn'"
              :stroke-width="2.25"
              aria-hidden="true"
            />
            <span class="min-w-0 truncate">{{ signalLabel(signal) }}</span>
          </li>
        </ul>
      </section>

      <section class="border-b border-hairline pt-3">
        <h3 class="flex items-baseline gap-1.5 px-4 pt-1 pb-2 font-medium">
          Members <span class="font-normal text-muted-foreground">{{ workspace.members.length }}</span>
        </h3>
        <div class="grid h-8 items-center gap-x-3 border-y border-hairline bg-sidebar px-4 whitespace-nowrap text-muted-foreground" :class="memberColumns">
          <span>Member</span>
          <span>Email</span>
          <span>Role</span>
          <span>Account age</span>
        </div>
        <div
          v-for="member in workspace.members"
          :key="(member.email ?? '') + (member.account?.login ?? '')"
          class="grid min-h-9 items-center gap-x-3 border-b border-hairline px-4 last:border-0"
          :class="memberColumns"
        >
          <a
            v-if="member.account"
            :href="githubProfile(member.account.login)"
            class="truncate hover:underline"
            @click.prevent="openExternal(githubProfile(member.account.login))"
          >{{ member.account.login }}</a>
          <span v-else class="truncate">{{ member.name }}</span>
          <span class="truncate text-muted-foreground" :title="member.email ?? undefined">{{ member.email }}</span>
          <span class="text-muted-foreground">{{ member.role }}</span>
          <span :class="youngOwner(member) ? 'text-status-warn' : 'text-muted-foreground'">
            {{ member.account ? age(member.account.createdAt) : '' }}
          </span>
        </div>
      </section>

      <section class="pt-3">
        <h3 class="flex items-baseline gap-1.5 px-4 pt-1 pb-2 font-medium">
          Environments <span v-if="workspace.environments.length" class="font-normal text-muted-foreground">{{ workspace.environments.length }}</span>
        </h3>
        <p v-if="!workspace.environments.length" class="px-4 pb-3 text-muted-foreground">No environments</p>
        <template v-else>
          <div class="grid h-8 items-center gap-x-3 border-y border-hairline bg-sidebar px-4 whitespace-nowrap text-muted-foreground" :class="serviceColumns">
            <span>Service</span>
            <span>Source</span>
            <span>Commit</span>
          </div>
          <template v-for="environment in workspace.environments" :key="environment.id">
            <div class="flex min-h-10 items-center gap-2 border-b border-hairline px-4">
              <span class="truncate font-medium">{{ environment.project }} / {{ environment.name }}</span>
              <span class="shrink-0 text-muted-foreground tabular-nums">
                {{ environment.runningPods }} running<template v-if="environment.pendingPods">, {{ environment.pendingPods }} pending</template>
              </span>
              <button type="button" class="ml-auto h-7 shrink-0 rounded-sm border px-2.5 hover:bg-muted" @click="toggleLogs(environment.id)">
                {{ openLogs.has(environment.id) ? 'Hide logs' : 'Logs' }}
              </button>
            </div>
            <div
              v-for="service in environment.services"
              :key="service.name"
              class="grid min-h-9 items-center gap-x-3 border-b border-hairline px-4"
              :class="serviceColumns"
            >
              <span class="truncate">{{ service.name }}</span>
              <a
                v-if="isGitHub(service.repository)"
                :href="service.repository"
                class="truncate font-mono text-muted-foreground hover:text-foreground hover:underline"
                @click.prevent="openExternal(service.repository)"
              >{{ repositoryName(service.repository) }}</a>
              <span v-else class="truncate font-mono text-muted-foreground" :title="(service.repository ?? service.image) ?? undefined">
                {{ service.repository ?? service.image }}
              </span>
              <a
                v-if="service.commit && isGitHub(service.repository)"
                :href="commitURL(service.repository, service.commit)"
                class="font-mono text-muted-foreground hover:text-foreground hover:underline"
                @click.prevent="openExternal(commitURL(service.repository, service.commit))"
              >{{ short(service.commit) }}</a>
              <span v-else class="font-mono text-muted-foreground">{{ service.commit ? short(service.commit) : '' }}</span>
            </div>
            <div v-if="openLogs.has(environment.id)" class="border-b border-hairline px-4 py-2">
              <EnvironmentLogs :environment-id="environment.id" />
            </div>
          </template>
        </template>
      </section>
    </div>

    <footer class="flex h-14 shrink-0 items-center gap-2 border-t bg-sidebar px-4">
      <button type="button" class="flex size-8 items-center justify-center rounded-sm border bg-card hover:bg-muted" aria-label="Previous workspace" @click="emit('previous')">
        <ChevronUp class="size-4" aria-hidden="true" />
      </button>
      <button type="button" class="flex size-8 items-center justify-center rounded-sm border bg-card hover:bg-muted" aria-label="Next workspace" @click="emit('next')">
        <ChevronDown class="size-4" aria-hidden="true" />
      </button>
      <span class="text-muted-foreground tabular-nums">{{ position }} of {{ total }}</span>
      <span class="flex gap-1 text-muted-foreground" aria-hidden="true">
        <kbd class="rounded-xs border px-1.5 font-mono">J</kbd>
        <kbd class="rounded-xs border px-1.5 font-mono">K</kbd>
      </span>
      <button v-if="workspace.abuse" type="button" class="ml-auto flex h-8 items-center gap-2 rounded-sm border bg-card px-3 hover:bg-muted" @click="emit('clear')">
        <RotateCcw class="size-4" aria-hidden="true" />
        Clear abuse
      </button>
      <button v-else type="button" class="ml-auto flex h-8 items-center gap-2 rounded-sm bg-destructive px-3 font-medium text-white" @click="emit('mark')">
        <Ban class="size-4" aria-hidden="true" />
        Mark as abuse
      </button>
    </footer>
  </aside>
</template>
