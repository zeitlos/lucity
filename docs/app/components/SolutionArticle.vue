<script setup lang="ts">
import { kebabCase } from 'scule';

const props = defineProps<{
  /** Section this page belongs to. */
  section: string;
  /** Section root, used for sibling lookup. */
  sectionTo: string;
  /** Whether the section has an index page worth linking back to. */
  hasIndex?: boolean;
}>();

const route = useRoute();

const { data: page } = await useAsyncData(kebabCase(route.path), () =>
  queryCollection('solutions').path(route.path).first(),
);

if (!page.value) {
  throw createError({ statusCode: 404, statusMessage: 'Page not found', fatal: true });
}

useSeo({
  title: page.value.title,
  description: page.value.description,
  type: 'article',
});

const { data: siblings } = await useAsyncData(`${kebabCase(route.path)}-siblings`, () =>
  queryCollection('solutions')
    .where('path', 'LIKE', `${props.sectionTo}/%`)
    .select('path', 'title', 'description')
    .all(),
);

const others = computed(() =>
  (siblings.value ?? []).filter(item => item.path !== route.path).slice(0, 2),
);
</script>

<template>
  <div
    v-if="page"
    class="solution"
  >
    <UContainer>
      <article class="solution-inner">
        <NuxtLink
          v-if="hasIndex && route.path !== sectionTo"
          :to="sectionTo"
          class="solution-back"
        >
          <UIcon
            name="i-lucide-arrow-left"
            class="size-4"
          />
          {{ section }}
        </NuxtLink>

        <header class="solution-header">
          <p class="solution-eyebrow">{{ section }}</p>
          <h1>{{ page.title }}</h1>
          <p
            v-if="page.description"
            class="solution-lede"
          >
            {{ page.description }}
          </p>

          <div
            v-if="page.actions?.length"
            class="solution-actions"
          >
            <UButton
              v-for="(action, index) in page.actions"
              :key="action.label"
              :to="action.to"
              :target="action.to?.startsWith('http') ? '_blank' : undefined"
              :color="index === 0 ? 'primary' : 'neutral'"
              :variant="index === 0 ? 'solid' : 'outline'"
              size="lg"
              :label="action.label"
              :trailing-icon="index === 0 ? 'i-lucide-arrow-right' : undefined"
            />
          </div>
        </header>

        <div class="solution-body">
          <ContentRenderer :value="page" />
        </div>

        <footer
          v-if="others.length"
          class="solution-footer"
        >
          <USeparator />
          <div class="solution-others">
            <NuxtLink
              v-for="other in others"
              :key="other.path"
              :to="other.path"
              class="solution-other"
            >
              <p class="solution-other-title">{{ other.title }}</p>
              <p
                v-if="other.description"
                class="solution-other-text"
              >
                {{ other.description }}
              </p>
              <span class="solution-other-cta">
                Read on
                <UIcon
                  name="i-lucide-arrow-right"
                  class="size-4"
                />
              </span>
            </NuxtLink>
          </div>
        </footer>
      </article>
    </UContainer>

    <FooterCta class="solution-cta" />
  </div>
</template>

<style scoped>
.solution {
  padding-top: 3rem;
  padding-bottom: 4rem;
}

.solution-inner {
  max-width: 720px;
  margin: 0 auto;
}

.solution-back {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.875rem;
  color: var(--ui-text-muted);
  text-decoration: none;
  transition: color 0.15s;
}

.solution-back:hover {
  color: var(--ui-text-highlighted);
}

.solution-header {
  margin-top: 2rem;
  margin-bottom: 2.5rem;
}

.solution-eyebrow {
  margin: 0 0 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--ui-primary);
}

.solution-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 2rem;
}

.solution-header h1 {
  font-family: var(--font-display);
  font-size: 2.5rem;
  font-weight: normal;
  line-height: 1.15;
  color: var(--ui-text-highlighted);
  margin: 0;
}

@media (min-width: 640px) {
  .solution-header h1 {
    font-size: 3.25rem;
  }
}

.solution-lede {
  font-size: 1.25rem;
  line-height: 1.6;
  color: var(--ui-text-muted);
  margin-top: 1rem;
}

.solution-body {
  font-size: 1.0625rem;
  line-height: 1.75;
}

.solution-body :deep(h2) {
  font-family: var(--font-display);
  font-size: 2rem;
  font-weight: normal;
  margin-top: 3rem;
  color: var(--ui-text-highlighted);
}

.solution-body :deep(h3) {
  font-size: 1.25rem;
  font-weight: 600;
  margin-top: 2rem;
  color: var(--ui-text-highlighted);
}

.solution-body :deep(li) {
  margin-top: 0.5rem;
}

/* Heading anchor links inherit the heading colour, not the link colour. */
.solution-body :deep(h2 a),
.solution-body :deep(h3 a) {
  color: inherit;
  text-decoration: none;
}

.solution-body :deep(table) {
  width: 100%;
  margin-top: 1.5rem;
  font-size: 0.9375rem;
  border-collapse: collapse;
}

.solution-body :deep(th),
.solution-body :deep(td) {
  border: 1px solid var(--ui-border);
  padding: 0.5rem 0.75rem;
  text-align: start;
}

.solution-cta {
  margin-top: 6rem;
}

.solution-footer {
  margin-top: 4rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.solution-others {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}

@media (min-width: 640px) {
  .solution-others {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.solution-other {
  display: flex;
  flex-direction: column;
  padding: 1.25rem;
  border: 1px solid var(--ui-border);
  border-radius: 0.875rem;
  text-decoration: none;
  transition: border-color 0.15s, transform 0.15s;
}

.solution-other:hover {
  border-color: color-mix(in oklab, var(--ui-primary) 45%, transparent);
  transform: translateY(-2px);
}

.solution-other-title {
  margin: 0;
  font-weight: 600;
  color: var(--ui-text-highlighted);
}

.solution-other-text {
  margin: 0.375rem 0 0;
  font-size: 0.9375rem;
  line-height: 1.55;
  color: var(--ui-text-muted);
}

.solution-other-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin-top: 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--ui-primary);
}
</style>

<style>
/* Direct children only: embedded components bring their own spacing. */
.solution-body > p,
.solution-body > ul,
.solution-body > ol {
  margin-top: 1.25rem;
}
</style>
