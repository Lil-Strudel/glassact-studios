import type { Manifest, ColorOverrides, GlassColor, GET } from "@glassact/data";

// Color resolution helpers for the customizer, keyed by stable manifest group
// keys ("group-0", ...) rather than source hex. A manifest's `glass_regions`
// carries each group's default `glass_color_id`; `ColorOverrides` layers
// per-group then per-piece overrides on top.

// What the user currently has selected for editing.
export type Selection =
  | { type: "group"; groupKey: string }
  | { type: "piece"; pieceId: string; groupKey: string };

export type GlassById = Map<number, GET<GlassColor>>;

// Neutral fill used only for the in-editor unassigned state (a saved catalog
// item always has a default glass_color_id on every group).
const NEUTRAL_FALLBACK = "#cccccc";

// pieceId -> groupKey, flattened from the manifest's glass_regions.
// Grout pieces are excluded — they are not interactive glass.
export function buildPieceSourceMap(
  manifest: Manifest | undefined,
): Map<string, string> {
  const map = new Map<string, string>();
  for (const [groupKey, region] of Object.entries(
    manifest?.glass_regions ?? {},
  )) {
    for (const id of region.piece_ids) {
      map.set(id, groupKey);
    }
  }
  return map;
}

// The piece IDs that make up the single grout region.
export function buildGroutPieceIds(manifest: Manifest | undefined): string[] {
  return manifest?.grout_region.piece_ids ?? [];
}

// Resolution order: piece override -> group override -> manifest group default
// glass_color_id -> neutral fallback.
export function resolvePieceHex(
  pieceId: string,
  groupKey: string,
  overrides: ColorOverrides,
  manifest: Manifest | undefined,
  glassById: GlassById,
): string {
  const piece = overrides.pieces?.[pieceId];
  if (piece) {
    return glassById.get(piece.glass_color_id)?.hex ?? NEUTRAL_FALLBACK;
  }
  const group = overrides.groups?.[groupKey];
  if (group) {
    return glassById.get(group.glass_color_id)?.hex ?? NEUTRAL_FALLBACK;
  }
  const defaultId = manifest?.glass_regions?.[groupKey]?.glass_color_id;
  if (defaultId != null) {
    return glassById.get(defaultId)?.hex ?? NEUTRAL_FALLBACK;
  }
  return NEUTRAL_FALLBACK;
}

// The glass color currently effective for a whole group: the group override if
// present, otherwise the manifest default. Returns null only when the group is
// unassigned (in-editor state).
export function groupGlassId(
  groupKey: string,
  overrides: ColorOverrides,
  manifest: Manifest | undefined,
): number | null {
  const override = overrides.groups?.[groupKey]?.glass_color_id;
  if (override != null) return override;
  return manifest?.glass_regions?.[groupKey]?.glass_color_id ?? null;
}

// Number of pieces in a group that carry an individual override (the price
// driver).
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
