import { useRoute } from 'vue-router';
import { useAuth } from './useAuth';

const REPO = 'zeitlos/lucity';

interface BugReportOptions {
  title?: string;
  error?: string;
}

function buildContext(route?: string) {
  const { activeWorkspace } = useAuth();
  const lines = [
    '## Context',
    '',
    `- **Page**: \`${route ?? window.location.pathname}\``,
    `- **Workspace**: ${activeWorkspace.value || 'unknown'}`,
    `- **Browser**: ${navigator.userAgent}`,
    `- **Version**: ${typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : 'dev'}`,
    `- **Time**: ${new Date().toISOString()}`,
  ];
  return lines.join('\n');
}

function openGitHubIssue(params: { title: string; body: string; labels: string }) {
  const url = new URL(`https://github.com/${REPO}/issues/new`);
  url.searchParams.set('title', params.title);
  url.searchParams.set('body', params.body);
  url.searchParams.set('labels', params.labels);
  window.open(url.toString(), '_blank');
}

/**
 * Standalone bug report opener for use outside Vue components (e.g. toasts).
 * Reads the page from the current location instead of the router.
 */
export function openBugReport(opts?: BugReportOptions) {
  const sections = ['## Description\n\n<!-- What happened? -->'];

  if (opts?.error) {
    sections.push(`## Error\n\n\`\`\`\n${opts.error}\n\`\`\``);
  }

  sections.push(buildContext());

  openGitHubIssue({
    title: opts?.title ? `[bug] ${opts.title}` : '[bug] ',
    body: sections.join('\n\n'),
    labels: 'bug',
  });
}

export function useReportBug() {
  const route = useRoute();

  function report(opts?: BugReportOptions) {
    const sections = ['## Description\n\n<!-- What happened? -->'];

    if (opts?.error) {
      sections.push(`## Error\n\n\`\`\`\n${opts.error}\n\`\`\``);
    }

    sections.push(buildContext(route.fullPath));

    openGitHubIssue({
      title: opts?.title ? `[bug] ${opts.title}` : '[bug] ',
      body: sections.join('\n\n'),
      labels: 'bug',
    });
  }

  function requestFeature() {
    const body = [
      '## Problem\n\n<!-- What problem does this solve? -->',
      '## Proposed Solution\n\n<!-- How should it work? -->',
      buildContext(route.fullPath),
    ].join('\n\n');

    openGitHubIssue({
      title: '[feature] ',
      body,
      labels: 'enhancement',
    });
  }

  return { report, requestFeature };
}
