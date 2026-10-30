type Point = [number, number];

type Shape = {
  bw: number;
  bh: number;
  flat: number;
  t1: number;
  t2: number;
  top: number;
  bot: number;
  dT: number;
  dB: number;
};

const rest: Shape = { bw: 60, bh: 36, flat: 10, t1: 0.74, t2: 0.74, top: 76, bot: 33, dT: 120, dB: 169 };

const upperKeys = new Set(['A0', 'A', 'b1c1', 'b1c2', 'B']);

/** Draws the bend mark in a 250 by 181 box, reshaped by changes to its rest shape and mirrored when flipped. */
export function markPath(changes: Partial<Shape>, flip = false): string {
  const s = { ...rest, ...changes };
  const width = 2 * s.bw + s.flat;
  const cx = s.bw + s.flat;
  const dx = 2 * s.bw + s.flat;
  const mirror = (x: number, y: number): Point => [width - x, y];

  const left = {
    A0: mirror(0, 0),
    A: mirror(0, s.top),
    b1c1: mirror(0, s.top + s.bh * s.t1),
    b1c2: mirror(s.bw * (1 - s.t1), s.top + s.bh),
    B: mirror(s.bw, s.top + s.bh),
    C: mirror(cx, s.top + s.bh),
    b2c1: mirror(cx + s.bw * s.t2, s.top + s.bh),
    b2c2: mirror(dx, s.top + 2 * s.bh - s.bh * s.t2),
    D: mirror(dx, s.top + 2 * s.bh),
    D1: mirror(dx, s.top + 2 * s.bh + s.bot),
  };

  const mid = (s.dT + s.dB) / 2;
  const upper = left.A[0] - left.B[0] || 1e-6;
  const lower = left.C[0] - left.D[0] || 1e-6;
  const clamp = (value: number) => Math.min(1, Math.max(0, value));

  let lo = Infinity;
  let hi = -Infinity;

  const right = Object.fromEntries(
    Object.entries(left).map(([key, [x, y]]) => {
      const offset = upperKeys.has(key)
        ? s.dT + (mid - s.dT) * clamp((left.A[0] - x) / upper)
        : mid + (s.dB - mid) * clamp((left.C[0] - x) / lower);

      lo = Math.min(lo, x, x + offset);
      hi = Math.max(hi, x, x + offset);

      return [key, [x + offset, y]];
    }),
  ) as typeof left;

  const shift = 125 - (lo + hi) / 2;
  const p = ([x, y]: Point) => `${(flip ? 250 - (x + shift) : x + shift).toFixed(2)} ${y.toFixed(2)}`;
  const l = left;
  const r = right;

  return `M ${p(l.A0)} L ${p(l.A)} C ${p(l.b1c1)} ${p(l.b1c2)} ${p(l.B)} L ${p(l.C)} C ${p(l.b2c1)} ${p(l.b2c2)} ${p(l.D)} L ${p(l.D1)}`
    + ` L ${p(r.D1)} L ${p(r.D)} C ${p(r.b2c2)} ${p(r.b2c1)} ${p(r.C)} L ${p(r.B)} C ${p(r.b1c2)} ${p(r.b1c1)} ${p(r.A)} L ${p(r.A0)} Z`;
}
