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

export const WORDMARK_WIDTH = 2681;
export const WORDMARK_CAP = 729;

/** The wordmark's outline, cap line at -729 and baseline at 0. The y's step reaches down to 200. */
export const WORDMARK_PATH = 'M2167.8 -209.6Q2167.8 -136.8 2191.1 -88Q2214.5 -39.2 2258.9 -15Q2303.4 9.1 2366.7 9.1Q2430.1 9.1 2478.6 -22.1Q2527.1 -53.3 2550.6 -110.3H2552.5V40C2552.5 100.7 2413.8 61.3 2413.8 122V200H2594.9V122C2594.9 61.3 2681 100.7 2681 40V-532.4H2552.5V-251.5Q2552.5 -205.3 2534.9 -169.6Q2517.3 -133.9 2485.8 -114.3Q2454.3 -94.7 2414.5 -94.7Q2352.5 -94.7 2324.4 -131.5Q2296.2 -168.3 2296.2 -240.9V-532.4H2167.8ZM465.8 -209.6Q465.8 -136.8 489.1 -88Q512.5 -39.2 556.9 -15Q601.4 9.1 664.7 9.1Q728.1 9.1 776.6 -22.1Q825.1 -53.3 848.6 -110.3H850.6L855.9 0H979V-532.4H850.5V-251.5Q850.5 -205.3 832.9 -169.6Q815.3 -133.9 783.8 -114.3Q752.3 -94.7 712.5 -94.7Q650.5 -94.7 622.4 -131.5Q594.2 -168.3 594.2 -240.9V-532.4H465.8ZM1031.8 -266.4Q1031.8 -183.7 1065.4 -121.3Q1098.9 -59 1160.6 -24.9Q1222.3 9.1 1305.3 9.1Q1380.6 9.1 1438.6 -16.1Q1496.7 -41.4 1531.7 -88.9Q1566.8 -136.3 1574.3 -201.5H1443.3Q1433.9 -150.4 1398.3 -122.8Q1362.7 -95.2 1305.8 -95.2Q1261.3 -95.2 1229.1 -114.9Q1196.9 -134.6 1179.3 -172.9Q1161.6 -211.1 1161.6 -265.8Q1161.6 -318.1 1178.5 -356.4Q1195.3 -394.7 1227.8 -416Q1260.4 -437.4 1306.9 -437.4Q1360.6 -437.4 1395.2 -411.5Q1429.9 -385.6 1441.3 -334.9H1573.1Q1566.2 -398.6 1532 -445.2Q1497.9 -491.8 1441.1 -516.8Q1384.3 -541.7 1310 -541.7Q1225.7 -541.7 1162.9 -507.8Q1100.1 -473.9 1065.9 -411.5Q1031.8 -349.1 1031.8 -266.4ZM0 -729V0H435.7V-118.9H135.3V-729ZM1889.9 -129.3Q1889.9 -62.9 1921.4 -31.5Q1953 0 2018.4 0H2135.4V-106.6H2067Q2042.3 -106.6 2030 -117.6Q2017.7 -128.7 2017.7 -154.9V-429.9H2134.7V-532.4H2017.7V-729H1889.9V-532.4H1790.7V-429.9H1889.9ZM1626.8 -532.4V0H1754.9V-532.4ZM1631.1 -609.4H1750.7V-729H1631.1Z';
