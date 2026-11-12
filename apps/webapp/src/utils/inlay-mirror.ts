export type MirrorFlags = { mirror_x: boolean; mirror_y: boolean };

export function isMirrored(inlay: MirrorFlags): boolean {
  return inlay.mirror_x || inlay.mirror_y;
}

export function mirrorScale(inlay: MirrorFlags): { x: number; y: number } {
  return { x: inlay.mirror_x ? -1 : 1, y: inlay.mirror_y ? -1 : 1 };
}

export function mirrorCssTransform(inlay: MirrorFlags): string | undefined {
  if (!isMirrored(inlay)) return undefined;
  const scale = mirrorScale(inlay);
  return `scale(${scale.x}, ${scale.y})`;
}

export function mirrorDescription(inlay: MirrorFlags): string | null {
  if (inlay.mirror_x && inlay.mirror_y) return "Mirrored left↔right and top↔bottom";
  if (inlay.mirror_x) return "Mirrored left↔right";
  if (inlay.mirror_y) return "Mirrored top↔bottom";
  return null;
}
