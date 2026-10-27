import { findDevicon } from '@/lib/devicons';

const ICON_SLUGS: Record<string, string> = {
  node: 'nodejs',
  nodejs: 'nodejs',
  bun: 'bun',
  deno: 'deno',
  python: 'python',
  go: 'go',
  golang: 'go',
  php: 'php',
  ruby: 'ruby',
  rust: 'rust',
  java: 'java',
  kotlin: 'kotlin',
  elixir: 'elixir',
  dotnet: 'dot-net',
  csharp: 'dot-net',
  next: 'nextjs',
  nextjs: 'nextjs',
  nuxt: 'nuxtjs',
  nuxtjs: 'nuxtjs',
  react: 'react',
  vue: 'vuejs',
  vuejs: 'vuejs',
  angular: 'angularjs',
  svelte: 'svelte',
  astro: 'astro',
  remix: 'remix',
  vite: 'vitejs',
  django: 'django',
  flask: 'flask',
  fastapi: 'fastapi',
  rails: 'rails',
  laravel: 'laravel',
  spring: 'spring',
};

const THEMED_SLUGS = new Set(['rust', 'nextjs', 'astro', 'flask']);

export function frameworkIconUrl(framework: string, language: string, theme: 'light' | 'dark'): string | null {
  const key = (framework || language || '').toLowerCase();
  if (!key) return null;
  const slug = ICON_SLUGS[key] ?? key;
  const name = THEMED_SLUGS.has(slug) ? `${slug}-${theme === 'dark' ? 'light' : 'dark'}` : slug;
  return findDevicon(name) ?? null;
}
