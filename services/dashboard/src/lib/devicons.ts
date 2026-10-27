const modules = import.meta.glob<string>('../assets/devicons/*.svg', { eager: true, query: '?url', import: 'default' });

const urls = new Map(
  Object.entries(modules).map(([path, url]) => [path.slice(path.lastIndexOf('/') + 1, -'.svg'.length), url]),
);

export function devicon(name: string): string {
  const url = urls.get(name);
  if (!url) throw new Error(`unknown devicon: ${name}`);
  return url;
}

export function findDevicon(name: string): string | undefined {
  return urls.get(name);
}
