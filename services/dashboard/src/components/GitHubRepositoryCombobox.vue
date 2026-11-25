<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useApolloClient, useQuery } from '@vue/apollo-composable';
import { Check, ChevronsUpDown, ExternalLink, Lock } from '@lucide/vue';
import {
  ComboboxAnchor,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxItemIndicator,
  ComboboxLabel,
  ComboboxPortal,
  ComboboxRoot,
  ComboboxTrigger,
  ComboboxViewport,
} from 'reka-ui';
import { graphql } from '@/gql';
import { brandIcon } from '@/lib/brandMarks';
import { useGitHubInstall } from '@/composables/useGitHubInstall';
import QueryError from '@/components/QueryError.vue';

const TriggerGitHubSourcesDocument = graphql(`
  query TriggerGitHubSources {
    githubSources {
      accountLogin
      accountAvatarUrl
    }
  }
`);

const TriggerGitHubRepositoriesDocument = graphql(`
  query TriggerGitHubRepositories($account: String!) {
    githubRepositories(account: $account) {
      id
      fullName
      private
    }
  }
`);

interface Repository {
  id: string;
  fullName: string;
  private: boolean;
}

interface Account {
  login: string;
  avatarUrl: string;
  repositories: Repository[];
}

const model = defineModel<string>({ required: true });

defineProps<{ id?: string }>();

const GitHubIcon = brandIcon('github');
const { resolveClient } = useApolloClient();

const { result: sourcesResult, error: sourcesError, refetch: refetchSources } = useQuery(TriggerGitHubSourcesDocument, null, { fetchPolicy: 'cache-and-network' });

const accounts = ref<Account[]>([]);
const loading = ref(true);
const reposError = ref<unknown>(null);

watch(() => sourcesResult.value?.githubSources, async (sources) => {
  if (!sources) return;
  loading.value = true;
  reposError.value = null;
  try {
    accounts.value = await Promise.all(sources.map(async (source) => {
      const { data } = await resolveClient().query({ query: TriggerGitHubRepositoriesDocument, variables: { account: source.accountLogin }, fetchPolicy: 'network-only' });
      return { login: source.accountLogin, avatarUrl: source.accountAvatarUrl, repositories: data?.githubRepositories ?? [] };
    }));
  } catch (e: unknown) {
    reposError.value = e;
  } finally {
    loading.value = false;
  }
}, { immediate: true });

const error = computed(() => sourcesError.value ?? reposError.value);

const { openInstallPopup } = useGitHubInstall(() => refetchSources());

function retry() {
  void refetchSources();
}
</script>

<template>
  <div class="space-y-2">
    <ComboboxRoot v-model="model" open-on-click open-on-focus reset-search-term-on-blur class="relative">
      <ComboboxAnchor class="flex h-9 w-full items-center gap-2 rounded-md border border-input bg-background px-3 text-sm focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2 focus-within:ring-offset-background">
        <component :is="GitHubIcon" v-if="GitHubIcon" class="size-3.5 shrink-0 text-foreground" />
        <ComboboxInput
          :id="id"
          :display-value="(value: string) => value"
          class="min-w-0 flex-1 bg-transparent font-mono text-sm outline-none placeholder:font-sans placeholder:text-muted-foreground"
          placeholder="Search repositories"
          autocomplete="off"
          spellcheck="false"
        />
        <ComboboxTrigger class="shrink-0 text-muted-foreground" aria-label="Show repositories">
          <ChevronsUpDown :size="14" />
        </ComboboxTrigger>
      </ComboboxAnchor>

      <ComboboxPortal>
        <ComboboxContent
          position="popper"
          :side-offset="6"
          class="z-50 w-(--reka-combobox-trigger-width) overflow-hidden rounded-lg border bg-popover text-popover-foreground shadow-md"
        >
          <ComboboxViewport class="max-h-72 overflow-y-auto p-1">
            <p v-if="loading && accounts.length === 0" class="px-2 py-1.5 text-sm text-muted-foreground">Loading repositories…</p>
            <ComboboxEmpty v-else class="px-2 py-1.5 text-sm text-muted-foreground">No repository matches.</ComboboxEmpty>
            <ComboboxGroup v-for="account in accounts" :key="account.login" class="py-1">
              <ComboboxLabel class="flex items-center gap-1.5 px-2 pt-1 pb-1 text-xs font-medium text-muted-foreground">
                <img :src="account.avatarUrl" alt="" class="size-4 rounded-full">
                {{ account.login }}
              </ComboboxLabel>
              <ComboboxItem
                v-for="repository in account.repositories"
                :key="repository.id"
                :value="repository.fullName"
                :text-value="repository.fullName"
                class="flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none select-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground"
              >
                <span class="truncate font-mono">{{ repository.fullName }}</span>
                <Lock v-if="repository.private" :size="12" class="shrink-0 text-muted-foreground" />
                <ComboboxItemIndicator class="ml-auto">
                  <Check :size="16" />
                </ComboboxItemIndicator>
              </ComboboxItem>
            </ComboboxGroup>
          </ComboboxViewport>
          <div class="flex items-center gap-1.5 border-t px-3 py-2 text-xs text-muted-foreground">
            Missing a repository?
            <button type="button" class="flex items-center gap-1 rounded-sm text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none" @click="openInstallPopup">
              Add it to the Lucity GitHub App
              <ExternalLink :size="12" />
            </button>
          </div>
        </ComboboxContent>
      </ComboboxPortal>
    </ComboboxRoot>

    <QueryError v-if="error && accounts.length === 0" title="Couldn't load your GitHub repositories" :error="error" @retry="retry" />
  </div>
</template>
