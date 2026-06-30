import { toValue, type MaybeRefOrGetter } from 'vue';
import { useStorage } from '@vueuse/core';

export interface NodePosition {
  x: number;
  y: number;
}

type EnvironmentLayouts = Record<string, Record<string, NodePosition>>;

const layouts = useStorage<EnvironmentLayouts>('lucity_canvas_layout', {});

function isUsable(position: NodePosition | undefined): position is NodePosition {
  return (
    !!position &&
    Number.isFinite(position.x) &&
    Number.isFinite(position.y)
  );
}

export function useCanvasLayout(environmentId: MaybeRefOrGetter<string | undefined>) {
  function positionFor(nodeId: string): NodePosition | null {
    const envId = toValue(environmentId);
    if (!envId) return null;
    const position = layouts.value[envId]?.[nodeId];
    return isUsable(position) ? position : null;
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
