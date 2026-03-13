import { defineNuxtModule } from '@nuxt/kit';
import { execSync } from 'node:child_process';
import { resolve, join } from 'node:path';
import { readdirSync, statSync, readFileSync, existsSync } from 'node:fs';

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

function filePathToRoutePath(relPath: string): string {
  return relPath
    .replace(/\.md$/, '')
    .replace(/\/index$/, '')
    .replace(/\/(\d+\.)/g, '/')
    .replace(/^\/(\d+\.)/, '/');
}

export default defineNuxtModule({
  meta: {
    name: 'content-git-dates',
  },
  setup(_options, nuxt) {
    const rootDir = nuxt.options.rootDir;
    const contentDir = resolve(rootDir, 'content');
    const preGenerated = resolve(rootDir, '_content-dates.json');

    // If a pre-generated file exists (created by CI before Docker build), use it
    if (existsSync(preGenerated)) {
      const dates = JSON.parse(readFileSync(preGenerated, 'utf-8'));
      nuxt.options.runtimeConfig.public.contentDates = dates;
      return;
    }

    // Otherwise, generate dates from git history (local dev)
    const dates: Record<string, string> = {};
    const files = collectMarkdownFiles(contentDir, contentDir);

    for (const relPath of files) {
      try {
        const fullPath = join(contentDir, relPath);
        const date = execSync(`git log -1 --format=%cI -- "${fullPath}"`, {
          cwd: rootDir,
          encoding: 'utf-8',
          timeout: 5000,
        }).trim();

        if (date) {
          dates[filePathToRoutePath(relPath)] = date;
        }
      }
      catch {
        // Git not available or file not tracked
      }
    }

    nuxt.options.runtimeConfig.public.contentDates = dates;
  },
});
