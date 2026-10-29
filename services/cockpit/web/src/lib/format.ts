const DAY = 24 * 60 * 60 * 1000;

export function relativeTime(iso: string): string {
  const then = new Date(iso);
  const minutes = Math.floor((Date.now() - then.getTime()) / 60000);

  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes} min ago`;
  if (minutes < 24 * 60) return `${Math.floor(minutes / 60)} h ago`;

  return dateTime(then);
}

export function daysSince(iso: string): number {
  return Math.floor((Date.now() - new Date(iso).getTime()) / DAY);
}

export function age(iso: string): string {
  const days = daysSince(iso);

  if (days < 1) return 'today';
  if (days < 365) return `${days} ${days === 1 ? 'day' : 'days'}`;

  const years = Math.floor(days / 365);
  return `${years} ${years === 1 ? 'year' : 'years'}`;
}

export function duration(days: number): string {
  if (days < 365) return `${days} ${days === 1 ? 'day' : 'days'}`;
  return `${(days / 365).toFixed(1)} years`;
}

export function bytes(value: number): string {
  if (value < 1e3) return `${Math.round(value)} B`;
  if (value < 1e6) return `${(value / 1e3).toFixed(1)} kB`;
  if (value < 1e9) return `${(value / 1e6).toFixed(1)} MB`;
  return `${(value / 1e9).toFixed(1)} GB`;
}

export function day(time: number): string {
  return new Date(time).toLocaleDateString(undefined, { day: 'numeric', month: 'short', year: 'numeric' });
}

export function daysUntil(iso: string): number {
  return Math.ceil((new Date(iso).getTime() - Date.now()) / DAY);
}

export function gigabytes(bytes: number): string {
  return `${(bytes / 1e9).toFixed(1)} GB`;
}

export function memory(bytes: number): string {
  if (bytes === 0) return '0';
  if (bytes < 2 ** 30) return `${Math.round(bytes / 2 ** 20)} MiB`;
  return `${(bytes / 2 ** 30).toFixed(1)} GiB`;
}

export function loadedTime(date: Date): string {
  if (date.toDateString() === new Date().toDateString()) {
    return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
  }

  return dateTime(date);
}

function dateTime(date: Date): string {
  const day = date.toLocaleDateString(undefined, { day: 'numeric', month: 'short' });
  const time = date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });

  return `${day} ${time}`;
}
