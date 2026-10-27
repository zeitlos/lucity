function storageKey(projectId: string) {
  return `lucity_last_environment_${projectId}`;
}

export function rememberEnvironment(projectId: string, environmentId: string) {
  localStorage.setItem(storageKey(projectId), environmentId);
}

export function preferredEnvironment<T extends { id: string }>(projectId: string, environments: T[]): T | undefined {
  const remembered = localStorage.getItem(storageKey(projectId));
  return environments.find(e => e.id === remembered) ?? environments[0];
}
