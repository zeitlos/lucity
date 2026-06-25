export interface EnvAssignment {
  key: string;
  value: string;
}

// Parses pasted `KEY=value` text into assignments. Handles single lines and
// multiline blocks, `export ` prefixes, `#` comments, blank lines, and
// surrounding quotes. Splits on the first `=` so values may contain `=`.
export function parseEnvAssignments(text: string): EnvAssignment[] {
  const assignments: EnvAssignment[] = [];

  for (const rawLine of text.split(/\r?\n/)) {
    let line = rawLine.trim();

    if (!line || line.startsWith('#')) continue;
    if (line.startsWith('export ')) line = line.slice('export '.length).trim();

    const eq = line.indexOf('=');
    if (eq <= 0) continue;

    const key = line.slice(0, eq).trim();
    let value = line.slice(eq + 1).trim();

    if (
      value.length >= 2 &&
      ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'")))
    ) {
      value = value.slice(1, -1);
    }

    if (key) assignments.push({ key, value });
  }

  return assignments;
}
