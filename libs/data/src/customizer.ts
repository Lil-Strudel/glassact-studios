// Shapes for the catalog inlay color customizer. The design manifest is produced
// by the Go SVG ingest step, perfected in the admin manifest editor, and stored
// on `catalog_items.manifest`. It carries the design's default coloring: each
// glass group and the grout region reference a real glass_color / grout id.
//
// ColorOverrides is the durable changelist the consumer customizer edits and the
// bake endpoint consumes; it layers on top of the manifest's default colors and
// mirrors what a future `inlay_proofs.color_overrides` will store.

// The single grout region. All pieces classified as grout collapse here (the
// design only ever has one grout color). `grout_id` is null until assigned in
// the editor — an unassigned region blocks saving the catalog item.
export interface GroutRegion {
  grout_id: number | null;
  piece_ids: string[];
  count: number;
}

// One recolorable glass color group. Keyed in the manifest by a stable group key
// (e.g. "group-0") assigned at ingest — never the display label, so consumer
// overrides keyed by group key survive a re-color in the editor.
export interface GlassRegion {
  glass_color_id: number | null;
  piece_ids: string[];
  count: number;
  // Provenance from the source SVG (display + best-guess matching only).
  source_class?: string | null;
  source_hex?: string | null;
}

export interface Manifest {
  view_box: string; // "0 0 W H", W = width * 300, H = height * 300
  grout_region: GroutRegion;
  glass_regions: Record<string, GlassRegion>; // keyed by stable group key
}

export interface GlassColorRef {
  glass_color_id: number;
}

export interface GroutRef {
  grout_id: number;
}

// Resolution order at render/bake: piece override -> group override -> manifest
// default. `groups` is keyed by group key; `pieces` is keyed by stable piece id
// (p0, p1, ...). Any entry in `pieces` is, by its presence, a tracked override.
export interface ColorOverrides {
  groups?: Record<string, GlassColorRef>;
  pieces?: Record<string, GlassColorRef>;
  background?: GroutRef;
}

export interface BakeRequest {
  scale_factor: number;
  width: number;
  height: number;
  color_overrides: ColorOverrides;
}

export interface BakeResult {
  design_asset_url: string;
  color_overrides: ColorOverrides;
  scale_factor: number;
  width: number;
  height: number;
}
