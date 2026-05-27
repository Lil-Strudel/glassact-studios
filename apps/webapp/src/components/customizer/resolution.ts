import type { Manifest, ColorOverrides, GlassColor, GET } from "@glassact/data";

// What the user currently has selected for editing.
export type Selection =
  | { type: "region"; sourceHex: string }
  | { type: "piece"; pieceId: string; sourceHex: string };

export type GlassById = Map<number, GET<GlassColor>>;

const DEFAULT_FILL = "#000000";

// Returns the set of source hexes that identify grout regions:
//   1. The region containing "p0" (first/back-most shape in document order).
//   2. The "#000000" region (classless implicit-black shapes like grout lines/eyes).
// Both may be the same region or different ones.
export function getGroutSourceHexes(manifest: Manifest | undefined): Set<string> {
  const groutHexes = new Set<string>();
  const regions = manifest?.regions ?? {};
  for (const [hex, region] of Object.entries(regions)) {
    if (hex === DEFAULT_FILL || region.piece_ids.includes("p0")) {
      groutHexes.add(hex);
    }
  }
  return groutHexes;
}

// pieceId -> original source hex, flattened from the manifest's regions.
// All grout regions are excluded — they are not interactive glass.
export function buildPieceSourceMap(manifest: Manifest | undefined): Map<string, string> {
  const groutHexes = getGroutSourceHexes(manifest);
  const map = new Map<string, string>();
  for (const [hex, region] of Object.entries(manifest?.regions ?? {})) {
    if (groutHexes.has(hex)) continue;
    for (const id of region.piece_ids) {
      map.set(id, hex);
    }
  }
  return map;
}

// Returns all piece IDs that belong to any grout region.
export function buildGroutPieceIds(manifest: Manifest | undefined): string[] {
  const groutHexes = getGroutSourceHexes(manifest);
  const ids: string[] = [];
  for (const [hex, region] of Object.entries(manifest?.regions ?? {})) {
    if (groutHexes.has(hex)) ids.push(...region.piece_ids);
  }
  return ids;
}

// Resolution order: piece override -> region mapping -> original source hex.
export function resolvePieceHex(
  pieceId: string,
  sourceHex: string,
  overrides: ColorOverrides,
  glassById: GlassById,
): string {
  const piece = overrides.pieces?.[pieceId];
  if (piece) {
    return glassById.get(piece.glass_color_id)?.hex ?? sourceHex;
  }
  const region = overrides.regions?.[sourceHex];
  if (region) {
    return glassById.get(region.glass_color_id)?.hex ?? sourceHex;
  }
  return sourceHex;
}

// The glass color currently assigned to a whole region (group), if any.
export function regionGlassId(
  sourceHex: string,
  overrides: ColorOverrides,
): number | null {
  return overrides.regions?.[sourceHex]?.glass_color_id ?? null;
}

// Number of pieces in a region that have an individual override (the price driver).
export function customPieceCount(
  pieceIds: string[],
  overrides: ColorOverrides,
): number {
  const pieces = overrides.pieces ?? {};
  return pieceIds.reduce((n, id) => (pieces[id] ? n + 1 : n), 0);
}

export function totalCustomPieces(overrides: ColorOverrides): number {
  return Object.keys(overrides.pieces ?? {}).length;
}
