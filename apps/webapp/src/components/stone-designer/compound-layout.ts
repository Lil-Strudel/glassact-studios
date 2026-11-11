import type { CatalogCompoundMember, StoneInlayElement } from "@glassact/data";
import type { PlacedInlayArt } from "./stone-canvas";
import { boxCorners } from "./stone-geometry";

const MARGIN_INCHES = 0.75;

// Lays a compound's members out on an artboard just big enough to hold them.
// Member positions are offsets from the compound's center; on the artboard the
// fake inlay uuid of each element is the member's catalog item uuid.
export function compoundArtboard(members: CatalogCompoundMember[]) {
  const corners = members.flatMap((m) =>
    boxCorners({ x: m.center_x, y: m.center_y }, m.rotation_deg, {
      x: -m.catalog_item.default_width / 2,
      y: -m.catalog_item.default_height / 2,
      width: m.catalog_item.default_width,
      height: m.catalog_item.default_height,
    }),
  );
  const halfWidth = Math.max(1, ...corners.map((c) => Math.abs(c.x))) + MARGIN_INCHES;
  const halfHeight = Math.max(1, ...corners.map((c) => Math.abs(c.y))) + MARGIN_INCHES;

  const elements: StoneInlayElement[] = members.map((m, i) => ({
    id: `member-${i}`,
    type: "inlay",
    inlay_uuid: m.catalog_item.uuid,
    x: m.center_x + halfWidth,
    y: m.center_y + halfHeight,
    rotation: m.rotation_deg,
    group_id: null,
  }));

  const art = new Map<string, PlacedInlayArt>(
    members.map((m) => [
      m.catalog_item.uuid,
      {
        name: m.catalog_item.name,
        imageUrl: m.catalog_item.svg_url,
        width: m.catalog_item.default_width,
        height: m.catalog_item.default_height,
      },
    ]),
  );

  return {
    artboard: { width: halfWidth * 2, height: halfHeight * 2 },
    elements,
    art,
  };
}
