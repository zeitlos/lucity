const LOCALE = 'en-US';

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

function trim(value: number, digits: number): string {
  return Number(value.toFixed(digits)).toLocaleString(LOCALE, { maximumFractionDigits: digits });
}

export function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) return '0 B';
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), BYTE_UNITS.length - 1);
  const value = bytes / Math.pow(1024, index);
  return `${trim(value, index === 0 || value >= 100 ? 0 : 1)} ${BYTE_UNITS[index]}`;
}

export function parseQuantity(quantity: string): number {
  const match = /^(\d+(?:\.\d+)?)\s*([KMGTPE]i?|m)?$/.exec(quantity.trim());
  if (!match) return 0;
  const value = parseFloat(match[1]!);
  const unit = match[2];
  if (!unit) return value;
  if (unit === 'm') return value / 1000;
  const exponents: Record<string, number> = { K: 1, M: 2, G: 3, T: 4, P: 5, E: 6 };
  const base = unit.endsWith('i') ? 1024 : 1000;
  return value * Math.pow(base, exponents[unit[0]!] ?? 0);
}

export function formatStorage(quantity: string): string {
  return formatBytes(parseQuantity(quantity));
}

export function formatCpu(cores: number): string {
  if (cores > 0 && cores < 0.01) return '<0.01 vCPU';
  return `${trim(cores, 2)} vCPU`;
}

export function formatCpuQuantity(quantity: string): string {
  return formatCpu(parseQuantity(quantity));
}

export function formatCount(value: number): string {
  return Math.round(value).toLocaleString(LOCALE);
}

export function formatCompact(value: number): string {
  if (value >= 999_500) return `${trim(value / 1_000_000, 1)}M`;
  if (value >= 100_000) return `${Math.round(value / 1000)}k`;
  if (value >= 1000) return `${trim(value / 1000, 1)}k`;
  if (value >= 10) return String(Math.round(value));
  return trim(value, 1);
}

export function formatPercent(ratio: number): string {
  return `${trim(ratio * 100, 1)}%`;
}

export function formatMoney(amount: number, currency: string): string {
  return new Intl.NumberFormat(LOCALE, { style: 'currency', currency: currency.toUpperCase() })
    .format(amount)
    .replace('-', '−');
}

export function formatDuration(milliseconds: number): string {
  const total = Math.max(0, Math.round(milliseconds / 1000));
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor((total % 3600) / 60);
  const seconds = total % 60;
  const pad = (n: number) => String(n).padStart(2, '0');
  if (hours > 0) return `${hours}h ${pad(minutes)}m`;
  if (minutes > 0) return `${minutes}m ${pad(seconds)}s`;
  return `${seconds}s`;
}

export function formatLatency(milliseconds: number): string {
  if (milliseconds < 1) return `${trim(milliseconds, 2)} ms`;
  if (milliseconds < 1000) return `${Math.round(milliseconds)} ms`;
  if (milliseconds < 60_000) return `${trim(milliseconds / 1000, 1)} s`;
  return `${trim(milliseconds / 60_000, 1)} min`;
}

type DateInput = Date | string | number | null | undefined;

function toDate(value: Date | string | number): Date {
  return value instanceof Date ? value : new Date(value);
}

export function formatDate(value: DateInput): string {
  if (value == null) return '';
  return toDate(value).toLocaleDateString(LOCALE, { month: 'short', day: 'numeric', year: 'numeric' });
}

export function formatDateTime(value: DateInput): string {
  if (value == null) return '';
  const date = toDate(value);
  const thisYear = date.getFullYear() === new Date().getFullYear();
  return date.toLocaleString(LOCALE, {
    month: 'short',
    day: 'numeric',
    year: thisYear ? undefined : 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  });
}

export function formatTime(value: DateInput): string {
  if (value == null) return '';
  return toDate(value).toLocaleTimeString(LOCALE, { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' });
}

export function formatExactTime(value: DateInput): string {
  if (value == null) return '';
  return toDate(value).toLocaleString(LOCALE, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
    timeZoneName: 'short',
  });
}

export function formatRelative(value: DateInput, now: number = Date.now()): string {
  if (value == null) return '';
  const seconds = Math.round((now - toDate(value).getTime()) / 1000);
  if (seconds < 45) return 'just now';
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.round(hours / 24);
  if (days < 30) return `${days}d ago`;
  return formatDate(value);
}
