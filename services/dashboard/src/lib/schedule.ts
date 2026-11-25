import { CronExpressionParser } from 'cron-parser';
import cronstrue from 'cronstrue';

export type Repeats = 'hourly' | 'daily' | 'weekly' | 'monthly' | 'custom';

export interface Schedule {
  repeats: Repeats;
  minute: number;
  time: string;
  days: number[];
  dayOfMonth: number;
  cron: string;
}

export const repeatOptions: { value: Repeats; label: string }[] = [
  { value: 'hourly', label: 'Hourly' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
  { value: 'custom', label: 'Custom' },
];

export const weekdays: { value: number; label: string; name: string }[] = [
  { value: 1, label: 'Mo', name: 'Monday' },
  { value: 2, label: 'Tu', name: 'Tuesday' },
  { value: 3, label: 'We', name: 'Wednesday' },
  { value: 4, label: 'Th', name: 'Thursday' },
  { value: 5, label: 'Fr', name: 'Friday' },
  { value: 6, label: 'Sa', name: 'Saturday' },
  { value: 0, label: 'Su', name: 'Sunday' },
];

export function defaultSchedule(): Schedule {
  return { repeats: 'weekly', minute: 0, time: '09:00', days: [1, 2, 3, 4, 5], dayOfMonth: 1, cron: '' };
}

export function toCron(schedule: Schedule): string {
  const [hour = '0', minute = '0'] = schedule.time.split(':');
  const at = `${Number(minute)} ${Number(hour)}`;

  switch (schedule.repeats) {
    case 'hourly':
      return `${schedule.minute} * * * *`;
    case 'daily':
      return `${at} * * *`;
    case 'weekly':
      return `${at} * * ${dayList(schedule.days)}`;
    case 'monthly':
      return `${at} ${schedule.dayOfMonth} * *`;
    default:
      return schedule.cron.trim();
  }
}

function dayList(days: number[]): string {
  const sorted = [...new Set(days)].sort((a, b) => a - b);
  const parts: string[] = [];

  for (let i = 0; i < sorted.length; i++) {
    let end = i;
    while (end + 1 < sorted.length && sorted[end + 1] === sorted[end]! + 1) end++;

    if (end - i >= 2) {
      parts.push(`${sorted[i]}-${sorted[end]}`);
      i = end;
    } else {
      parts.push(`${sorted[i]}`);
    }
  }

  return parts.join(',');
}

function parseDays(list: string): number[] {
  return list.split(',').flatMap((part) => {
    const [start, end = start] = part.split('-').map(Number);
    return Array.from({ length: end! - start! + 1 }, (_, offset) => start! + offset);
  });
}

const number = '(\\d{1,2})';
const hourlyRe = new RegExp(`^${number} \\* \\* \\* \\*$`);
const dailyRe = new RegExp(`^${number} ${number} \\* \\* \\*$`);
const weeklyRe = new RegExp(`^${number} ${number} \\* \\* ([0-6](?:-[0-6])?(?:,[0-6](?:-[0-6])?)*)$`);
const monthlyRe = new RegExp(`^${number} ${number} ${number} \\* \\*$`);

function time(hour: string, minute: string): string {
  return `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;
}

export function fromCron(cron: string): Schedule {
  const schedule = { ...defaultSchedule(), cron };
  let match: RegExpMatchArray | null;

  if ((match = cron.match(hourlyRe))) return { ...schedule, repeats: 'hourly', minute: Number(match[1]) };
  if ((match = cron.match(dailyRe))) return { ...schedule, repeats: 'daily', time: time(match[2]!, match[1]!) };
  if ((match = cron.match(weeklyRe))) {
    return { ...schedule, repeats: 'weekly', time: time(match[2]!, match[1]!), days: parseDays(match[3]!) };
  }
  if ((match = cron.match(monthlyRe))) {
    return { ...schedule, repeats: 'monthly', time: time(match[2]!, match[1]!), dayOfMonth: Number(match[3]) };
  }

  return { ...schedule, repeats: 'custom' };
}

export function describeCron(cron: string): string | null {
  try {
    return cronstrue.toString(cron, { use24HourTimeFormat: true, throwExceptionOnParseError: true });
  } catch {
    return null;
  }
}

export function nextRuns(cron: string, timeZone: string, count: number): Date[] {
  try {
    return CronExpressionParser.parse(cron, { tz: timeZone }).take(count).map(run => run.toDate());
  } catch {
    return [];
  }
}
