<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { AlertCircle, CircleArrowUp, Coins, CreditCard, ExternalLink, Users } from '@lucide/vue';
import { graphql } from '@/gql';
import { Plan, WorkspaceRole, type GatePlansQuery } from '@/gql/graphql';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import DialogIcon from '@/components/DialogIcon.vue';
import { failureToast } from '@/components/ui/sonner';
import { useAuth } from '@/composables/useAuth';
import { rememberPendingCreate, useGate } from '@/composables/useGate';
import { CapabilitiesDocument } from '@/composables/useCapabilities';
import { apolloClient } from '@/lib/apollo';
import { formatCpuQuantity, formatMoney, formatStorage } from '@/lib/format';
import { planTitle } from '@/lib/capabilities';

const GatePlansDocument = graphql(`
  query GatePlans {
    plans {
      plan
      priceCents
      currency
      limits {
        cpu
        memory
        replicas
        volumes
        volumeSize
        databaseStorage
        members
        customDomains
      }
    }
  }
`);

const GateSubscriptionDocument = graphql(`
  query GateSubscription {
    subscription {
      plan
      status
    }
  }
`);

const GateMembersDocument = graphql(`
  query GateMembers {
    workspace {
      id
      members {
        id
        name
        email
        role
      }
    }
  }
`);

const GateSelectPlanDocument = graphql(`
  mutation GateSelectPlan($plan: Plan!) {
    selectPlan(plan: $plan) {
      checkoutUrl
      subscription {
        plan
        status
      }
    }
  }
`);

const AddPaymentMethodDocument = graphql(`
  mutation AddPaymentMethod($returnPath: String!) {
    addPaymentMethod(returnPath: $returnPath) {
      url
    }
  }
`);

type Limits = GatePlansQuery['plans'][number]['limits'];

const groups: { label: string; rows: { label: string; value: (limits: Limits) => string }[] }[] = [
  {
    label: 'Per instance',
    rows: [
      { label: 'CPU', value: l => formatCpuQuantity(l.cpu) },
      { label: 'Memory', value: l => formatStorage(l.memory) },
      { label: 'Replicas per service', value: l => (l.replicas === 1 ? '1' : `Up to ${l.replicas}`) },
    ],
  },
  {
    label: 'Storage',
    rows: [
      { label: 'Volumes', value: l => (l.volumes ? `Up to ${formatStorage(l.volumeSize)} each` : 'Not included') },
      { label: 'Databases', value: l => `Up to ${formatStorage(l.databaseStorage)} each` },
    ],
  },
  {
    label: 'Workspace',
    rows: [
      { label: 'Members', value: l => (l.members === 1 ? 'Just you' : l.members == null ? 'Unlimited' : String(l.members)) },
      { label: 'Custom domains', value: l => (l.customDomains == null ? 'Unlimited' : l.customDomains === 0 ? 'None' : String(l.customDomains)) },
    ],
  },
];

const route = useRoute();
const router = useRouter();
const { request, close } = useGate();
const { activeWorkspace, isWorkspaceAdmin } = useAuth();

const open = computed({
  get: () => !!request.value,
  set: (value: boolean) => {
    if (!value) close();
  },
});

const planRequest = computed(() => (request.value?.kind === 'plan' ? request.value : null));
const cardRequest = computed(() => (request.value?.kind === 'card' ? request.value : null));
const selectedPlan = ref<Plan>(Plan.Hobby);

watch(planRequest, (value) => {
  if (value) selectedPlan.value = value.plan;
});

const comparing = computed(() => !!planRequest.value && isWorkspaceAdmin.value);
const asking = computed(() => !!planRequest.value && !isWorkspaceAdmin.value);

const { result: plansResult, loading: plansLoading } = useQuery(GatePlansDocument, null, () => ({ enabled: comparing.value }));
const { result: subscriptionResult } = useQuery(GateSubscriptionDocument, null, () => ({ enabled: comparing.value, errorPolicy: 'all' }));
const { result: membersResult } = useQuery(GateMembersDocument, null, () => ({ enabled: asking.value }));

const offers = computed(() => plansResult.value?.plans ?? []);
const currentPlan = computed(() => subscriptionResult.value?.subscription?.plan ?? null);
const admins = computed(() => (membersResult.value?.workspace.members ?? []).filter(m => m.role === WorkspaceRole.Admin));

const { mutate: selectPlanMutate, loading: upgrading } = useMutation(GateSelectPlanDocument);
const { mutate: addCardMutate, loading: addingCard } = useMutation(AddPaymentMethodDocument);

function price(offer: { priceCents: number; currency: string }) {
  return formatMoney(offer.priceCents / 100, offer.currency);
}

function selectable(plan: Plan | null | undefined): plan is Plan {
  return !!plan && plan !== currentPlan.value;
}

function cellClass(plan: Plan | null | undefined, index: number) {
  return [
    plan && plan === selectedPlan.value && 'bg-primary/5',
    index === offers.value.length - 1 ? 'pl-4 pr-6' : 'px-4',
  ];
}

async function upgrade() {
  try {
    const res = await selectPlanMutate({ plan: selectedPlan.value });
    const url = res?.data?.selectPlan?.checkoutUrl;

    if (url) {
      window.location.href = url;
      return;
    }

    await apolloClient.refetchQueries({ include: [CapabilitiesDocument] });
    close();
  } catch (e: unknown) {
    failureToast('Failed to change the plan', e);
  }
}

async function addCard() {
  const returnPath = cardRequest.value?.returnPath ?? route.fullPath;

  try {
    const res = await addCardMutate({ returnPath });
    const url = res?.data?.addPaymentMethod?.url;

    if (url) {
      rememberPendingCreate(returnPath);
      window.location.href = url;
    }
  } catch (e: unknown) {
    failureToast('Failed to start adding a card', e);
  }
}

function choosePlan() {
  close();
  router.push({ name: 'workspace', params: { drawer: 'billing' } });
}
</script>

<template>
  <Dialog v-model:open="open">
    <DialogContent v-if="planRequest && comparing" class="sm:max-w-4xl">
      <DialogHeader>
        <DialogIcon :icon="CircleArrowUp" />
        <DialogTitle>{{ planRequest.title }}</DialogTitle>
        <DialogDescription>
          Choose a plan for "{{ activeWorkspace }}". Its monthly fee comes back as credit for what your projects use.
        </DialogDescription>
      </DialogHeader>

      <div v-if="plansLoading && !offers.length" class="h-80" />

      <div v-else role="radiogroup" aria-label="Plan" class="-mx-6 grid grid-cols-4 text-sm">
        <div class="pl-6" />
        <label
          v-for="(offer, index) in offers"
          :key="offer.plan ?? 'trial'"
          class="flex flex-col gap-1 pb-3 pt-4"
          :class="[cellClass(offer.plan, index), selectable(offer.plan) && 'cursor-pointer']"
        >
          <span class="flex items-center justify-between">
            <span class="font-semibold">{{ offer.plan ? planTitle(offer.plan) : 'Trial' }}</span>
            <span v-if="(offer.plan ?? null) === currentPlan" class="text-xs text-muted-foreground">Current</span>
            <input
              v-else-if="selectable(offer.plan)"
              v-model="selectedPlan"
              type="radio"
              name="gate-plan"
              :value="offer.plan"
              class="accent-primary"
            >
          </span>
          <span v-if="offer.plan" class="text-xl font-semibold tabular-nums">
            {{ price(offer) }}<span class="text-xs font-normal text-muted-foreground"> a month</span>
          </span>
          <span v-else class="text-xl font-semibold">14 days</span>
          <span class="text-xs text-muted-foreground">
            {{ offer.plan ? `${price(offer)} credit included` : 'Free, until it runs out' }}
          </span>
        </label>

        <template v-for="(group, groupIndex) in groups" :key="group.label">
          <div class="pb-1.5 pl-6 text-xs font-medium text-muted-foreground" :class="groupIndex === 0 ? 'pt-3' : 'pt-4'">
            {{ group.label }}
          </div>
          <div v-for="(offer, index) in offers" :key="offer.plan ?? 'trial'" :class="cellClass(offer.plan, index)" />

          <template v-for="row in group.rows" :key="row.label">
            <div class="border-t py-2 pl-6 pr-3 text-muted-foreground">{{ row.label }}</div>
            <div
              v-for="(offer, index) in offers"
              :key="offer.plan ?? 'trial'"
              class="border-t py-2 tabular-nums"
              :class="cellClass(offer.plan, index)"
            >
              {{ row.value(offer.limits) }}
            </div>
          </template>
        </template>
      </div>

      <DialogFooter class="sm:items-center sm:justify-between">
        <p class="text-xs text-muted-foreground">
          The new plan applies right away, prorated for this period. Usage beyond the credit is billed at the published rates.
        </p>
        <div class="flex shrink-0 gap-2">
          <Button variant="outline" @click="close">Cancel</Button>
          <Button :pending="upgrading" @click="upgrade">Upgrade to {{ planTitle(selectedPlan) }}</Button>
        </div>
      </DialogFooter>
    </DialogContent>

    <DialogContent v-else-if="planRequest && asking" class="sm:max-w-md">
      <DialogHeader>
        <DialogIcon :icon="Users" />
        <DialogTitle>{{ planRequest.title }}</DialogTitle>
        <DialogDescription>
          Only admins of "{{ activeWorkspace }}" can change its plan. Ask one of them to upgrade to {{ planTitle(planRequest.plan) }}.
        </DialogDescription>
      </DialogHeader>

      <div v-if="admins.length" class="space-y-2">
        <p class="text-sm font-medium text-muted-foreground">Admins</p>
        <div class="divide-y rounded-lg border">
          <div v-for="admin in admins" :key="admin.id" class="flex flex-col px-3 py-2.5">
            <span class="text-sm">{{ admin.name || admin.email }}</span>
            <span v-if="admin.name" class="text-xs text-muted-foreground">{{ admin.email }}</span>
          </div>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="close">Close</Button>
      </DialogFooter>
    </DialogContent>

    <DialogContent v-else-if="cardRequest" class="sm:max-w-md">
      <DialogHeader>
        <DialogIcon :icon="CreditCard" />
        <DialogTitle>Add a card to start creating</DialogTitle>
        <DialogDescription>
          "{{ activeWorkspace }}" needs a card on file before it can run services, databases or other resources.
        </DialogDescription>
      </DialogHeader>

      <div v-if="cardRequest.error" role="alert" class="flex items-start gap-3 rounded-md bg-destructive/10 px-3.5 py-3">
        <AlertCircle :size="16" class="mt-0.5 shrink-0 text-destructive" />
        <span class="space-y-0.5">
          <span class="block text-sm font-medium">That card is prepaid</span>
          <span class="block text-xs text-muted-foreground">Use a debit or credit card, or choose a plan, which accepts prepaid cards. Nothing was charged.</span>
        </span>
      </div>

      <div v-else class="flex items-center gap-3">
        <span class="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary">
          <Coins :size="16" />
        </span>
        <span class="space-y-0.5">
          <span class="block text-sm font-medium">Nothing is charged yet</span>
          <span class="block text-xs text-muted-foreground">You're only billed once you choose a plan.</span>
        </span>
      </div>

      <DialogFooter>
        <Button v-if="cardRequest.error" variant="outline" @click="choosePlan">Choose a plan</Button>
        <Button v-else variant="outline" @click="close">Cancel</Button>
        <Button :pending="addingCard" @click="addCard">
          {{ cardRequest.error ? 'Try another card' : 'Add card' }}
          <ExternalLink />
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
