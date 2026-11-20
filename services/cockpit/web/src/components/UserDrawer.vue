<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { Check, ChevronLeft, Copy, ExternalLink, Globe, Play, X } from '@lucide/vue';

import GitHubIcon from '@/components/GitHubIcon.vue';
import UserAvatar from '@/components/UserAvatar.vue';
import { age, day, relativeTime } from '@/lib/format';
import {
  UserSessionsDocument,
  browserIcon,
  clockTime,
  countryName,
  dayLabel,
  displayName,
  flag,
  referrerHost,
  sessionLength,
  systemIcon,
  type User,
} from '@/lib/users';
import {
  avatarStyle,
  githubProfile,
  initial,
  openExternal,
  owner,
  plan,
  status,
  toneClass,
  youngAccount,
} from '@/lib/workspaces';

const ReplayOverlay = defineAsyncComponent(() => import('@/components/ReplayOverlay.vue'));

const props = defineProps<{
  user: User;
  canGoBack: boolean;
}>();

const emit = defineEmits<{
  back: [];
  close: [];
  workspace: [id: string];
}>();

const { result, loading, error, refetch } = useQuery(UserSessionsDocument, () => ({ userID: props.user.id }));

type Session = NonNullable<typeof result.value>['userSessions']['sessions'][number];

const sessions = computed(() => result.value?.userSessions.sessions ?? []);
const analyticsURL = computed(() => result.value?.userSessions.url ?? null);

const groups = computed(() => {
  const days: { label: string; sessions: typeof sessions.value }[] = [];

  for (const session of sessions.value) {
    const label = dayLabel(new Date(session.startedAt));
    const last = days[days.length - 1];

    if (last?.label === label) last.sessions.push(session);
    else days.push({ label, sessions: [session] });
  }

  return days;
});

const name = computed(() => displayName(props.user));

const youngOwner = computed(() => props.user.memberships.some(({ workspace }) => owner(workspace)?.id === props.user.id && youngAccount(workspace)));

const workspaceColumns = 'grid-cols-[minmax(0,1.3fr)_56px_minmax(0,1fr)_minmax(0,1fr)]';

const copied = ref(false);

async function copyID() {
  await navigator.clipboard.writeText(props.user.id);
  copied.value = true;
  setTimeout(() => (copied.value = false), 1500);
}

function openAnalytics() {
  if (analyticsURL.value) openExternal(analyticsURL.value);
}

const replaying = ref<Session | null>(null);

function openSession(session: Session) {
  if (session.hasReplay) replaying.value = session;
  else openAnalytics();
}

function sessionSummary(session: Session): string {
  const place = [session.city, countryName(session.country)].filter(Boolean).join(', ');
  const parts = [
    place,
    `${session.browser} on ${session.operatingSystem}`,
    session.language,
    `${session.pageviews} ${session.pageviews === 1 ? 'page' : 'pages'}`,
    session.referrer ? `from ${referrerHost(session.referrer)}` : '',
  ];

  return parts.filter(Boolean).join(' · ');
}

function replayDetail(session: Session): string {
  const started = new Date(session.startedAt);
  return `${dayLabel(started)} ${clockTime(started)} · ${sessionLength(session.duration)}`;
}
</script>

<template>
  <aside class="flex h-full flex-col border-l bg-card shadow-lg" :aria-label="`User ${name}`">
    <header class="flex h-13 shrink-0 items-center gap-2.5 border-b bg-sidebar pr-2.5 pl-4">
      <button v-if="canGoBack" type="button" class="-ml-2 flex size-8 shrink-0 items-center justify-center rounded-sm text-muted-foreground hover:bg-muted" aria-label="Back" @click="emit('back')">
        <ChevronLeft class="size-4" aria-hidden="true" />
      </button>
      <UserAvatar :src="user.avatar" :label="name" class="size-6" />
      <h2 class="truncate font-semibold">{{ name }}</h2>
      <button type="button" class="ml-auto flex size-8 shrink-0 items-center justify-center rounded-sm hover:bg-muted" aria-label="Close" @click="emit('close')">
        <X class="size-4" aria-hidden="true" />
      </button>
    </header>

    <div class="flex-1 overflow-y-auto">
      <dl class="grid grid-cols-3 divide-x divide-hairline border-b border-hairline">
        <div class="col-span-2 flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Email</dt>
          <dd class="truncate" :title="user.email ?? undefined">{{ user.email ?? 'None' }}</dd>
        </div>
        <div class="flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Signed up</dt>
          <dd :title="day(Date.parse(user.signedUpAt))">{{ relativeTime(user.signedUpAt) }}</dd>
        </div>
      </dl>
      <dl class="grid grid-cols-3 divide-x divide-hairline border-b border-hairline">
        <div class="col-span-2 flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">GitHub</dt>
          <dd v-if="user.account" class="flex min-w-0 items-center gap-1.5">
            <GitHubIcon class="size-3.5 shrink-0" />
            <a
              :href="githubProfile(user.account.login)"
              class="truncate hover:underline"
              @click.prevent="openExternal(githubProfile(user.account.login))"
            >{{ user.account.login }}</a>
            <span class="shrink-0" :class="youngOwner ? 'text-status-warn' : 'text-muted-foreground'">{{ age(user.account.createdAt) }} old</span>
          </dd>
          <dd v-else class="text-muted-foreground">Not connected</dd>
        </div>
        <div class="flex min-w-0 flex-col gap-0.5 px-4 py-3">
          <dt class="text-muted-foreground">Last sign-in</dt>
          <dd :title="user.lastSignInAt ? day(Date.parse(user.lastSignInAt)) : undefined">
            {{ user.lastSignInAt ? relativeTime(user.lastSignInAt) : 'Never' }}
          </dd>
        </div>
      </dl>

      <section class="pt-4">
        <div class="grid h-8 items-center gap-x-3 border-y border-hairline bg-sidebar px-4 whitespace-nowrap text-muted-foreground" :class="workspaceColumns">
          <span>Workspace</span>
          <span>Role</span>
          <span>Plan</span>
          <span>Status</span>
        </div>
        <div
          v-for="{ workspace, role } in user.memberships"
          :key="workspace.id"
          class="grid min-h-10 items-center gap-x-3 border-b border-hairline px-4"
          :class="workspaceColumns"
        >
          <button type="button" class="flex min-w-0 items-center gap-2.5 text-left hover:underline" @click="emit('workspace', workspace.id)">
            <span class="flex size-6 shrink-0 items-center justify-center rounded-xs font-semibold" :style="avatarStyle(workspace.id)" aria-hidden="true">
              {{ initial(workspace.id) }}
            </span>
            <span class="truncate">{{ workspace.id }}</span>
          </button>
          <span class="text-muted-foreground">{{ role }}</span>
          <span class="truncate" :class="toneClass[plan(workspace).tone]">{{ plan(workspace).text }}</span>
          <span class="truncate" :class="toneClass[status(workspace).tone]">
            {{ status(workspace).text }}<template v-if="status(workspace).note"> · {{ status(workspace).note }}</template>
          </span>
        </div>
      </section>

      <section class="pt-4">
        <h3 class="flex items-baseline gap-1.5 px-4 pt-1 pb-2 font-medium">
          Sessions <span v-if="sessions.length" class="font-normal text-muted-foreground">{{ sessions.length }}</span>
        </h3>

        <div v-if="error" class="mx-4 mb-4 flex items-start gap-3 rounded-sm border border-status-danger/40 px-4 py-3">
          <div class="flex min-w-0 flex-1 flex-col gap-1">
            <p class="font-medium text-status-danger">Sessions could not be loaded</p>
            <p class="text-muted-foreground">{{ error.message }}</p>
          </div>
          <button type="button" class="h-7 shrink-0 rounded-sm border bg-card px-2.5 hover:bg-muted" @click="refetch()">Retry</button>
        </div>

        <template v-else-if="loading && !result">
          <div v-for="index in 4" :key="index" class="flex flex-col gap-2 border-t border-hairline px-4 py-2.5">
            <span class="h-4 w-56 animate-pulse rounded-md bg-muted" />
            <span class="h-4 w-72 animate-pulse rounded-md bg-muted" />
          </div>
        </template>

        <p v-else-if="sessions.length === 0" class="border-y border-hairline px-4 py-2.5 text-muted-foreground">No sessions recorded</p>

        <template v-for="group in groups" :key="group.label">
          <div class="flex h-8 items-center border-y border-hairline bg-sidebar px-4 text-muted-foreground">{{ group.label }}</div>
          <button
            v-for="session in group.sessions"
            :key="session.id"
            type="button"
            class="group grid min-h-10 w-full grid-cols-[44px_minmax(0,1fr)_auto_44px_16px] items-center gap-x-3 border-b border-hairline px-4 text-left transition-colors hover:bg-accent"
            :title="sessionSummary(session)"
            :aria-label="session.hasReplay ? `Play the replay from ${clockTime(new Date(session.startedAt))}` : `Open the session from ${clockTime(new Date(session.startedAt))} in Rybbit`"
            @click="openSession(session)"
          >
            <span class="tabular-nums">{{ clockTime(new Date(session.startedAt)) }}</span>
            <span class="flex min-w-0 items-center gap-2">
              <span v-if="flag(session.country)" aria-hidden="true">{{ flag(session.country) }}</span>
              <span class="truncate">{{ session.city || countryName(session.country) || 'Unknown location' }}</span>
            </span>
            <span class="flex items-center gap-1.5">
              <img v-if="browserIcon(session.browser)" :src="browserIcon(session.browser)!" alt="" class="size-3.5">
              <Globe v-else class="size-3.5 text-muted-foreground" aria-hidden="true" />
              <img v-if="systemIcon(session.operatingSystem)" :src="systemIcon(session.operatingSystem)!" alt="" class="size-3.5">
            </span>
            <span class="text-right text-muted-foreground tabular-nums">{{ sessionLength(session.duration) }}</span>
            <Play v-if="session.hasReplay" class="size-3.5 fill-current" aria-hidden="true" />
            <ExternalLink v-else class="size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" aria-hidden="true" />
          </button>
        </template>
        <div class="h-6" />
      </section>
    </div>

    <footer class="flex h-14 shrink-0 items-center gap-2 border-t bg-sidebar px-4">
      <button type="button" class="flex h-8 min-w-0 items-center gap-2 rounded-sm border bg-card px-2.5 hover:bg-muted" title="Copy user ID" @click="copyID">
        <component :is="copied ? Check : Copy" class="size-3.5 shrink-0" aria-hidden="true" />
        <span class="truncate font-mono text-muted-foreground">{{ user.id }}</span>
      </button>
      <button
        v-if="user.account"
        type="button"
        class="ml-auto flex h-8 shrink-0 items-center gap-2 rounded-sm border bg-card px-3 hover:bg-muted"
        @click="openExternal(githubProfile(user.account.login))"
      >
        <GitHubIcon class="size-3.5" />
        GitHub
      </button>
      <button
        type="button"
        class="flex h-8 shrink-0 items-center gap-2 rounded-sm border bg-card px-3 hover:bg-muted disabled:opacity-50"
        :class="user.account ? '' : 'ml-auto'"
        :disabled="!analyticsURL"
        @click="openAnalytics"
      >
        <ExternalLink class="size-3.5" aria-hidden="true" />
        Rybbit
      </button>
    </footer>

    <ReplayOverlay
      v-if="replaying"
      :session-id="replaying.id"
      :title="name"
      :detail="replayDetail(replaying)"
      :analytics-url="analyticsURL"
      @close="replaying = null"
    />
  </aside>
</template>
