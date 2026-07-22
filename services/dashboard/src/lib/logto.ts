import LogtoClient, { UserScope } from '@logto/browser';

export let logto: LogtoClient;
export let apiResource = '';

export async function initLogto(): Promise<void> {
  const res = await fetch('/auth/config');
  const cfg = (await res.json()) as { endpoint: string; audience: string; spaClientId: string };

  apiResource = cfg.audience;
  logto = new LogtoClient({
    endpoint: cfg.endpoint,
    appId: cfg.spaClientId,
    resources: [cfg.audience],
    scopes: [
      UserScope.Email,
      UserScope.Profile,
      UserScope.Identities,
      UserScope.Organizations,
      UserScope.OrganizationRoles,
      'admin',
      'member',
      'deployer',
    ],
  });
}
