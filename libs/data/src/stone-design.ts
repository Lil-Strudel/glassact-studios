import type { StandardTable } from "./helpers";

export type StoneShape =
  | "flat-top"
  | "serpentine"
  | "half-serpentine"
  | "oval-top"
  | "cathedral"
  | "heart"
  | "flat-marker"
  | "slant";

export const STONE_SHAPES: StoneShape[] = [
  "flat-top",
  "serpentine",
  "half-serpentine",
  "oval-top",
  "cathedral",
  "heart",
  "flat-marker",
  "slant",
];

export type StoneTextFont = "roman" | "serif" | "script";
export const STONE_TEXT_FONTS: StoneTextFont[] = ["roman", "serif", "script"];

export type StoneTextAlign = "left" | "center" | "right";

export type StoneTextFinish = "engraved" | "frosted";
export const STONE_TEXT_FINISHES: StoneTextFinish[] = ["engraved", "frosted"];

export const MAX_STONE_ELEMENTS = 300;
export const MAX_STONE_TEXT_LENGTH = 500;

// x/y are the element's center in inches from the stone's top-left.
type StoneElementBase = {
  id: string;
  x: number;
  y: number;
  rotation: number;
  group_id: string | null;
};

// An inlay's size is the inlay's own; the layout only positions it.
export type StoneInlayElement = StoneElementBase & {
  type: "inlay";
  inlay_uuid: string;
};

export type StoneBackgroundElement = StoneElementBase & {
  type: "background";
  catalog_item_uuid: string;
  width: number;
  flip_x?: boolean;
};

export type StoneTextElement = StoneElementBase & {
  type: "text";
  text: string;
  font: StoneTextFont;
  // Cap height in inches.
  size: number;
  align: StoneTextAlign;
  finish: StoneTextFinish;
};

export type StoneElement =
  | StoneInlayElement
  | StoneBackgroundElement
  | StoneTextElement;

export type StoneGroup = {
  id: string;
  name: string;
  compound_uuid: string | null;
};

export type StoneLayout = {
  elements: StoneElement[];
  groups: StoneGroup[];
};

export type StoneDesign = StandardTable<{
  project_id: number;
  stone_shape: StoneShape;
  stone_width: number;
  stone_height: number;
  granite_key: string;
  granite_custom_hex: string | null;
  layout: StoneLayout;
}>;

export type PutStoneDesignRequest = {
  // 0 when the project has no saved design yet.
  version: number;
  stone_shape: StoneShape;
  stone_width: number;
  stone_height: number;
  granite_key: string;
  granite_custom_hex: string | null;
  layout: StoneLayout;
};
