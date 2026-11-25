<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { ExternalLink, MessageSquare } from '@lucide/vue';
import { graphql } from '@/gql';
import { ChatAppKind } from '@/gql/graphql';
import { chatAppKinds, resourceName } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import BrandTile from '@/components/BrandTile.vue';
import ChoiceRows from '@/components/ChoiceRows.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import SecretInput from '@/components/SecretInput.vue';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { failureToast } from '@/components/ui/sonner';

const CreateChatAppDocument = graphql(`
  mutation CreateChatApp($name: String!, $kind: ChatAppKind!, $url: String, $token: String!) {
    createChatApp(name: $name, kind: $kind, url: $url, token: $token) {
      id
      name
      kind
      account
      url
      credentialHint
      status
      checkedAt
    }
  }
`);

const open = defineModel<boolean>('open', { default: false });

const emit = defineEmits<{
  (e: 'created'): void;
}>();

const router = useRouter();

const kind = ref<ChatAppKind>(ChatAppKind.Telegram);
const name = ref('');
const url = ref('');
const token = ref('');

const isMattermost = computed(() => kind.value === ChatAppKind.Mattermost);

watch(open, (value) => {
  if (!value) return;
  kind.value = ChatAppKind.Telegram;
  name.value = 'telegram';
  url.value = '';
  token.value = '';
});

watch(kind, (value, previous) => {
  if (!name.value || name.value === previous.toLowerCase()) name.value = value.toLowerCase();
});

function focusNextField() {
  const field = isMattermost.value && !url.value.trim() ? 'chat-app-url' : 'chat-app-token';
  document.getElementById(field)?.focus();
}

const { mutate: createChatApp, loading: pending } = useMutation(CreateChatAppDocument);

const canCreate = computed(() => !pending.value && !!name.value.trim() && !!token.value && (!isMattermost.value || !!url.value.trim()));

async function create() {
  if (!canCreate.value) return;
  try {
    const res = await createChatApp({
      name: name.value.trim(),
      kind: kind.value,
      url: isMattermost.value ? url.value.trim() : null,
      token: token.value,
    });
    const id = res?.data?.createChatApp.id;
    open.value = false;
    emit('created');
    if (id) router.push({ name: 'connection', params: { kind: 'chat-apps', name: resourceName(id) } });
  } catch (e: unknown) {
    failureToast('Failed to create chat app', e);
  }
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-md']">
      <form class="space-y-4" @submit.prevent="create">
        <DialogHeader>
          <DialogIcon :icon="MessageSquare" />
          <DialogTitle>New chat app</DialogTitle>
          <DialogDescription>Agents talk through this bot in the chats they are added to.</DialogDescription>
        </DialogHeader>

        <fieldset :disabled="pending" class="min-w-0 space-y-4">
          <div class="space-y-2">
            <p id="chat-app-kind" class="text-sm font-medium text-foreground">App</p>
            <ChoiceRows v-model="kind" :options="chatAppKinds" aria-labelledby="chat-app-kind" @pointer="focusNextField">
              <template #visual="{ option }">
                <BrandTile :name="option.label" />
              </template>
            </ChoiceRows>
          </div>

          <div v-if="isMattermost" class="space-y-2">
            <Label for="chat-app-url">Server URL</Label>
            <Input
              id="chat-app-url"
              v-model="url"
              class="font-mono text-sm"
              placeholder="https://chat.example.com"
              autocomplete="off"
              spellcheck="false"
            />
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between gap-2">
              <Label for="chat-app-token">Bot token</Label>
              <a
                v-if="!isMattermost"
                href="https://t.me/BotFather"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-1 rounded-sm text-xs text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              >
                Open @BotFather
                <ExternalLink :size="12" />
              </a>
            </div>
            <SecretInput id="chat-app-token" v-model="token" />
            <p v-if="isMattermost" class="text-xs text-muted-foreground">From a bot account under Integrations.</p>
          </div>

          <div class="space-y-2">
            <Label for="chat-app-name">Name</Label>
            <Input id="chat-app-name" v-model="name" :placeholder="kind.toLowerCase()" autocomplete="off" spellcheck="false" />
          </div>
        </fieldset>

        <DialogFooter>
          <Button type="button" variant="ghost" :disabled="pending" @click="open = false">Cancel</Button>
          <Button type="submit" :pending="pending" :disabled="!canCreate">
            {{ pending ? 'Creating…' : 'Create chat app' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
