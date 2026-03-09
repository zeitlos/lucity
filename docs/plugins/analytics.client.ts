export default defineNuxtPlugin(() => {
  const { analyticsScriptUrl, analyticsSiteId } = useRuntimeConfig().public;
  if (!analyticsScriptUrl || !analyticsSiteId) return;

  useHead({
    script: [{ src: analyticsScriptUrl, defer: true, 'data-site-id': analyticsSiteId }]
  });
});
