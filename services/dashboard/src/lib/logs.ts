import { ServiceLogEvent } from '@/gql/graphql';
import { formatDateTime, formatTime } from '@/lib/format';

export type LogRange = 'live' | '1h' | '24h' | '7d' | 'custom';

export const LIVE_WINDOW = 15 * 60_000;

export const rangeSpans: Record<Exclude<LogRange, 'live' | 'custom'>, number> = {
  '1h': 60 * 60_000,
  '24h': 24 * 60 * 60_000,
  '7d': 7 * 24 * 60 * 60_000,
};

export const instanceColors = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-4)', 'var(--chart-5)', 'var(--chart-3)'];

export function shortPod(name: string): string {
  return name.split('-').pop() ?? name;
}

function pad(value: number, length = 2): string {
  return String(value).padStart(length, '0');
}

export function logClock(value: string): { time: string; fraction: string } {
  const date = new Date(value);
  return {
    time: `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`,
    fraction: `.${pad(date.getMilliseconds(), 3)}`,
  };
}

export function formatRangeEdge(at: number, other: number): string {
  const sameDay = new Date(at).toDateString() === new Date(other).toDateString();
  const today = new Date(at).toDateString() === new Date().toDateString();
  return sameDay && today ? formatTime(at) : formatDateTime(at);
}

const tickSteps = [
  60_000, 2 * 60_000, 5 * 60_000, 10 * 60_000, 15 * 60_000, 30 * 60_000,
  3_600_000, 2 * 3_600_000, 3 * 3_600_000, 6 * 3_600_000, 12 * 3_600_000, 86_400_000,
];

export function axisTicks(from: number, to: number): { at: number; label: string }[] {
  const span = to - from;
  const step = tickSteps.find(candidate => span / candidate <= 5) ?? tickSteps[tickSteps.length - 1]!;
  const offset = new Date(from).getTimezoneOffset() * 60_000;
  const ticks: { at: number; label: string }[] = [];
  for (let time = Math.ceil((from - offset) / step) * step + offset; time <= to; time += step) {
    const label = step >= 86_400_000
      ? new Date(time).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
      : formatTime(time);
    ticks.push({ at: (time - from) / span, label });
  }
  return ticks;
}

export function eventText(event: ServiceLogEvent, line: string): string {
  switch (event) {
    case ServiceLogEvent.Started:
      return 'Started';
    case ServiceLogEvent.Restarted:
      return 'Exited and restarted';
    case ServiceLogEvent.Stopping:
      return 'Stopping';
    case ServiceLogEvent.HealthCheckFailed:
      return line ? `Health check failed: ${line}` : 'Health check failed';
    case ServiceLogEvent.RestartDelayed:
      return 'Exited again, waiting before the next restart';
  }
}
