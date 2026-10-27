const paletteSize = 10;

function hash(value: string): number {
  let result = 0x811c9dc5;
  for (const character of value) {
    result ^= character.codePointAt(0) ?? 0;
    result = Math.imul(result, 0x01000193);
  }
  return result >>> 0;
}

export function workspaceColorIndexes(workspaceIds: string[]): Map<string, number> {
  const taken = new Set<number>();
  const indexes = new Map<string, number>();
  for (const id of [...new Set(workspaceIds)].sort()) {
    let index = hash(id) % paletteSize;
    while (taken.size < paletteSize && taken.has(index)) {
      index = (index + 1) % paletteSize;
    }
    taken.add(index);
    indexes.set(id, index);
  }
  return indexes;
}

export function workspaceColorStyle(index: number) {
  return {
    backgroundColor: `var(--workspace-${index}-bg)`,
    color: `var(--workspace-${index}-fg)`,
  };
}

export function fallbackColorIndex(workspaceId: string): number {
  return hash(workspaceId) % paletteSize;
}
