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
      return { x: box.x, y: box.y, width: box.width, height: box.height };
    } catch {
      return null;
    } finally {
      container.innerHTML = "";
    }
  }

  return measure;
}
