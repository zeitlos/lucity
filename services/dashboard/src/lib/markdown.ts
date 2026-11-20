import MarkdownIt from 'markdown-it';

const markdown = new MarkdownIt({ html: false, linkify: true, breaks: true }).disable('image');

const renderLinkOpen =
  markdown.renderer.rules.link_open ?? ((tokens, index, options, _env, self) => self.renderToken(tokens, index, options));

markdown.renderer.rules.link_open = (tokens, index, options, env, self) => {
  tokens[index]!.attrSet('target', '_blank');
  tokens[index]!.attrSet('rel', 'noopener noreferrer nofollow');
  return renderLinkOpen(tokens, index, options, env, self);
};

/** Renders untrusted markdown to HTML: raw HTML is escaped, images are left as text and links open in a new tab. */
export function renderMarkdown(source: string): string {
  return markdown.render(source);
}
