import { queryCollection } from '@nuxt/content/server';

interface SitemapUrl {
  loc: string;
  lastmod?: string;
}

export default defineEventHandler(async (event) => {
  const siteUrl = 'https://lucity.cloud';
  const config = useRuntimeConfig();
  const contentDates = (config.public as unknown as Record<string, unknown>).contentDates as Record<string, string> | undefined;

  const urls: SitemapUrl[] = [
    { loc: '/', lastmod: mostRecentDate(contentDates) },
    { loc: '/blog' },
  ];

  for (const collection of ['docs', 'landing', 'blog']) {
    try {
      const pages = await queryCollection(event, collection as 'docs').all();

      for (const page of pages) {
        const meta = page as unknown as Record<string, unknown>;
        const pagePath = (page.path as string) || '/';

        if (meta.sitemap === false) continue;
        if (pagePath.endsWith('.navigation') || pagePath.includes('/.navigation')) continue;

        const urlEntry: SitemapUrl = { loc: pagePath };

        const gitDate = contentDates?.[pagePath];
        if (gitDate) {
          urlEntry.lastmod = gitDate.split('T')[0];
        }

        urls.push(urlEntry);
      }
    }
    catch {
      // Collection might not exist, skip
    }
  }

  const sitemap = generateSitemap(urls, siteUrl);
  setResponseHeader(event, 'content-type', 'application/xml');
  return sitemap;
});

function generateSitemap(urls: SitemapUrl[], siteUrl: string): string {
  const urlEntries = urls
    .map((url) => {
      const loc = `${siteUrl}${url.loc}`;
      let entry = `  <url>\n    <loc>${escapeXml(loc)}</loc>`;
      if (url.lastmod) {
        entry += `\n    <lastmod>${escapeXml(url.lastmod)}</lastmod>`;
      }
      entry += '\n  </url>';
      return entry;
    })
    .join('\n');

  return `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urlEntries}
</urlset>`;
}

function mostRecentDate(dates: Record<string, string> | undefined): string | undefined {
  if (!dates) return undefined;
  const values = Object.values(dates);
  if (values.length === 0) return undefined;
  return values.sort().pop()!.split('T')[0];
}

function escapeXml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}
