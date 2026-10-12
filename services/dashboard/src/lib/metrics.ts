import { MetricWindow } from '@/gql/graphql';
import { formatBytes } from '@/lib/utils';

export type MetricPoint = { timestamp: string; value?: number | null };

export type MetricLine = { key: string; label: string; color: string; points: MetricPoint[] };

export type MetricReading = { key: string; color: string; text: string };

export type MetricMarker = { time: number; label: string };

export const metricWindows: { value: MetricWindow; label: string; ms: number }[] = [
  { value: MetricWindow.Last_1H, label: '1h', ms: 3_600_000 },
  { value: MetricWindow.Last_6H, label: '6h', ms: 6 * 3_600_000 },
  { value: MetricWindow.Last_24H, label: '24h', ms: 24 * 3_600_000 },
  { value: MetricWindow.Last_7D, label: '7d', ms: 7 * 24 * 3_600_000 },
  { value: MetricWindow.Last_30D, label: '30d', ms: 30 * 24 * 3_600_000 },
];

export const lineColors = ['var(--chart-1)', 'var(--chart-5)', 'var(--chart-2)', 'var(--chart-4)', 'var(--chart-3)'];

export function windowOf(value: MetricWindow) {
  return metricWindows.find(w => w.value === value) ?? metricWindows[2]!;
}

export function timeOf(point: MetricPoint): number {
  return new Date(point.timestamp).getTime();
}

export function latestValue(points: MetricPoint[]): number | null {
  for (let i = points.length - 1; i >= 0; i--) {
    const value = points[i]!.value;
    if (value != null) return value;
  }
  return null;
}

export function valueAt(points: MetricPoint[], time: number): number | null {
  return points.find(p => timeOf(p) === time)?.value ?? null;
}

export function nearestTime(points: MetricPoint[], cursor: number): number | null {
  let best: number | null = null;
  for (const point of points) {
    const time = timeOf(point);
    if (best == null || Math.abs(time - cursor) < Math.abs(best - cursor)) best = time;
  }
  return best;
}

export function formatMetricTime(time: number, short: boolean): string {
  return new Date(time).toLocaleString(
    undefined,
    short
      ? { hour: 'numeric', minute: '2-digit' }
      : { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' },
  );
}

export function metricTicks(from: number, to: number, short: boolean): string[] {
  return [0, 0.25, 0.5, 0.75, 1].map(fraction => {
    if (fraction === 1) return 'now';
    const time = from + (to - from) * fraction;
    return new Date(time).toLocaleString(
      undefined,
      short ? { hour: 'numeric', minute: '2-digit' } : { month: 'short', day: 'numeric' },
    );
  });
}

export function formatCores(value: number): string {
  if (value < 0.0005) return '0';
  return value < 0.1 ? value.toFixed(3) : value.toFixed(2);
}

export function formatCoreLimit(value: number): string {
  const cores = Number(value.toFixed(2));
  return `${cores} ${cores === 1 ? 'core' : 'cores'}`;
}

export function formatSize(value: number): string {
  return formatBytes(value).replace('.0 ', ' ');
}

export function formatCount(value: number): string {
  return Math.round(value).toLocaleString();
}

export function formatRate(value: number): string {
  if (value >= 1000) return `${Number((value / 1000).toFixed(1))}k`;
  if (value >= 10) return String(Math.round(value));
  return String(Number(value.toFixed(1)));
}

export function formatRatio(value: number): string {
  return `${(value * 100).toFixed(1)}%`;
}

export function formatSeconds(value: number): string {
  if (value < 1) return `${Math.round(value * 1000)} ms`;
  if (value < 60) return `${Number(value.toFixed(1))} s`;
  return `${Number((value / 60).toFixed(1))} min`;
}

export function formatMilliseconds(value: number): string {
  if (value < 1) return `${Number(value.toFixed(2))} ms`;
  if (value < 1000) return `${Math.round(value)} ms`;
  return formatSeconds(value / 1000);
}
