import { onCleanup, onMount } from "solid-js";
import type { ContentBBox } from "@glassact/data";

// Measures the artwork's content bounding box from the structure SVG. The SVG is
// rendered in an attached but offscreen container (NOT display:none, which yields
// an empty getBBox()) and measured after mount.
//
// - Freshly analyzed SVG: measure the root <svg> getBBox() (all drawn content).
// - Stored/edited item whose SVG has been baked with a <g id="gac-fit"> wrapper:
//   measure that inner group in its local pre-transform coordinate space.
//
// Returns a getter that lazily renders + measures the provided SVG text on demand
// so callers can re-measure at save time against the latest structure SVG.

// The source artboard that ingest clips the artwork to (artboardClipID in
// apps/api/svg/normalize.go). getBBox() ignores clipping, so a stray shape off
// the artboard would otherwise widen the box the server fits the design into.
const ARTBOARD_CLIP_ID = "gac-artboard";

function intersect(a: DOMRect, b: DOMRect): ContentBBox | null {
  const x = Math.max(a.x, b.x);
  const y = Math.max(a.y, b.y);
  const width = Math.min(a.x + a.width, b.x + b.width) - x;
  const height = Math.min(a.y + a.height, b.y + b.height) - y;
  if (width <= 0 || height <= 0) return null;
  return { x, y, width, height };
}

export function useContentBBox() {
  let container: HTMLDivElement | null = null;

  onMount(() => {
    container = document.createElement("div");
    container.style.position = "absolute";
    container.style.left = "-99999px";
    container.style.top = "0";
    container.style.width = "1000px";
    container.style.height = "1000px";
    container.style.pointerEvents = "none";
    document.body.appendChild(container);
  });

  onCleanup(() => {
    if (container && container.parentNode) {
      container.parentNode.removeChild(container);
    }
    container = null;
  });

  function measure(svgText: string): ContentBBox | null {
    if (!container || !svgText) return null;
    container.innerHTML = svgText;
    const svg = container.querySelector("svg");
    if (!svg) return null;

    const fitGroup = svg.querySelector<SVGGraphicsElement>("#gac-fit");
    const target: SVGGraphicsElement = fitGroup ?? (svg as SVGGraphicsElement);

    try {
      const box = target.getBBox();
      if (box.width === 0 && box.height === 0) return null;
      const artboard = svg.querySelector<SVGRectElement>(
        `#${ARTBOARD_CLIP_ID} rect`,
      );
      return artboard
        ? intersect(box, artboard.getBBox())
        : { x: box.x, y: box.y, width: box.width, height: box.height };
    } catch {
      return null;
    } finally {
      container.innerHTML = "";
    }
  }

  return measure;
}
