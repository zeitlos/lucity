import { coolify } from './coolify';
import { deploio } from './deploio';
import { fly } from './fly';
import { heroku } from './heroku';
import { lucityGroups } from './lucity';
import { railway } from './railway';
import { render } from './render';
import { vercel } from './vercel';
import type { CompareGroup, Comparison, ComparisonPlatform } from './types';

export { fits as comparisonFits, matrixGroups as comparisonMatrix, summaries as comparisonSummaries } from './overview';

export const comparisons: Comparison[] = [railway, heroku, render, fly, vercel, deploio, coolify];

export const comparisonPlatforms: ComparisonPlatform[] = [
  { key: 'lucity', name: 'Lucity', icon: '' },
  { key: 'railway', name: 'Railway', icon: 'i-simple-icons-railway', to: '/comparisons/vs-railway' },
  { key: 'heroku', name: 'Heroku', icon: 'i-simple-icons-heroku', to: '/comparisons/vs-heroku' },
  { key: 'render', name: 'Render', icon: 'i-simple-icons-render', to: '/comparisons/vs-render' },
  { key: 'fly', name: 'Fly.io', icon: 'i-simple-icons-flydotio', to: '/comparisons/vs-fly-io' },
  { key: 'vercel', name: 'Vercel', icon: 'i-simple-icons-vercel', to: '/comparisons/vs-vercel' },
  { key: 'deploio', name: 'Deploio', icon: '', to: '/comparisons/vs-deploio' },
  { key: 'coolify', name: 'Coolify', icon: 'i-simple-icons-coolify', to: '/comparisons/vs-coolify' },
];

export const comparisonCheckedOn = '2026-11-12';

export function findComparison(slug: string) {
  return comparisons.find(comparison => comparison.slug === slug);
}

export function comparisonGroups(comparison: Comparison): CompareGroup[] {
  return lucityGroups.map(group => ({
    title: group.title,
    rows: group.rows.flatMap((row) => {
      const other = comparison.other[row.key];
      if (!other) {
        return [];
      }
      const differs = row.lucity.mark !== other.mark || row.lucity.mark === 'info' || !!other.differs;
      return [{ key: row.key, name: row.name, lucity: row.lucity, other, differs }];
    }),
  })).filter(group => group.rows.length > 0);
}

export function formatComparisonDate(date: string) {
  return new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC' }).format(new Date(date));
}
