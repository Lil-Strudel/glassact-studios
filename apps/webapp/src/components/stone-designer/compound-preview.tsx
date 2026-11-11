import { createMemo } from "solid-js";
import type { CatalogCompoundMember } from "@glassact/data";
import { StoneCanvas } from "./stone-canvas";
import { compoundArtboard } from "./compound-layout";

interface CompoundPreviewProps {
  members: CatalogCompoundMember[];
  class?: string;
}

export function CompoundPreview(props: CompoundPreviewProps) {
  const layout = createMemo(() => compoundArtboard(props.members));
  return (
    <StoneCanvas
      stone={null}
      artboard={layout().artboard}
      elements={layout().elements}
      resolveInlay={(uuid) => layout().art.get(uuid)}
      resolveBackground={() => undefined}
      class={props.class}
    />
  );
}
