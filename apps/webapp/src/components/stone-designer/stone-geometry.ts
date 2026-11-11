export interface Point {
  x: number;
  y: number;
}

// A box in an element's own (unrotated) frame, relative to its center.
export interface LocalBox {
  x: number;
  y: number;
  width: number;
  height: number;
}

export function rotatePoint(p: Point, center: Point, degrees: number): Point {
  const rad = (degrees * Math.PI) / 180;
  const cos = Math.cos(rad);
  const sin = Math.sin(rad);
  const dx = p.x - center.x;
  const dy = p.y - center.y;
  return {
    x: center.x + dx * cos - dy * sin,
    y: center.y + dx * sin + dy * cos,
  };
}

export function boxCorners(
  center: Point,
  rotation: number,
  box: LocalBox,
): Point[] {
  const corners = [
    { x: box.x, y: box.y },
    { x: box.x + box.width, y: box.y },
    { x: box.x + box.width, y: box.y + box.height },
    { x: box.x, y: box.y + box.height },
  ];
  return corners.map((c) =>
    rotatePoint({ x: center.x + c.x, y: center.y + c.y }, center, rotation),
  );
}

export function boundsOf(points: Point[]) {
  const xs = points.map((p) => p.x);
  const ys = points.map((p) => p.y);
  const minX = Math.min(...xs);
  const minY = Math.min(...ys);
  const maxX = Math.max(...xs);
  const maxY = Math.max(...ys);
  return {
    x: minX,
    y: minY,
    width: maxX - minX,
    height: maxY - minY,
    center: { x: (minX + maxX) / 2, y: (minY + maxY) / 2 },
  };
}

export function angleDegrees(from: Point, to: Point): number {
  return (Math.atan2(to.y - from.y, to.x - from.x) * 180) / Math.PI;
}

export function normalizeRotation(degrees: number): number {
  const wrapped = ((degrees % 360) + 540) % 360 - 180;
  return Number(wrapped.toFixed(2));
}

export function newElementId(): string {
  return crypto.randomUUID();
}

// Solid store proxies (including TanStack Query results) cannot go through
// structuredClone; designer state is plain JSON, so this is lossless.
export function cloneJson<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}
