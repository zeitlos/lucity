import { useRoute } from 'vue-router';
import { useAuth } from './useAuth';

const REPO = 'zeitlos/lucity';

interface BugReportOptions {
  title?: string;
  error?: string;
}

function buildContext(extra?: { route?: string; workspace?: string }) {
  const lines = [
    '## Context',
    '',
    `- **Page**: \`${extra?.route ?? window.location.pathname}\``,
    `- **Workspace**: ${extra?.workspace || 'unknown'}`,
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
 * Standalone bug report opener for use outside Vue components (e.g. Apollo error link).
 * Does not access route or workspace context.
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
  const { activeWorkspace } = useAuth();

  function report(opts?: BugReportOptions) {
    const sections = ['## Description\n\n<!-- What happened? -->'];

    if (opts?.error) {
      sections.push(`## Error\n\n\`\`\`\n${opts.error}\n\`\`\``);
    }

    sections.push(buildContext({
      route: route.fullPath,
      workspace: activeWorkspace.value || undefined,
    }));

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
      buildContext({
        route: route.fullPath,
        workspace: activeWorkspace.value || undefined,
      }),
    ].join('\n\n');

    openGitHubIssue({
      title: '[feature] ',
      body,
      labels: 'enhancement',
    });
  }

  return { report, requestFeature };
}
