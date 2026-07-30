import { createEffect, createMemo, createSignal, onCleanup, onMount } from "solid-js";
import type { GlassColor, Grout, Manifest, GET } from "@glassact/data";
import { cn } from "@glassact/ui";
import {
  buildGroutPieceIds,
  buildPieceSourceMap,
  resolvePieceHex,
  type GlassById,
} from "./resolution";

interface SvgPreviewProps {
  svgText: string;
  manifest: Manifest;
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
  class?: string;
}

// Non-interactive thumbnail of a structure SVG rendered with the manifest's
// colors — no pan/zoom, no hit-testing. Like CustomizerCanvas it injects svgText
// once on mount, so remount it (e.g. a keyed <Show>) when the artwork is
// replaced. The SVG scales to fill the host element, so give it a sized class.
export function SvgPreview(props: SvgPreviewProps) {
  let host!: HTMLDivElement;
  const [ready, setReady] = createSignal(false);
  const pieceEls = new Map<string, SVGElement>();
  const groutEls = new Map<string, SVGElement>();

  const glassById = createMemo<GlassById>(
    () => new Map(props.glassColors.map((g) => [g.id, g])),
  );
  const pieceSource = createMemo(() => buildPieceSourceMap(props.manifest));
  const groutHex = createMemo(() => {
    const id = props.manifest.grout_region.grout_id;
    if (id == null) return "#000000";
    return props.grouts.find((g) => g.id === id)?.hex ?? "#000000";
  });

  onMount(() => {
    host.innerHTML = props.svgText;
    const svg = host.querySelector("svg");
    if (svg) {
      svg.setAttribute("width", "100%");
      svg.setAttribute("height", "100%");
      svg.style.display = "block";
    }
    for (const id of pieceSource().keys()) {
      const el = host.querySelector<SVGElement>(`#${CSS.escape(id)}`);
      if (el) pieceEls.set(id, el);
    }
    for (const id of buildGroutPieceIds(props.manifest)) {
      const el = host.querySelector<SVGElement>(`#${CSS.escape(id)}`);
      if (el) groutEls.set(id, el);
    }
    setReady(true);
  });

  createEffect(() => {
    if (!ready()) return;
    for (const [id, groupKey] of pieceSource().entries()) {
      const el = pieceEls.get(id);
      if (el) {
        el.style.fill = resolvePieceHex(
          id,
          groupKey,
          {},
          props.manifest,
          glassById(),
        );
      }
    }
  });

  createEffect(() => {
    if (!ready()) return;
    for (const el of groutEls.values()) {
      el.style.fill = groutHex();
    }
  });

  onCleanup(() => {
    pieceEls.clear();
    groutEls.clear();
  });

  return <div ref={host} class={cn("h-full w-full", props.class)} />;
}
