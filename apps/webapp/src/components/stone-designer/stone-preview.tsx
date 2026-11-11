import { createMemo } from "solid-js";
import type { GET, InlayWithInfo, StoneDesign } from "@glassact/data";
import { graniteChoice } from "../granite/granite";
import { StoneCanvas } from "./stone-canvas";
import { inlayArt } from "./stone-draft";
import { useBackgroundArt } from "./use-background-art";

interface StonePreviewProps {
  design: GET<StoneDesign>;
  inlays: InlayWithInfo[];
  class?: string;
}

export function StonePreview(props: StonePreviewProps) {
  const inlaysByUuid = createMemo(
    () => new Map(props.inlays.map((inlay) => [inlay.uuid, inlay])),
  );
  const resolveBackground = useBackgroundArt(() => props.design.layout.elements);

  return (
    <StoneCanvas
      stone={{
        shape: props.design.stone_shape,
        width: props.design.stone_width,
        height: props.design.stone_height,
        granite: graniteChoice(props.design.granite_key, props.design.granite_custom_hex ?? ""),
      }}
      elements={props.design.layout.elements}
      resolveInlay={(uuid) => {
        const inlay = inlaysByUuid().get(uuid);
        return inlay ? inlayArt(inlay) : undefined;
      }}
      resolveBackground={resolveBackground}
      class={props.class}
    />
  );
}
