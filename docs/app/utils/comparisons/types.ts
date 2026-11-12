export type Mark = 'yes' | 'partly' | 'no' | 'info';

export interface Cell {
  mark: Mark;
  label?: string;
  text?: string;
  href?: string;
  differs?: boolean;
}

export interface LucityGroup {
  title: string;
  rows: { key: string; name: string; lucity: Cell }[];
}

export interface CompareRow {
  key: string;
  name: string;
  lucity: Cell;
  other: Cell;
  differs: boolean;
}

export interface CompareGroup {
  title: string;
  rows: CompareRow[];
}

export interface Point {
  title: string;
  text: string;
}

export interface Price {
  setup: string;
  detail: string;
  other: string;
  otherNote: string;
  lucity: string;
  lucityNote: string;
}

export interface Shot {
  name: string;
  title: string;
  note: string;
}

export interface Comparison {
  slug: string;
  key: string;
  name: string;
  icon: string;
  checked: string;
  description: string;
  intro: string;
  respect: string;
  sources: string;
  short: string;
  fitOther: string[];
  fitLucity: string[];
  shotsIntro: string;
  shots: Shot[];
  tableIntro: string;
  tableFoot: string;
  other: Record<string, Cell>;
  aheadOther: Point[];
  aheadLucity: Point[];
  differ: Point[];
  priceIntro: string;
  prices: Price[];
  priceMethod: string;
  moveIntro: string;
  map: string[][];
  steps: string[];
  history: { date: string; text: string }[];
}

export interface ComparisonPlatform {
  key: string;
  name: string;
  icon: string;
  to?: string;
}

export interface MatrixGroup {
  title: string;
  rows: { name: string; cells: Record<string, { mark: Mark; text: string }> }[];
}

export interface Fit {
  need: string;
  platforms: string[];
  why: string;
}

export interface Capture {
  kind: 'still' | 'video';
  title: string;
  file: string;
  poster?: string;
  seconds?: number;
  capturedAt: string;
}
