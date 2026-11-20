import { graphql } from '@/gql';

export const PlatformTotalsDocument = graphql(`
  query PlatformTotals {
    platformTotals {
      runningPods
      pendingPods
      failedPods
      runningContainers
      runningWorkspaces
      workspaces
      releases
      failedReleases
      firingAlerts
      runningPodsTrend
      runningContainersTrend
      runningWorkspacesTrend
      releasesTrend
      firingAlertsTrend
    }
  }
`);

export const AppChartVersionsDocument = graphql(`
  query AppChartVersions {
    appChartVersions { version releases since }
  }
`);

export const RoutinesDocument = graphql(`
  query Routines {
    routines { name hours { start passed failed } averageDuration averageDurationChange lastPassedAt }
  }
`);

export const LimitsDocument = graphql(`
  query Limits {
    limits { name limits { name unit value cap history { time value } } }
  }
`);

export const NodePoolsDocument = graphql(`
  query NodePools {
    nodePools {
      name
      nodes {
        name
        instanceType
        createdAt
        attachedVolumes
        volumeLimit
        pods
        podLimit
        memoryAllocatable
        memoryRequested
        memoryUsed
        cpuAllocatable
        cpuRequested
        cpuUsed
      }
    }
  }
`);

export const PlatformLogsDocument = graphql(`
  query PlatformLogs {
    platformLogs {
      hours { start errors warnings }
      errors24h
      errors7d
      warnings24h
      warnings7d
      messages { level message service count24h count7d firstSeenAt }
    }
  }
`);

const integer = new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 });
const decimal = new Intl.NumberFormat(undefined, { minimumFractionDigits: 1, maximumFractionDigits: 1 });

export function count(value: number): string {
  return integer.format(value);
}

export function oneDecimal(value: number): string {
  return decimal.format(value);
}

export function percent(ratio: number): string {
  return `${Math.round(ratio * 100)}%`;
}

export function change(ratio: number): string {
  const rounded = Math.round(ratio * 100);

  if (rounded === 0) return '0%';
  return rounded > 0 ? `+${rounded}%` : `−${-rounded}%`;
}

export function hour(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' });
}

export function shortDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, { day: 'numeric', month: 'short' });
}

export function ago(iso: string): string {
  const seconds = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 1000));

  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;

  return shortDate(iso);
}

export function seconds(value: number): string {
  if (value < 0.001) return '<1 ms';
  if (value < 0.1) return `${Math.round(value * 1000)} ms`;
  if (value < 60) return `${oneDecimal(value)}s`;

  return `${Math.floor(value / 60)}m ${Math.round(value % 60)}s`;
}

export const GiB = 2 ** 30;

export function gibibytes(bytes: number): string {
  const value = bytes / GiB;
  return value >= 100 ? count(value) : oneDecimal(value);
}

export type Tone = 'none' | 'warn' | 'danger';

export function tone(ratio: number, warn: number): Tone {
  if (ratio >= 1) return 'danger';
  if (ratio >= warn) return 'warn';
  return 'none';
}

export const toneText: Record<Tone, string> = {
  none: 'text-foreground',
  warn: 'text-status-warn',
  danger: 'text-status-danger',
};

/** Builds the line and area paths of a chart width × height wide from points
 * spread over [start, end], scaled to top. */
export function chartPaths(points: { at: number; value: number }[], start: number, end: number, top: number, width: number, height: number) {
  if (points.length === 0 || top <= 0) return { line: '', area: '' };

  const x = (at: number) => ((at - start) / (end - start || 1)) * width;
  const y = (value: number) => height - 1 - ((height - 4) * Math.min(value, top)) / top;
  const coordinates = points.map((point) => `${x(point.at).toFixed(1)} ${y(point.value).toFixed(1)}`);
  const line = `M${coordinates.join('L')}`;

  return {
    line,
    area: `${line}L${x(points[points.length - 1]!.at).toFixed(1)} ${height}L${x(points[0]!.at).toFixed(1)} ${height}Z`,
  };
}

/** Builds a sparkline for evenly spaced values, scaled between their minimum
 * and maximum. */
export function sparkline(values: number[], width: number, height: number) {
  if (values.length < 2) return { line: '', area: '' };

  const low = Math.min(...values);
  const span = Math.max(...values) - low || 1;
  const coordinates = values.map((value, index) => `${((index * width) / (values.length - 1)).toFixed(1)} ${(height - 3 - ((value - low) / span) * (height - 8)).toFixed(1)}`);
  const line = `M${coordinates.join('L')}`;

  return { line, area: `${line}L${width} ${height}L0 ${height}Z` };
}
