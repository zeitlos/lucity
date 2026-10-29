import { Cable, CreditCard, KeyRound, LayoutGrid, SlidersHorizontal, Users } from '@lucide/vue';

export function workspaceNavItems(isAdmin: boolean, agents: boolean) {
  const items = [
    { key: 'projects', label: 'Projects', icon: LayoutGrid, to: { name: 'projects' } },
    { key: 'members', label: 'Members', icon: Users, to: { name: 'workspace', params: { drawer: 'members' } } },
    { key: 'connections', label: 'Connections', icon: Cable, to: { name: 'workspace', params: { drawer: 'connections' } } },
    { key: 'api-tokens', label: 'API tokens', icon: KeyRound, to: { name: 'workspace', params: { drawer: 'api-tokens' } } },
    { key: 'billing', label: 'Billing', icon: CreditCard, to: { name: 'workspace', params: { drawer: 'billing' } } },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal, to: { name: 'workspace', params: { drawer: 'settings' } } },
  ];
  return items.filter(item => (isAdmin || item.key !== 'api-tokens') && (agents || item.key !== 'connections'));
}
