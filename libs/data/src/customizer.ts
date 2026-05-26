// Shapes for the catalog inlay color customizer. The design manifest is produced
// by the Go SVG ingest step and stored on `catalog_items.manifest`. ColorOverrides
// is the durable changelist the customizer edits and the bake endpoint consumes;
// it mirrors what a future `inlay_proofs.color_overrides` will store.

// One recolorable color region of a design, keyed in the manifest by the design's
// original source hex. `class` is the source SVG CSS class (null for implicit black).
export interface ManifestRegion {
  class: string | null;
  piece_ids: string[];
  count: number;
}

export interface Manifest {
  view_box: string;
  regions: Record<string, ManifestRegion>;
}

export interface GlassColorRef {
  glass_color_id: number;
}

export interface GroutRef {
  grout_id: number;
}

// Resolution order at render/bake: piece override -> region mapping -> source hex.
// `regions` is keyed by source hex; `pieces` is keyed by stable piece id (p0, p1, ...).
// Any entry in `pieces` is, by its presence, a tracked individual override.
export interface ColorOverrides {
  regions?: Record<string, GlassColorRef>;
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
