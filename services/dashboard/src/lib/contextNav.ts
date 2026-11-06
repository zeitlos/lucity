import { environmentRoute, projectRoute } from '@/lib/routes';
import { Braces, Cable, CreditCard, KeyRound, Layers, LayoutGrid, SlidersHorizontal, Users, Workflow } from '@lucide/vue';

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

export function projectNavItems(projectId: string) {
  return [
    { key: 'environments', label: 'Environments', icon: Layers, to: projectRoute(projectId) },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal, to: projectRoute(projectId, 'settings') },
  ];
}

export function environmentNavItems(projectId: string, environmentId: string) {
  return [
    { key: 'canvas', label: 'Canvas', icon: Workflow, to: environmentRoute(environmentId) },
    { key: 'variables', label: 'Variables', icon: Braces, to: environmentRoute(environmentId, 'variables') },
    { key: 'settings', label: 'Settings', icon: SlidersHorizontal, to: environmentRoute(environmentId, 'settings') },
  ];
}
