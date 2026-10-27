import { CreditCard, KeyRound, LayoutGrid, SlidersHorizontal, Users } from '@lucide/vue';

export function workspaceNavItems(isAdmin: boolean) {
  const items = [
    { key: 'projects', label: 'Projects', icon: LayoutGrid, to: { name: 'projects' } },
    { key: 'members', label: 'Members', icon: Users, to: { name: 'workspace', params: { drawer: 'members' } } },
    { key: 'api-tokens', label: 'API tokens', icon: KeyRound, to: { name: 'workspace', params: { drawer: 'api-tokens' } } },
    { key: 'billing', label: 'Billing', icon: CreditCard, to: { name: 'workspace', params: { drawer: 'billing' } } },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal, to: { name: 'workspace', params: { drawer: 'settings' } } },
  ];
  return isAdmin ? items : items.filter(item => item.key !== 'api-tokens');
}
