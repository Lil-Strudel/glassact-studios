import type { Manifest, ColorOverrides, GlassColor, GET } from "@glassact/data";

// What the user currently has selected for editing.
export type Selection =
  | { type: "region"; sourceHex: string }
  | { type: "piece"; pieceId: string; sourceHex: string };

export type GlassById = Map<number, GET<GlassColor>>;

// pieceId -> original source hex, flattened from the manifest's regions.
export function buildPieceSourceMap(manifest: Manifest | undefined): Map<string, string> {
  const map = new Map<string, string>();
  const regions = manifest?.regions ?? {};
  for (const [hex, region] of Object.entries(regions)) {
    for (const id of region.piece_ids) {
      map.set(id, hex);
    }
  }
  return map;
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
