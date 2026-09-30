export default defineNitroPlugin((nitroApp) => {
  const contentDates = (useRuntimeConfig().public as unknown as Record<string, unknown>).contentDates as Record<string, string> | undefined;

  nitroApp.hooks.hook('sitemap:resolved', (ctx) => {
    for (const url of ctx.urls) {
      const date = contentDates?.[new URL(url.loc, 'https://lucity.cloud').pathname];
      if (date) url.lastmod = date.split('T')[0];
    }
  });
});
