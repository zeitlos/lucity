import type { RouteLocationNamedRaw, RouteLocationNormalizedLoaded } from 'vue-router';

export const RESERVED_WORKSPACE_IDS: ReadonlySet<string> = new Set([
  'api-tokens', 'app', 'assets', 'auth', 'billing', 'brand', 'checkout', 'connections', 'fonts',
  'graphql', 'login', 'members', 'playground', 'projects', 'settings', 'webhooks', 'workspace',
]);

function drawerParams(drawer?: string) {
  return drawer ? { tilde: '~', drawer } : {};
}

export function projectRoute(projectId: string, drawer?: string): RouteLocationNamedRaw {
  const [workspace = '', project = ''] = projectId.split('/');
  return { name: 'project', params: { workspace, project, ...drawerParams(drawer) } };
}

export function environmentRoute(environmentId: string, drawer?: string): RouteLocationNamedRaw {
  const [workspace = '', project = '', environment = ''] = environmentId.split('/');
  return { name: 'environment', params: { workspace, project, environment, ...drawerParams(drawer) } };
}

function param(route: RouteLocationNormalizedLoaded, key: string): string | undefined {
  const value = route.params[key];
  return (Array.isArray(value) ? value[0] : value) || undefined;
}

export function routeProjectId(route: RouteLocationNormalizedLoaded): string | undefined {
  const workspace = param(route, 'workspace');
  const project = param(route, 'project');
  return workspace && project ? `${workspace}/${project}` : undefined;
}

export function routeEnvironmentId(route: RouteLocationNormalizedLoaded): string | undefined {
  const project = routeProjectId(route);
  const environment = param(route, 'environment');
  return project && environment ? `${project}/${environment}` : undefined;
}

export function routeDrawer(route: RouteLocationNormalizedLoaded): string | undefined {
  return param(route, 'drawer');
}
