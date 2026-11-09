<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { MessageSquare } from '@lucide/vue';
import { graphql } from '@/gql';
import { ChatAppKind } from '@/gql/graphql';
import { chatAppKinds, resourceName } from '@/lib/agents';
import { mobileSheetClass } from '@/lib/dialog';
import DialogIcon from '@/components/DialogIcon.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
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
            <p class="text-sm font-medium text-foreground">App</p>
            <SegmentedControl v-model="kind" label="Chat app" :options="chatAppKinds" class="flex w-full" item-class="h-7 flex-1 text-sm" />
          </div>

          <div class="space-y-2">
            <Label for="chat-app-name">Name</Label>
            <Input id="chat-app-name" v-model="name" placeholder="telegram" autocomplete="off" spellcheck="false" />
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
            <Label for="chat-app-token">Bot token</Label>
            <Input id="chat-app-token" v-model="token" type="password" class="font-mono text-sm" autocomplete="off" data-1p-ignore />
            <p class="text-xs text-muted-foreground">
              <template v-if="isMattermost">From a bot account under Integrations. Checked by asking the server who the bot is.</template>
              <template v-else>From @BotFather. Checked by asking Telegram who the bot is.</template>
            </p>
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
