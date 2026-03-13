import { defineNuxtModule } from '@nuxt/kit';
import { execSync } from 'node:child_process';
import { resolve, join } from 'node:path';
import { readdirSync, statSync } from 'node:fs';

function collectMarkdownFiles(dir: string, base: string): string[] {
  const files: string[] = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      files.push(...collectMarkdownFiles(full, base));
    }
    else if (entry.endsWith('.md')) {
      files.push(full.slice(base.length));
    }
  }
  return files;
}

export default defineNuxtModule({
  meta: {
    name: 'content-git-dates',
  },
  setup(_options, nuxt) {
    const contentDir = resolve(nuxt.options.rootDir, 'content');
    const dates: Record<string, string> = {};

    const files = collectMarkdownFiles(contentDir, contentDir);
    for (const relPath of files) {
      try {
        const fullPath = join(contentDir, relPath);
        const date = execSync(`git log -1 --format=%cI -- "${fullPath}"`, {
          cwd: nuxt.options.rootDir,
          encoding: 'utf-8',
          timeout: 5000,
        }).trim();

        if (date) {
          // Convert file path to route path:
          // /1.getting-started/1.quick-start.md → /getting-started/quick-start
          const routePath = relPath
            .replace(/\.md$/, '')
            .replace(/\/index$/, '')
            .replace(/\/(\d+\.)/g, '/')
            .replace(/^\/(\d+\.)/, '/');
          dates[routePath] = date;
        }
      }
      catch {
        // Git not available or file not tracked
      }
    }

    // Inject dates into public runtime config so pages can access them
    nuxt.options.runtimeConfig.public.contentDates = dates;
  },
});
