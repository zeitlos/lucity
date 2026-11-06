export const ID_MAX_LENGTH = 16;

const ID_PATTERN = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;

export function isValidId(id: string): boolean {
  return id.length >= 2 && id.length <= ID_MAX_LENGTH && ID_PATTERN.test(id);
}

function slugify(name: string): string {
  return name
    .toLowerCase()
    .normalize('NFKD')
    .replace(/\p{Mn}/gu, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

export function sanitizeId(value: string): string {
  return value
    .toLowerCase()
    .normalize('NFKD')
    .replace(/\p{Mn}/gu, '')
    .replace(/[\s_.]+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/^-+/, '')
    .slice(0, ID_MAX_LENGTH);
}

function cut(value: string, length: number): string {
  return value.slice(0, length).replace(/-+$/, '');
}

export function deriveId(name: string, fallback: string, taken: ReadonlySet<string> = new Set()): string {
  const slug = slugify(name);
  const base = slug.length >= 2 ? slug : fallback;
  let id = cut(base, ID_MAX_LENGTH);

  for (let n = 2; taken.has(id); n++) {
    const suffix = `-${n}`;
    id = cut(base, ID_MAX_LENGTH - suffix.length) + suffix;
  }

  return id;
}

export function lastSegment(id: string): string {
  return id.slice(id.lastIndexOf('/') + 1);
}
