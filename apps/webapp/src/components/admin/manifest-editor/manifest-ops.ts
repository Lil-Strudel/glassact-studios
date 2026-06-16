import type { Manifest, GlassRegion } from "@glassact/data";

// Pure manifest-editing helpers for the admin manifest editor. Every function
// returns a NEW manifest (never mutates) so callers can drop the result straight
// into a signal setter. Group keys are stable opaque strings ("group-0", ...) and
// are NEVER renamed — merges drop a key, splits mint a fresh one.

function cloneManifest(manifest: Manifest): Manifest {
  return {
    view_box: manifest.view_box,
    grout_region: {
      ...manifest.grout_region,
      piece_ids: [...manifest.grout_region.piece_ids],
    },
    glass_regions: Object.fromEntries(
      Object.entries(manifest.glass_regions).map(([key, region]) => [
        key,
        { ...region, piece_ids: [...region.piece_ids] },
      ]),
    ),
  };
}

// The next free "group-N" key that does not collide with an existing one.
export function nextGroupKey(manifest: Manifest): string {
  const existing = new Set(Object.keys(manifest.glass_regions));
  let n = 0;
  while (existing.has(`group-${n}`)) n += 1;
  return `group-${n}`;
}

// Assign a glass color id to a whole group. id can be null to clear it back to
// the unassigned (save-blocking) state.
export function assignGroupColor(
  manifest: Manifest,
  groupKey: string,
  glassColorId: number | null,
): Manifest {
  const next = cloneManifest(manifest);
  const region = next.glass_regions[groupKey];
  if (!region) return manifest;
  region.glass_color_id = glassColorId;
  return next;
}

// Assign the grout color id (or null to clear).
export function assignGroutColor(
  manifest: Manifest,
  groutId: number | null,
): Manifest {
  const next = cloneManifest(manifest);
  next.grout_region.grout_id = groutId;
  return next;
}

// Merge `sourceKey` into `targetKey`: the surviving group keeps its key + color,
// absorbs the source's pieces, and the source group is dropped.
export function mergeGroups(
  manifest: Manifest,
  sourceKey: string,
  targetKey: string,
): Manifest {
  if (sourceKey === targetKey) return manifest;
  const next = cloneManifest(manifest);
  const source = next.glass_regions[sourceKey];
  const target = next.glass_regions[targetKey];
  if (!source || !target) return manifest;

  const merged = new Set([...target.piece_ids, ...source.piece_ids]);
  target.piece_ids = [...merged];
  target.count = target.piece_ids.length;
  delete next.glass_regions[sourceKey];
  return next;
}

// Split the given pieces out of their current groups into a brand-new group.
// Returns the new manifest and the freshly minted key. Pieces are removed from
// whichever glass group currently holds them (grout pieces are ignored).
export function splitGroup(
  manifest: Manifest,
  pieceIds: string[],
): { manifest: Manifest; newKey: string } {
  const next = cloneManifest(manifest);
  const moving = new Set(pieceIds);
  for (const region of Object.values(next.glass_regions)) {
    region.piece_ids = region.piece_ids.filter((id) => !moving.has(id));
    region.count = region.piece_ids.length;
  }
  // Drop now-empty groups to avoid dangling empty regions.
  for (const [key, region] of Object.entries(next.glass_regions)) {
    if (region.piece_ids.length === 0) delete next.glass_regions[key];
  }
  const newKey = nextGroupKey(next);
  const newRegion: GlassRegion = {
    glass_color_id: null,
    piece_ids: [...moving],
    count: moving.size,
  };
  next.glass_regions[newKey] = newRegion;
  return { manifest: next, newKey };
}

// Move pieces into a target glass group (creating membership there, removing them
// from every other glass group and from grout). Empty source groups are dropped.
export function movePiecesToGroup(
  manifest: Manifest,
  pieceIds: string[],
  targetKey: string,
): Manifest {
  const target = manifest.glass_regions[targetKey];
  if (!target) return manifest;
  const next = cloneManifest(manifest);
  const moving = new Set(pieceIds);

  next.grout_region.piece_ids = next.grout_region.piece_ids.filter(
    (id) => !moving.has(id),
  );
  next.grout_region.count = next.grout_region.piece_ids.length;

  for (const region of Object.values(next.glass_regions)) {
    region.piece_ids = region.piece_ids.filter((id) => !moving.has(id));
  }

  const dest = next.glass_regions[targetKey];
  const destSet = new Set([...dest.piece_ids, ...moving]);
  dest.piece_ids = [...destSet];

  for (const [key, region] of Object.entries(next.glass_regions)) {
    region.count = region.piece_ids.length;
    if (region.piece_ids.length === 0 && key !== targetKey) {
      delete next.glass_regions[key];
    }
  }
  return next;
}

// Move pieces into the grout region, removing them from every glass group.
export function movePiecesToGrout(
  manifest: Manifest,
  pieceIds: string[],
): Manifest {
  const next = cloneManifest(manifest);
  const moving = new Set(pieceIds);

  for (const [key, region] of Object.entries(next.glass_regions)) {
    region.piece_ids = region.piece_ids.filter((id) => !moving.has(id));
    region.count = region.piece_ids.length;
    if (region.piece_ids.length === 0) delete next.glass_regions[key];
  }

  const groutSet = new Set([...next.grout_region.piece_ids, ...moving]);
  next.grout_region.piece_ids = [...groutSet];
  next.grout_region.count = next.grout_region.piece_ids.length;
  return next;
}

// Promote an entire glass group to grout: its pieces become grout pieces and the
// group is removed.
export function markGroupAsGrout(
  manifest: Manifest,
  groupKey: string,
): Manifest {
  const region = manifest.glass_regions[groupKey];
  if (!region) return manifest;
  return movePiecesToGrout(manifest, region.piece_ids);
}

// Pull pieces out of grout into a new glass group (the inverse of marking grout).
export function unmarkGroutPieces(
  manifest: Manifest,
  pieceIds: string[],
): { manifest: Manifest; newKey: string } {
  const next = cloneManifest(manifest);
  const moving = new Set(pieceIds);
  next.grout_region.piece_ids = next.grout_region.piece_ids.filter(
    (id) => !moving.has(id),
  );
  next.grout_region.count = next.grout_region.piece_ids.length;

  const newKey = nextGroupKey(next);
  next.glass_regions[newKey] = {
    glass_color_id: null,
    piece_ids: [...moving],
    count: moving.size,
  };
  return { manifest: next, newKey };
}

// Save is blocked while any glass group OR the grout region lacks a color id.
export function unassignedGroupKeys(manifest: Manifest): string[] {
  return Object.entries(manifest.glass_regions)
    .filter(([, region]) => region.glass_color_id == null)
    .map(([key]) => key);
}

export function isGroutAssigned(manifest: Manifest): boolean {
  return manifest.grout_region.grout_id != null;
}

export function isManifestComplete(manifest: Manifest): boolean {
  return unassignedGroupKeys(manifest).length === 0 && isGroutAssigned(manifest);
}
