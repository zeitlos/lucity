export type Grouping = 'hour' | 'day' | 'week';

export type Period = {
  key: string;
  label: string;
  groupings: Grouping[];
  fallback: Grouping;
  start: (now: Date) => Date;
};

export type Bar = {
  title: string;
  values: number[];
  partial: boolean;
};

export type Layer = {
  label: string;
  fill: string;
  partialFill: string;
  swatch: string;
};

export type Tick = {
  text: string;
  left: string;
  transform: string;
};

const HOUR = 60 * 60 * 1000;

function daysBack(now: Date, days: number): Date {
  return new Date(now.getFullYear(), now.getMonth(), now.getDate() - days);
}

export const periods: Period[] = [
  { key: 'today', label: 'Today', groupings: ['hour'], fallback: 'hour', start: (now) => daysBack(now, 0) },
  { key: '3d', label: '3 days', groupings: ['hour', 'day'], fallback: 'hour', start: (now) => daysBack(now, 2) },
  { key: '7d', label: '7 days', groupings: ['hour', 'day'], fallback: 'day', start: (now) => daysBack(now, 6) },
  { key: '30d', label: '30 days', groupings: ['day', 'week'], fallback: 'day', start: (now) => daysBack(now, 29) },
  {
    key: '3m',
    label: '3 months',
    groupings: ['day', 'week'],
    fallback: 'week',
    start: (now) => new Date(now.getFullYear(), now.getMonth() - 3, now.getDate()),
  },
];

export const groupings: { value: Grouping; label: string }[] = [
  { value: 'hour', label: 'Hour' },
  { value: 'day', label: 'Day' },
  { value: 'week', label: 'Week' },
];

export function time(date: Date): string {
  return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
}

export function fullDate(date: Date): string {
  return date.toLocaleDateString(undefined, { weekday: 'short', day: 'numeric', month: 'short' });
}

function weekday(date: Date): string {
  return date.toLocaleDateString(undefined, { weekday: 'short', day: 'numeric' });
}

export function shortDate(date: Date): string {
  return date.toLocaleDateString(undefined, { day: 'numeric', month: 'short' });
}

export function bucketStart(at: Date, grouping: Grouping): number {
  if (grouping === 'hour') return new Date(at).setMinutes(0, 0, 0);

  const day = new Date(at.getFullYear(), at.getMonth(), at.getDate());

  if (grouping === 'day') return day.getTime();
  return new Date(day.getFullYear(), day.getMonth(), day.getDate() - ((day.getDay() + 6) % 7)).getTime();
}

function nextStart(start: number, grouping: Grouping): number {
  if (grouping === 'hour') return start + HOUR;

  const day = new Date(start);
  return new Date(day.getFullYear(), day.getMonth(), day.getDate() + (grouping === 'day' ? 1 : 7)).getTime();
}

export function bucketStarts(since: Date, end: number, grouping: Grouping): number[] {
  const starts: number[] = [];

  for (let start = bucketStart(since, grouping); start <= end; start = nextStart(start, grouping)) {
    starts.push(start);
  }

  return starts;
}

export function spanTitle(start: number, grouping: Grouping): string {
  const date = new Date(start);

  if (grouping === 'hour') return `${fullDate(date)}, ${time(date)}`;
  if (grouping === 'day') return fullDate(date);
  return `Week of ${shortDate(date)}`;
}

export function endTitle(start: number, grouping: Grouping): string {
  if (grouping === 'hour') return spanTitle(start + HOUR, grouping);
  if (grouping === 'day') return `End of ${fullDate(new Date(start))}`;
  return `End of the week of ${shortDate(new Date(start))}`;
}

function tickText(date: Date, period: Period, grouping: Grouping): string {
  if (grouping === 'hour') {
    if (period.key === 'today') return date.getHours() % 4 === 0 ? time(date) : '';
    if (date.getHours() === 0) return weekday(date);
    if (period.key === '3d' && date.getHours() === 12) return time(date);
    return '';
  }

  if (grouping === 'day') {
    if (period.key === '3d' || period.key === '7d') return weekday(date);
    if (period.key === '30d') return date.getDay() === 1 ? shortDate(date) : '';
    return date.getDate() === 1 ? date.toLocaleDateString(undefined, { month: 'short' }) : '';
  }

  return shortDate(date);
}

export function ticks(starts: number[], period: Period, grouping: Grouping): Tick[] {
  return starts.flatMap((start, index) => {
    const text = tickText(new Date(start), period, grouping);
    const left = ((index + 0.5) / starts.length) * 100;

    if (!text) return [];
    return [{ text, left: `${left}%`, transform: left < 3 ? 'translateX(-15%)' : left > 97 ? 'translateX(-85%)' : 'translateX(-50%)' }];
  });
}
