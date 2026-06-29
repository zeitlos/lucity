import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Extract an error message from an unknown caught value.
 * Use in catch blocks: `catch (e: unknown) { errorToast(errorMessage(e)); }`
 */
export function errorMessage(e: unknown): string {
  if (e instanceof Error) return e.message;
  if (typeof e === 'string') return e;
  return String(e);
}

export function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / Math.pow(1024, i);
  return `${value.toFixed(i === 0 || value >= 100 ? 0 : 1)} ${units[i]}`;
}

export function parseCpu(cpu: string): number {
  const trimmed = cpu.trim();
  if (trimmed.endsWith('m')) {
    const millis = parseFloat(trimmed.slice(0, -1));
    return isNaN(millis) ? 0 : millis / 1000;
  }
  const cores = parseFloat(trimmed);
  return isNaN(cores) ? 0 : cores;
}

export function parseStorageSize(size: string): number {
  const match = /^(\d+(?:\.\d+)?)\s*([KMGTPE]i?)?$/.exec(size.trim());
  if (!match) return 0;
  const value = parseFloat(match[1]!);
  const unit = match[2];
  if (!unit) return value;
  const exponents: Record<string, number> = { K: 1, M: 2, G: 3, T: 4, P: 5, E: 6 };
  const base = unit.endsWith('i') ? 1024 : 1000;
  return value * Math.pow(base, exponents[unit[0]!] ?? 0);
}
