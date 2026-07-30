export function severityColor(severity: string): string {
  switch (severity) {
    case 'CRITICAL':
      return 'var(--status-danger)';
    case 'HIGH':
      return 'color-mix(in oklch, var(--status-danger), var(--status-warn))';
    case 'MEDIUM':
      return 'var(--status-warn)';
    case 'LOW':
      return 'var(--status-progress)';
    default:
      return 'var(--status-neutral)';
  }
}

export function severityStyle(severity: string) {
  const color = severityColor(severity);

  return {
    color,
    backgroundColor: `color-mix(in srgb, ${color} 12%, transparent)`,
  };
}

export function severityLabel(severity: string): string {
  return severity.charAt(0) + severity.slice(1).toLowerCase();
}
