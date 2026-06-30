export const SLUG_PATTERN = /^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$/;

// Mirrors the conductor's resource_name constraint (resource_name,min=2,max=16):
// lowercase alphanumeric with hyphens, must start and end with [a-z0-9].
export const RESOURCE_NAME_PATTERN = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/;

export function deriveSlug(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 63);
}

export function isValidSlug(slug: string): boolean {
  return SLUG_PATTERN.test(slug);
}

export function deriveServiceName(input: string): string {
  return deriveSlug(input).slice(0, 16).replace(/-+$/g, '');
}

export function isValidServiceName(name: string): boolean {
  return name.length >= 2 && name.length <= 16 && RESOURCE_NAME_PATTERN.test(name);
}
