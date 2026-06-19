import { toValue, type MaybeRefOrGetter } from 'vue';
import { useStorage } from '@vueuse/core';

export interface NodePosition {
  x: number;
  y: number;
}

type EnvironmentLayouts = Record<string, Record<string, NodePosition>>;

const layouts = useStorage<EnvironmentLayouts>('lucity_canvas_layout', {});

export function useCanvasLayout(environmentId: MaybeRefOrGetter<string | undefined>) {
  function positionFor(nodeId: string): NodePosition | null {
    const envId = toValue(environmentId);
    if (!envId) return null;
    return layouts.value[envId]?.[nodeId] ?? null;
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

  return { positionFor, setPosition };
}
