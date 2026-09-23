interface DocPreview {
  path: string;
  title: string;
  description?: string;
}

/**
 * Titles and descriptions for every docs page, fetched once per page render
 * and shared by every link that wants to preview its target.
 */
export function useDocPreviews() {
  const { data } = useNuxtData<DocPreview[]>('doc-previews');

  const previews = useState<Map<string, DocPreview>>('doc-previews-map', () => new Map());

  async function load() {
    if (previews.value.size > 0) return;

    const pages = data.value ?? (await useAsyncData('doc-previews', () =>
      queryCollection('docs').select('path', 'title', 'description').all(),
    )).data.value;

    previews.value = new Map((pages ?? []).map(page => [page.path, page as DocPreview]));
  }

  function find(href?: string): DocPreview | undefined {
    if (!href?.startsWith('/')) return undefined;
    return previews.value.get(href.split(/[#?]/)[0]!);
  }

  return { load, find };
}
