import { reactive } from 'vue';

export type View = 'workspaces' | 'platform';

export const views: { key: View; name: string }[] = [
  { key: 'workspaces', name: 'Workspaces' },
  { key: 'platform', name: 'Platform' },
];

export const dock = reactive({
  view: 'workspaces' as View,
  hidden: false,
  right: 0,
});
