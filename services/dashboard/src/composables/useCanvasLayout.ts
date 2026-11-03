import { toValue, type MaybeRefOrGetter } from 'vue';
import { useStorage } from '@vueuse/core';

export interface NodePosition {
  x: number;
  y: number;
}

export interface NodeRect extends NodePosition {
  width: number;
  height: number;
}

type EnvironmentLayouts = Record<string, Record<string, NodePosition>>;

export const gridSize = 24;
const nodeGap = 2 * gridSize;
const laneWidth = 12 * gridSize + nodeGap;
const laneCount = 3;

const layouts = useStorage<EnvironmentLayouts>('lucity_canvas_layout', {});

function isUsable(position: NodePosition | undefined): position is NodePosition {
  return (
    !!position &&
    Number.isFinite(position.x) &&
    Number.isFinite(position.y)
  );
}

function snapToGrid(value: number) {
  return Math.round(value / gridSize) * gridSize;
}

function fits(rect: NodeRect, occupied: NodeRect[]) {
  return occupied.every(other =>
    rect.x >= other.x + other.width + nodeGap ||
    other.x >= rect.x + rect.width + nodeGap ||
    rect.y >= other.y + other.height + nodeGap ||
    other.y >= rect.y + rect.height + nodeGap,
  );
}

export function firstFreeSpot(size: { width: number; height: number }, occupied: NodeRect[]): NodePosition {
  for (let y = 0; ; y += gridSize) {
    for (let lane = 0; lane < laneCount; lane++) {
      const spot = { x: lane * laneWidth, y };
      if (fits({ ...spot, ...size }, occupied)) return spot;
    }
  }
}

export function nearestFreeSpot(rect: NodeRect, occupied: NodeRect[]): NodePosition {
  let best: NodePosition = rect;
  let bestDistance = Infinity;
  for (let ring = 0; ring * gridSize <= bestDistance; ring++) {
    for (let dx = -ring; dx <= ring; dx++) {
      for (let dy = -ring; dy <= ring; dy++) {
        if (Math.max(Math.abs(dx), Math.abs(dy)) !== ring) continue;
        const distance = Math.hypot(dx, dy) * gridSize;
        const spot = { x: rect.x + dx * gridSize, y: rect.y + dy * gridSize };
        if (distance < bestDistance && fits({ ...rect, ...spot }, occupied)) {
          best = spot;
          bestDistance = distance;
        }
      }
    }
  }
  return { x: best.x, y: best.y };
}

export function useCanvasLayout(environmentId: MaybeRefOrGetter<string | undefined>) {
  function positionFor(nodeId: string): NodePosition | null {
    const envId = toValue(environmentId);
    if (!envId) return null;
    const position = layouts.value[envId]?.[nodeId];
    return isUsable(position) ? { x: snapToGrid(position.x), y: snapToGrid(position.y) } : null;
  }

  function setPosition(nodeId: string, position: NodePosition) {
    const envId = toValue(environmentId);
    if (!envId) return;
    layouts.value = {
      ...layouts.value,
      [envId]: {
        ...layouts.value[envId],
        [nodeId]: { x: position.x, y: position.y },
      },
    };
  }

  // Drop stored positions for nodes that no longer exist so a stale layout
  // can never accumulate or shadow the live set discovered from the API.
  function reconcile(nodeIds: string[]) {
    const envId = toValue(environmentId);
    if (!envId) return;
    const current = layouts.value[envId];
    if (!current) return;

    const live = new Set(nodeIds);
    const kept = Object.fromEntries(
      Object.entries(current).filter(([id]) => live.has(id)),
    );

    if (Object.keys(kept).length !== Object.keys(current).length) {
      layouts.value = { ...layouts.value, [envId]: kept };
    }
  }

  // Clear the saved layout for this environment (every node falls back to its
  // computed default position). The escape hatch when a layout goes bad.
  function reset() {
    const envId = toValue(environmentId);
    if (!envId) return;
    const { [envId]: _removed, ...rest } = layouts.value;
    layouts.value = rest;
  }

  return { positionFor, setPosition, reconcile, reset };
}
