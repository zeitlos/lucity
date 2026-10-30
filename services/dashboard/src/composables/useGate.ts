import { ref } from 'vue';
import { Availability, Plan } from '@/gql/graphql';
import { planTitle, type Gate, type GatedChoice } from '@/lib/capabilities';

export type GateRequest =
  | { kind: 'plan'; plan: Plan; title: string }
  | { kind: 'card'; error?: string; returnPath?: string };

const request = ref<GateRequest | null>(null);

const PENDING_KEY = 'lucity_card_pending';

export function gateRequest(gate: Gate | undefined, title: string): GateRequest | null {
  if (!gate) return null;
  if (gate.availability === Availability.PlanRequired && gate.plan) return { kind: 'plan', plan: gate.plan, title };
  if (gate.availability === Availability.PaymentMethodRequired) return { kind: 'card' };
  return null;
}

export function gateFromError(error: unknown): GateRequest | null {
  const errors = (error as { graphQLErrors?: { message: string; extensions?: Record<string, unknown> }[] })?.graphQLErrors ?? [];

  for (const e of errors) {
    const code = e.extensions?.code;
    if (code === Availability.PlanRequired && typeof e.extensions?.plan === 'string') {
      return { kind: 'plan', plan: e.extensions.plan as Plan, title: e.message };
    }
    if (code === Availability.PaymentMethodRequired) return { kind: 'card' };
    if (code === 'PREPAID_CARD') return { kind: 'card', error: e.message };
  }

  return null;
}

export function rememberPendingCreate(path: string) {
  try {
    sessionStorage.setItem(PENDING_KEY, path);
  } catch {
    return;
  }
}

export function takePendingCreate(): string | null {
  try {
    const path = sessionStorage.getItem(PENDING_KEY);
    sessionStorage.removeItem(PENDING_KEY);
    return path;
  } catch {
    return null;
  }
}

export function useGate() {
  function open(next: GateRequest | null) {
    if (next) request.value = next;
  }

  function openFor(gate: Gate | undefined, title: string): boolean {
    const next = gateRequest(gate, title);
    open(next);
    return !!next;
  }

  function close() {
    request.value = null;
  }

  function pick<T extends { value: string; label: string }>(choices: GatedChoice<T>[], value: unknown, apply: (value: string) => void) {
    const choice = choices.find(c => c.value === value);
    if (!choice) return;

    if (choice.state.kind === 'plan') {
      open({ kind: 'plan', plan: choice.state.plan, title: `${choice.label} needs the ${planTitle(choice.state.plan)} plan` });
      return;
    }

    apply(choice.value);
  }

  return { request, open, openFor, close, pick };
}
