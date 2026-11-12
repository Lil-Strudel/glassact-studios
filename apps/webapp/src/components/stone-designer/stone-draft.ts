import type {
  GET,
  InlayWithInfo,
  PutStoneDesignRequest,
  StoneDesign,
  StoneElement,
} from "@glassact/data";
import { DEFAULT_CUSTOM_HEX, DEFAULT_GRANITE_KEY } from "../granite/granite";
import type { StoneDraft } from "./create-stone-editor";
import type { PlacedInlayArt } from "./stone-canvas";
import { STONE_SHAPE_INFO } from "./stone-shapes";
import { cloneJson } from "./stone-geometry";

export function defaultStoneDraft(): StoneDraft {
  const { width, height } = STONE_SHAPE_INFO.serpentine.defaultSize;
  return {
    shape: "serpentine",
    width,
    height,
    graniteKey: DEFAULT_GRANITE_KEY,
    customHex: DEFAULT_CUSTOM_HEX,
    layout: { elements: [], groups: [] },
  };
}

export function draftFromDesign(design: GET<StoneDesign> | null): StoneDraft {
  if (!design) return defaultStoneDraft();
  return {
    shape: design.stone_shape,
    width: design.stone_width,
    height: design.stone_height,
    graniteKey: design.granite_key,
    customHex: design.granite_custom_hex ?? DEFAULT_CUSTOM_HEX,
    layout: design.layout,
  };
}

export function requestFromDraft(
  draft: StoneDraft,
  version: number,
): PutStoneDesignRequest {
  return {
    version,
    stone_shape: draft.shape,
    stone_width: draft.width,
    stone_height: draft.height,
    granite_key: draft.graniteKey,
    granite_custom_hex: draft.customHex,
    layout: cloneJson(draft.layout),
  };
}

export function inlayArt(inlay: InlayWithInfo): PlacedInlayArt | undefined {
  if (inlay.width == null || inlay.height == null) return undefined;
  return {
    name: inlay.name,
    imageUrl: inlay.preview_url || null,
    width: inlay.width,
    height: inlay.height,
  };
}

export function describeElement(
  el: StoneElement,
  inlayName: (uuid: string) => string | undefined,
  backgroundName: (uuid: string) => string | undefined,
): string {
  switch (el.type) {
    case "inlay":
      return inlayName(el.inlay_uuid) ?? "Inlay";
    case "background":
      return `Background: ${backgroundName(el.background_uuid) ?? "sandblast"}`;
    case "text": {
      const firstLine = el.text.split("\n")[0].trim();
      return `Text: ${firstLine || "(empty)"}`;
    }
  }
}
