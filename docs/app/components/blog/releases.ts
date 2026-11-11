import shipped from '~/assets/landing/releases.json';

export interface Release {
  date: string;
  tag?: string;
  items: { name: string; text: string }[];
}

export const releasesUrl = 'https://github.com/zeitlos/lucity/releases';

const tags: Record<string, string> = {
  '11 Nov': 'v26.11.3',
  '6 Nov': 'v26.11.2',
  '2 Nov': 'v26.11.1',
};

const earlier: Release[] = [
  {
    date: '2 Oct',
    tag: 'v26.10.1',
    items: [
      { name: 'A global edge in front of every service', text: 'Visitors connect to a location near them, with nothing to change in DNS.' },
    ],
  },
  {
    date: '29 Sep',
    tag: 'v26.9.2',
    items: [
      { name: 'New docs, part of this site', text: 'A quickstart with screenshots, framework guides and a page per building block.' },
      { name: 'Secrets stay out of your agent\'s context', text: 'The MCP server no longer returns variable values or credentials, and masks secrets in logs.' },
      { name: 'Eject comes with a build script', text: 'An ejected project ships with a README and a script that rebuilds your services.' },
    ],
  },
  {
    date: '21 Sep',
    tag: 'v26.9.1',
    items: [
      { name: 'A vulnerability report for every service', text: 'For the exact image that is running, refreshed as new advisories are published.' },
      { name: 'Postgres backups with point-in-time recovery', text: 'Restore to any moment of the last 30 days, into a second database.' },
      { name: 'Redirects for custom domains', text: 'Send example.com to www.example.com, path and query included.' },
    ],
  },
];

export const releases: Release[] = [
  ...shipped.map(release => ({ ...release, tag: tags[release.date] })),
  ...earlier.filter(release => !shipped.some(recent => recent.date === release.date)),
];
