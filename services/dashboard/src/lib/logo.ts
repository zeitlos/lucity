export const LOGO_WIDTH = 250;
export const LOGO_HEIGHT = 181;

const BEND_HEIGHT = 36;
const TOP = 76;
const BOTTOM = 33;
const DEPTH_TOP = 120;
const DEPTH_BOTTOM = 169;

type Point = [number, number];
type Edge = Record<'A0' | 'A' | 'b1c1' | 'b1c2' | 'B' | 'C' | 'b2c1' | 'b2c2' | 'D' | 'D1', Point>;

function profile(bend: number, flat: number, tension: number): Edge {
  const width = 2 * bend + flat;
  const middle = TOP + BEND_HEIGHT;
  return {
    A0: [width, 0],
    A: [width, TOP],
    b1c1: [width, TOP + BEND_HEIGHT * tension],
    b1c2: [width - bend * (1 - tension), middle],
    B: [bend + flat, middle],
    C: [bend, middle],
    b2c1: [bend * (1 - tension), middle],
    b2c2: [0, middle + BEND_HEIGHT * (1 - tension)],
    D: [0, middle + BEND_HEIGHT],
    D1: [0, middle + BEND_HEIGHT + BOTTOM],
  };
}

function extrude(left: Edge): Edge {
  const middle = (DEPTH_TOP + DEPTH_BOTTOM) / 2;
  const clamp = (value: number) => Math.min(1, Math.max(0, value));
  const upper = ([x, y]: Point): Point => [x + DEPTH_TOP + (middle - DEPTH_TOP) * clamp((left.A[0] - x) / (left.A[0] - left.B[0])), y];
  const lower = ([x, y]: Point): Point => [x + middle + (DEPTH_BOTTOM - middle) * clamp((left.C[0] - x) / (left.C[0] - left.D[0])), y];
  return {
    A0: upper(left.A0),
    A: upper(left.A),
    b1c1: upper(left.b1c1),
    b1c2: upper(left.b1c2),
    B: upper(left.B),
    C: lower(left.C),
    b2c1: lower(left.b2c1),
    b2c2: lower(left.b2c2),
    D: lower(left.D),
    D1: lower(left.D1),
  };
}

function point([x, y]: Point) {
  return `${+x.toFixed(2)} ${+y.toFixed(2)}`;
}

/**
 * The mark's outline. A flex of 1 widens the bend and softens it, the way the
 * logo stretches on hover; 0 is the mark at rest.
 */
export function logoPath(flex = 0) {
  const l = profile(60 + 16 * flex, 10 + 6.4 * flex, 0.74 - 0.08 * flex);
  const r = extrude(l);
  return `M ${point(l.A0)} L ${point(l.A)} C ${point(l.b1c1)}, ${point(l.b1c2)}, ${point(l.B)} L ${point(l.C)} C ${point(l.b2c1)}, ${point(l.b2c2)}, ${point(l.D)} L ${point(l.D1)}`
    + ` L ${point(r.D1)} L ${point(r.D)} C ${point(r.b2c2)}, ${point(r.b2c1)}, ${point(r.C)} L ${point(r.B)} C ${point(r.b1c2)}, ${point(r.b1c1)}, ${point(r.A)} L ${point(r.A0)} Z`;
}

export const LOGO_PATH = logoPath();
