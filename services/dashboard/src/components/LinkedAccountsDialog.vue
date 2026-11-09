<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { ExternalLink, Link2, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { ChatAppKind } from '@/gql/graphql';
import { chatAppKindLabel } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import BrandTile from '@/components/BrandTile.vue';
import CopyButton from '@/components/CopyButton.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import IconButton from '@/components/IconButton.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import QueryError from '@/components/QueryError.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { failureToast } from '@/components/ui/sonner';

const LinkedAccountsDocument = graphql(`
  query LinkedAccounts {
    me {
      chatAccounts {
        kind
        server
        handle
      }
    }
    chatApps {
      id
      name
      kind
      account
      url
    }
  }
`);

const LinkChatAccountDocument = graphql(`
  mutation LinkChatAccount($chatApp: ChatAppID!) {
    linkChatAccount(chatApp: $chatApp) {
      url
      code
      expiresAt
    }
  }
`);

const RemoveChatAccountDocument = graphql(`
  mutation RemoveChatAccount($kind: ChatAppKind!, $server: String) {
    removeChatAccount(kind: $kind, server: $server)
  }
`);

const open = defineModel<boolean>('open', { default: false });

interface Pending {
  app: { id: string; kind: ChatAppKind; account: string; url?: string | null };
  url?: string | null;
  code: string;
  expiresAt: string;
}

const pending = ref<Pending | null>(null);
const expired = ref(false);

const { result, loading, error, refetch } = useQuery(LinkedAccountsDocument, null, () => ({
  enabled: open.value,
  fetchPolicy: 'network-only',
  pollInterval: pending.value ? 3000 : 0,
}));

const accounts = computed(() => result.value?.me.chatAccounts ?? []);
const apps = computed(() => result.value?.chatApps ?? []);

function linked(kind: ChatAppKind, server?: string | null) {
  return accounts.value.find(account => account.kind === kind && (kind === ChatAppKind.Telegram || account.server === server)) ?? null;
}

const unlinkedApps = computed(() => {
  const seen = new Set<string>();
  return apps.value.filter((app) => {
    const key = app.kind === ChatAppKind.Telegram ? app.kind : `${app.kind}\n${app.url}`;
    if (seen.has(key) || linked(app.kind, app.url)) return false;
    seen.add(key);
    return true;
  });
});

function host(url?: string | null) {
  if (!url) return '';
  try {
    return new URL(url).host;
  } catch {
    return url;
  }
}

function accountTitle(kind: ChatAppKind, server?: string | null) {
  return server ? `${chatAppKindLabel(kind)} on ${host(server)}` : chatAppKindLabel(kind);
}

watch(open, (value) => {
  if (!value) {
    pending.value = null;
    expired.value = false;
  }
});

watch(accounts, () => {
  if (pending.value && linked(pending.value.app.kind, pending.value.app.url)) pending.value = null;
});

let expiry: ReturnType<typeof setTimeout> | undefined;

watch(pending, (value) => {
  clearTimeout(expiry);
  if (!value) return;
  expiry = setTimeout(() => {
    pending.value = null;
    expired.value = true;
  }, Math.max(0, new Date(value.expiresAt).getTime() - Date.now()));
});

const { mutate: startLink, loading: starting } = useMutation(LinkChatAccountDocument);
const { mutate: removeAccount } = useMutation(RemoveChatAccountDocument);
const startingApp = ref<string | null>(null);
const removing = ref<string | null>(null);

async function link(app: Pending['app']) {
  startingApp.value = app.id;
  expired.value = false;
  try {
    const res = await startLink({ chatApp: app.id });
    const started = res?.data?.linkChatAccount;
    if (started) pending.value = { app, url: started.url, code: started.code, expiresAt: started.expiresAt };
  } catch (e: unknown) {
    failureToast('Failed to start linking', e);
  } finally {
    startingApp.value = null;
  }
}

async function remove(kind: ChatAppKind, server?: string | null) {
  removing.value = `${kind}\n${server ?? ''}`;
  try {
    await removeAccount({ kind, server: server ?? null });
    await refetch();
  } catch (e: unknown) {
    failureToast('Failed to remove account', e);
  } finally {
    removing.value = null;
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-md']">
      <template v-if="pending">
        <DialogHeader>
          <DialogIcon :icon="Link2" />
          <DialogTitle>Link your {{ chatAppKindLabel(pending.app.kind) }} account</DialogTitle>
          <DialogDescription v-if="pending.url">Open the link where you use Telegram and tap Start. It works once, for 10 minutes.</DialogDescription>
          <DialogDescription v-else>Send this to {{ pending.app.account }} in a direct message. It works once, for 10 minutes.</DialogDescription>
        </DialogHeader>

        <div class="min-w-0 space-y-3">
          <Button v-if="pending.url" as-child class="w-full">
            <a :href="pending.url" target="_blank" rel="noopener">
              Open {{ pending.app.account }} in Telegram
              <ExternalLink />
            </a>
          </Button>
          <div class="flex items-center gap-2 rounded-md bg-muted/40 py-1.5 pr-1.5 pl-3">
            <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ pending.url ?? `link ${pending.code}` }}</span>
            <CopyButton :value="pending.url ?? `link ${pending.code}`" :label="pending.url ? 'Copy link' : 'Copy message'" />
          </div>
          <p class="flex items-center gap-2 text-xs text-muted-foreground">
            <Spinner :size="12" />
            Waiting for {{ chatAppKindLabel(pending.app.kind) }}…
          </p>
        </div>

        <DialogFooter>
          <Button type="button" variant="ghost" @click="pending = null">Cancel</Button>
        </DialogFooter>
      </template>

      <template v-else>
        <DialogHeader>
          <DialogIcon :icon="Link2" />
          <DialogTitle>Linked accounts</DialogTitle>
          <DialogDescription>Chat accounts that count as you when you talk to agents, in every workspace you are a member of.</DialogDescription>
        </DialogHeader>

        <div v-if="loading && !result" class="divide-y overflow-hidden rounded-lg border">
          <div v-for="width in ['w-28', 'w-36']" :key="width" class="flex items-center gap-3 px-3 py-2.5">
            <Skeleton class="size-8 rounded-sm" />
            <div class="flex-1 space-y-1.5">
              <Skeleton :class="['h-3', width]" />
              <Skeleton class="h-2.5 w-20" />
            </div>
          </div>
        </div>

        <QueryError v-else-if="error && !result" title="Couldn't load your accounts" :error="error" @retry="refetch()" />

        <template v-else>
          <p v-if="expired" class="text-sm text-muted-foreground">The link expired. Start it again.</p>

          <p v-if="accounts.length === 0 && unlinkedApps.length === 0" class="text-sm text-muted-foreground">
            This workspace has no chat apps yet, so there's nothing to link.
          </p>

          <div v-else class="divide-y overflow-hidden rounded-lg border">
            <div v-for="account in accounts" :key="`${account.kind}\n${account.server ?? ''}`" class="flex items-center gap-3 px-3 py-2.5">
              <BrandTile :name="chatAppKindLabel(account.kind)" />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">{{ accountTitle(account.kind, account.server) }}</p>
                <p class="truncate text-xs text-muted-foreground">{{ account.handle }}</p>
              </div>
              <IconButton
                :label="`Remove ${account.handle}`"
                size="icon-xs"
                class="shrink-0"
                :pending="removing === `${account.kind}\n${account.server ?? ''}`"
                @click="remove(account.kind, account.server)"
              >
                <Trash2 />
              </IconButton>
            </div>
            <div v-for="app in unlinkedApps" :key="app.id" class="flex items-center gap-3 px-3 py-2.5">
              <BrandTile :name="chatAppKindLabel(app.kind)" />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">{{ accountTitle(app.kind, app.url) }}</p>
                <p class="truncate text-xs text-muted-foreground">Not linked</p>
              </div>
              <Button variant="outline" size="xs" class="shrink-0" :pending="startingApp === app.id" :disabled="starting" @click="link(app)">Link</Button>
            </div>
          </div>
        </template>

        <DialogFooter>
          <Button type="button" variant="outline" @click="open = false">Done</Button>
        </DialogFooter>
      </template>
    </DialogContent>
  </Dialog>
</template>
