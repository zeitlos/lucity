import { graphql } from '@/gql';
import type { Workspace } from '@/lib/workspaces';

export const UserSessionsDocument = graphql(`
  query UserSessions($userID: ID!) {
    userSessions(userID: $userID) {
      url
      sessions {
        id
        startedAt
        duration
        entryPage
        pageviews
        country
        city
        language
        browser
        operatingSystem
        referrer
        hasReplay
      }
    }
  }
`);

export const SessionReplayDocument = graphql(`
  query SessionReplay($sessionID: ID!) {
    sessionReplay(sessionID: $sessionID)
  }
`);

type Member = Workspace['members'][number];

export type User = Omit<Member, 'role'> & {
  memberships: { workspace: Workspace; role: string | null | undefined }[];
};

export function findUser(workspaces: Workspace[], id: string): User | null {
  let user: User | null = null;

  for (const workspace of workspaces) {
    const member = workspace.members.find((candidate) => candidate.id === id);

    if (!member) continue;

    user ??= { ...member, memberships: [] };
    user.memberships.push({ workspace, role: member.role });
  }

  return user;
}

export function displayName(user: Pick<Member, 'name' | 'email' | 'account'>): string {
  return user.name || user.account?.login || user.email || 'Unknown user';
}

const browserFiles: Record<string, string> = {
  Chrome: 'Chrome',
  'Mobile Chrome': 'Chrome',
  'Chrome WebView': 'Chrome',
  'Chrome Headless': 'Chromium',
  Chromium: 'Chromium',
  Firefox: 'Firefox',
  'Mobile Firefox': 'Firefox',
  'Firefox Focus': 'Firefox',
  Safari: 'Safari',
  'Mobile Safari': 'Safari',
  Edge: 'Edge',
  Opera: 'Opera',
  'Samsung Internet': 'SamsungInternet',
  Brave: 'Brave',
  'Android Browser': 'Android',
};

const systemFiles: Record<string, string> = {
  Windows: 'Windows',
  macOS: 'macOS',
  iOS: 'Apple',
  Android: 'Android',
  Linux: 'Tux',
  Ubuntu: 'Ubuntu',
};

export function browserIcon(browser: string): string | null {
  const file = browserFiles[browser];
  return file ? new URL(`../assets/browsers/${file}.svg`, import.meta.url).href : null;
}

export function systemIcon(system: string): string | null {
  const file = systemFiles[system];
  return file ? new URL(`../assets/operating-systems/${file}.svg`, import.meta.url).href : null;
}

export function flag(country: string): string {
  if (!/^[A-Z]{2}$/.test(country)) return '';
  return String.fromCodePoint(...[...country].map((letter) => 0x1f1e6 + letter.charCodeAt(0) - 65));
}

const regions = new Intl.DisplayNames(undefined, { type: 'region' });

export function countryName(country: string): string {
  if (!/^[A-Z]{2}$/.test(country)) return country;
  return regions.of(country) ?? country;
}

export function referrerHost(referrer: string): string {
  try {
    return new URL(referrer).hostname.replace(/^www\./, '');
  } catch {
    return referrer;
  }
}

export function pagePath(path: string): string {
  try {
    return decodeURIComponent(path) || '/';
  } catch {
    return path || '/';
  }
}

export function sessionLength(seconds: number): string {
  const minutes = Math.round(seconds / 60);

  if (minutes < 1) return '<1m';
  if (minutes < 60) return `${minutes}m`;

  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

export function dayLabel(date: Date): string {
  const today = new Date();
  const yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1);

  if (date.toDateString() === today.toDateString()) return 'Today';
  if (date.toDateString() === yesterday.toDateString()) return 'Yesterday';

  return date.toLocaleDateString(undefined, {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: date.getFullYear() === today.getFullYear() ? undefined : 'numeric',
  });
}

export function clockTime(date: Date): string {
  return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
}
