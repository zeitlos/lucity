import { Availability, Plan } from '@/gql/graphql';
import { parseQuantity } from '@/lib/format';

export interface Gate {
  availability: Availability;
  plan?: Plan | null;
}

export interface QuantityLimit {
  min: string;
  max: string;
  upgrades: { plan: Plan; max: string }[];
}

export type StepState = { kind: 'available' } | { kind: 'plan'; plan: Plan } | { kind: 'hidden' };

export function planTitle(plan: Plan): string {
  return plan === Plan.Pro ? 'Pro' : 'Hobby';
}

export function stepState(limit: QuantityLimit | undefined, value: string, current?: string): StepState {
  if (!limit) return { kind: 'available' };

  const amount = parseQuantity(value);

  if (amount < parseQuantity(limit.min)) return { kind: 'hidden' };
  if (amount <= parseQuantity(limit.max)) return { kind: 'available' };
  if (current && amount <= parseQuantity(current)) return { kind: 'available' };

  const upgrade = limit.upgrades.find(u => amount <= parseQuantity(u.max));

  return upgrade ? { kind: 'plan', plan: upgrade.plan } : { kind: 'hidden' };
}

export type GatedChoice<T> = T & { state: StepState };

export function gatedChoices<T extends { value: string }>(options: T[], limit: QuantityLimit | undefined, current?: string): GatedChoice<T>[] {
  return options
    .map(option => ({ ...option, state: stepState(limit, option.value, current) }))
    .filter(option => option.state.kind !== 'hidden');
}

export function aboveLimit(limit: QuantityLimit | undefined, current: string | undefined): boolean {
  if (!limit || !current) return false;
  return parseQuantity(current) > parseQuantity(limit.max);
}

export function nextUpgrade(limit: QuantityLimit | undefined): { plan: Plan; max: string } | undefined {
  return limit?.upgrades[limit.upgrades.length - 1];
}

export function usable(gate: Gate | undefined): boolean {
  return !gate || gate.availability === Availability.Available;
}

export function hidden(gate: Gate | undefined): boolean {
  return gate?.availability === Availability.Unavailable;
}
