import { onBeforeUnmount, ref } from 'vue';

const STIFFNESS = 260;
const DAMPING = 0.42;

/**
 * Springs the logo's flex between rest (0) and stretched (1), with a small
 * bounce on the way back. Does nothing when the viewer prefers reduced motion.
 */
export function useLogoFlex() {
  const flex = ref(0);
  let velocity = 0;
  let target = 0;
  let frame = 0;
  let lastFrame = 0;

  function step(now: number) {
    const dt = Math.min(0.032, (now - lastFrame) / 1000);
    lastFrame = now;
    velocity += (STIFFNESS * (target - flex.value) - 2 * DAMPING * Math.sqrt(STIFFNESS) * velocity) * dt;
    const next = flex.value + velocity * dt;
    if (Math.abs(target - next) < 0.001 && Math.abs(velocity) < 0.001) {
      flex.value = target;
      velocity = 0;
      frame = 0;
      return;
    }
    flex.value = next;
    frame = requestAnimationFrame(step);
  }

  function flexTo(value: number) {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    target = value;
    if (frame) return;
    lastFrame = performance.now();
    frame = requestAnimationFrame(step);
  }

  onBeforeUnmount(() => cancelAnimationFrame(frame));

  return { flex, flexTo };
}
