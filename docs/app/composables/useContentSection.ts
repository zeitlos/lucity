import type { ContentNavigationItem } from '@nuxt/content';

const sections = {
  docs: { collection: 'docs', label: 'Docs', home: '/docs/quickstart' },
  legal: { collection: 'legal', label: 'Legal', home: '/legal/privacy-policy' },
} as const;

/**
 * The section the current page belongs to. Each section is a content
 * collection with its own sidebar and breadcrumb root.
 */
export function useContentSection() {
  const route = useRoute();
  return computed(() => (route.path.startsWith('/legal') ? sections.legal : sections.docs));
}

/** The current section's sidebar navigation, without the collection's top-level folder. */
export async function useContentSectionNavigation() {
  const section = useContentSection();
  const { data } = await useAsyncData(
    () => `section-navigation-${section.value.collection}`,
    () => queryCollectionNavigation(section.value.collection),
    {
      transform: (items: ContentNavigationItem[]) =>
        items.find(item => item.path === `/${section.value.collection}`)?.children ?? items,
    },
  );
  return data;
}
