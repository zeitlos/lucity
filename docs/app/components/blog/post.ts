export type ProseNode = string | [string, Record<string, unknown>, ...ProseNode[]];

export interface PostImage {
  src: string;
  alt: string;
}

export interface PostSummary {
  path: string;
  title: string;
  description?: string;
  date: string;
  author?: string;
  minutes: number;
  image?: PostImage;
  quote?: string;
}

interface PostSource {
  path: string;
  title: string;
  description?: string;
  date: string;
  author?: string;
  body?: unknown;
}

const silentTags = new Set(['style', 'script']);

function isElement(node: unknown): node is [string, Record<string, unknown>, ...ProseNode[]] {
  return Array.isArray(node) && typeof node[0] === 'string';
}

export function proseNodes(body: unknown): ProseNode[] {
  const value = (body as { value?: unknown } | undefined)?.value;
  return Array.isArray(value) ? value as ProseNode[] : [];
}

function visit(nodes: ProseNode[], enter: (node: ProseNode) => boolean | undefined): boolean {
  for (const node of nodes) {
    if (enter(node)) return true;
    if (isElement(node) && !silentTags.has(node[0]) && visit(node.slice(2) as ProseNode[], enter)) return true;
  }
  return false;
}

function textOf(node: ProseNode): string {
  if (typeof node === 'string') return node;
  if (!isElement(node) || silentTags.has(node[0])) return '';
  return (node.slice(2) as ProseNode[]).map(textOf).join('');
}

function countWords(text: string) {
  return text.trim().split(/\s+/).filter(Boolean).length;
}

export function readingMinutes(body: unknown) {
  let words = 0;
  let images = 0;
  visit(proseNodes(body), (node) => {
    if (typeof node === 'string') words += countWords(node);
    else if (isElement(node) && node[0] === 'img') images += 1;
    return false;
  });
  let imageSeconds = 0;
  for (let index = 0; index < images; index++) {
    imageSeconds += Math.max(12 - index, 3);
  }
  return Math.max(1, Math.ceil(((words / 225) * 60 + imageSeconds) / 60));
}

function firstImage(nodes: ProseNode[]): PostImage | undefined {
  let image: PostImage | undefined;
  visit(nodes, (node) => {
    if (!isElement(node) || node[0] !== 'img' || typeof node[1]?.src !== 'string') return false;
    image = { src: node[1].src, alt: typeof node[1].alt === 'string' ? node[1].alt : '' };
    return true;
  });
  return image;
}

function firstQuote(nodes: ProseNode[]): string | undefined {
  let quote: string | undefined;
  visit(nodes, (node) => {
    if (!isElement(node) || node[0] !== 'blockquote') return false;
    quote = textOf(node).trim();
    return true;
  });
  return quote || undefined;
}

export function summarize(post: PostSource): PostSummary {
  const nodes = proseNodes(post.body);
  return {
    path: post.path,
    title: post.title,
    description: post.description,
    date: post.date,
    author: post.author,
    minutes: readingMinutes(post.body),
    image: firstImage(nodes),
    quote: firstQuote(nodes),
  };
}

const months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

export function formatDate(date: string) {
  const parsed = new Date(date);
  if (Number.isNaN(parsed.getTime())) return date;
  return `${parsed.getUTCDate()} ${months[parsed.getUTCMonth()]} ${parsed.getUTCFullYear()}`;
}

function isBlank(node: ProseNode) {
  return typeof node === 'string' && node.trim() === '';
}

function asFigure(node: ProseNode, eager: boolean): ProseNode {
  if (!isElement(node) || node[0] !== 'p') return node;
  const children = (node.slice(2) as ProseNode[]).filter(child => !isBlank(child));
  const image = children[0];
  const caption = children[1];
  if (!isElement(image) || image[0] !== 'img' || children.length > 2) return node;
  if (caption !== undefined && !(isElement(caption) && caption[0] === 'em')) return node;
  const picture: ProseNode = ['img', { ...image[1], title: undefined, loading: eager ? 'eager' : 'lazy', decoding: 'async' }];
  if (!caption) return ['figure', {}, picture];
  return ['figure', {}, picture, ['figcaption', {}, ...(caption.slice(2) as ProseNode[])]];
}

export function splitPost(body: unknown): { main: ProseNode[]; closing: ProseNode[] } {
  const nodes = proseNodes(body);
  let seenFigure = false;
  const shaped = nodes.map((node) => {
    const figure = asFigure(node, !seenFigure);
    if (figure !== node) seenFigure = true;
    return figure;
  });
  let rule = -1;
  shaped.forEach((node, index) => {
    if (isElement(node) && node[0] === 'hr') rule = index;
  });
  if (rule === -1) return { main: shaped, closing: [] };
  return { main: shaped.slice(0, rule), closing: shaped.slice(rule + 1) };
}

export interface Author {
  role?: string;
  avatar?: string;
}

const authors: Record<string, Author> = {
  'Christian Blättler': { role: 'Creator of Lucity', avatar: '/img/founder-avatar.png' },
};

export function authorOf(name?: string): Author {
  return (name && authors[name]) || {};
}

export function initialsOf(name?: string) {
  return (name ?? '').split(/\s+/).filter(Boolean).map(part => part[0]).slice(0, 2).join('').toUpperCase();
}
