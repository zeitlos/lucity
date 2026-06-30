const DEVICON_BASE = 'https://devicons.railway.com/i/';

const ICON_SLUGS: Record<string, string> = {
  node: 'nodejs',
  nodejs: 'nodejs',
  bun: 'bun',
  deno: 'denojs',
  python: 'python',
  go: 'go',
  golang: 'go',
  php: 'php',
  ruby: 'ruby',
  rust: 'rust',
  java: 'java',
  kotlin: 'kotlin',
  elixir: 'elixir',
  dotnet: 'dotnetcore',
  csharp: 'dotnetcore',
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

export function frameworkIconUrl(framework: string, language: string): string | null {
  const key = (framework || language || '').toLowerCase();
  if (!key) return null;
  const slug = ICON_SLUGS[key] ?? key;
  return `${DEVICON_BASE}${slug}.svg`;
}
