const sites: Record<string, string> = {
  'lucity.cloud': '42cfa77ed96d'
};

export default defineNuxtPlugin(() => {
  const id = sites[window.location.hostname];
  if (!id) return;

  useHead({
    script: [{ src: 'https://p.lucity.cloud/api/script.js', defer: true, 'data-site-id': id }]
  });
});
