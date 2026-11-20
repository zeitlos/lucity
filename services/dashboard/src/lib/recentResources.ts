const LIMIT = 5;

function storageKey(workspace: string) {
  return `lucity_recent_resources_${workspace}`;
}

export function recentResources(workspace: string): string[] {
  try {
    const stored = JSON.parse(localStorage.getItem(storageKey(workspace)) ?? '[]');
    return Array.isArray(stored) ? stored.filter((id): id is string => typeof id === 'string') : [];
  } catch {
    return [];
  }
}

export function rememberResource(workspace: string, resourceId: string) {
  const ids = [resourceId, ...recentResources(workspace).filter(id => id !== resourceId)].slice(0, LIMIT);
  localStorage.setItem(storageKey(workspace), JSON.stringify(ids));
}
