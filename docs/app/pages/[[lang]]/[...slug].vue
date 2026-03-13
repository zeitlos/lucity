<script setup lang="ts">
import { kebabCase } from 'scule';
import type { ContentNavigationItem, Collections, DocsCollectionItem } from '@nuxt/content';
import { findPageHeadline } from '@nuxt/content/utils';

definePageMeta({
  layout: 'docs',
});

const route = useRoute();
const { locale, isEnabled, t } = useDocusI18n();
const appConfig = useAppConfig();
const navigation = inject<Ref<ContentNavigationItem[]>>('navigation');
const { shouldPushContent: shouldHideToc } = useAssistant();

const collectionName = computed(() => isEnabled.value ? `docs_${locale.value}` : 'docs');

const [{ data: page }, { data: surround }] = await Promise.all([
  useAsyncData(kebabCase(route.path), () => queryCollection(collectionName.value as keyof Collections).path(route.path).first() as Promise<DocsCollectionItem>),
  useAsyncData(`${kebabCase(route.path)}-surround`, () => {
    return queryCollectionItemSurroundings(collectionName.value as keyof Collections, route.path, {
      fields: ['description'],
    });
  }),
]);

const runtimeConfig = useRuntimeConfig();
const contentDates = (runtimeConfig.public as unknown as Record<string, unknown>).contentDates as Record<string, string> | undefined;

if (!page.value) {
  throw createError({ statusCode: 404, statusMessage: 'Page not found', fatal: true });
}

const title = page.value.seo?.title || page.value.title;
const description = page.value.seo?.description || page.value.description;
const modifiedAt = computed(() => contentDates?.[route.path] || null);

const headline = ref(findPageHeadline(navigation?.value, page.value?.path));
const breadcrumbs = computed(() => findPageBreadcrumbs(navigation?.value, page.value?.path || ''));

useSeo({
  title,
  description,
  type: 'article',
  modifiedAt,
  breadcrumbs,
});
watch(() => navigation?.value, () => {
  headline.value = findPageHeadline(navigation?.value, page.value?.path) || headline.value;
});

defineOgImageComponent('Docs', {
  headline: headline.value,
});

const github = computed(() => appConfig.github ? appConfig.github : null);

const editLink = computed(() => {
  if (!github.value) {
    return;
  }

  return [
    github.value.url,
    'edit',
    github.value.branch,
    github.value.rootDir,
    'content',
    `${page.value?.stem}.${page.value?.extension}`,
  ].filter(Boolean).join('/');
});

const formattedDate = computed(() => {
  if (!modifiedAt.value) return null;
  const date = new Date(modifiedAt.value);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
});

// Add the page path to the prerender list
addPrerenderPath(`/raw${route.path}.md`);
</script>

<template>
  <UPage
    v-if="page"
    :key="`page-${shouldHideToc}`"
  >
    <UPageHeader
      :title="page.title"
      :description="page.description"
      :headline="headline"
      :ui="{
        wrapper: 'flex-row items-center flex-wrap justify-between',
      }"
    >
      <template #links>
        <UButton
          v-for="(link, index) in (page as DocsCollectionItem).links"
          :key="index"
          size="sm"
          v-bind="link"
        />

        <DocsPageHeaderLinks />
      </template>
    </UPageHeader>

    <UPageBody>
      <ContentRenderer
        v-if="page"
        :value="page"
      />

      <USeparator v-if="github || formattedDate">
        <div
          class="flex items-center gap-2 text-sm text-muted"
        >
          <span
            v-if="formattedDate"
            class="flex items-center gap-1"
          >
            <UIcon
              name="i-lucide-calendar"
              class="size-4"
            />
            Last updated {{ formattedDate }}
          </span>
          <template v-if="github && formattedDate">
            <span>&middot;</span>
          </template>
          <UButton
            v-if="github"
            variant="link"
            color="neutral"
            :to="editLink"
            target="_blank"
            icon="i-lucide-pen"
            :ui="{ leadingIcon: 'size-4' }"
          >
            {{ t('docs.edit') }}
          </UButton>
          <span v-if="github">{{ t('common.or') }}</span>
          <UButton
            v-if="github"
            variant="link"
            color="neutral"
            :to="`${github.url}/issues/new/choose`"
            target="_blank"
            icon="i-lucide-alert-circle"
            :ui="{ leadingIcon: 'size-4' }"
          >
            {{ t('docs.report') }}
          </UButton>
        </div>
      </USeparator>
      <UContentSurround :surround="surround" />
    </UPageBody>

    <template
      v-if="page?.body?.toc?.links?.length && !shouldHideToc"
      #right
    >
      <UContentToc
        highlight
        :title="appConfig.toc?.title || t('docs.toc')"
        :links="page.body?.toc?.links"
      >
        <template #bottom>
          <DocsAsideRightBottom />
        </template>
      </UContentToc>
    </template>
  </UPage>
</template>
