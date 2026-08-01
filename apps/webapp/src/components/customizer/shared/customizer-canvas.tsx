import { createEffect, createSignal, onCleanup, onMount, Show } from "solid-js";
import { Button, createPanZoom } from "@glassact/ui";
import {
  graniteBackgroundStyle,
  type GranitePreset,
} from "../../granite/granite";
import { GranitePill } from "../../granite/granite-pill";

interface CustomizerCanvasProps {
  svgText: string;
  // pieceId -> group key. The key is opaque to the canvas; it is only echoed
  // back through the hover/click callbacks and compared for region highlighting.
  pieceSource: Map<string, string>;
  groutPieceIds: string[];
  // Resolves the current fill for a piece. Reads reactive state, so calling it
  // inside an effect re-applies fills whenever overrides/hover-preview change.
  resolveHex: (pieceId: string, groupKey: string) => string;
  groutHex: string | null;
  selectedPieceId: string | null;
  // Group key whose pieces should be region-highlighted (e.g. on hover).
  highlightedRegion: string | null;
  // Granite backdrop the inlay is previewed against. Optional: when omitted (e.g.
  // the admin manifest editor) the canvas shows a plain backdrop and no picker.
  granite?: GranitePreset;
  graniteKey?: string;
  onSelectGranite?: (key: string) => void;
  onPieceClick: (pieceId: string, groupKey: string) => void;
  onPieceHover: (pieceId: string | null, groupKey: string | null) => void;
  // Fired when the user clicks empty backdrop (not a piece, not a pan).
  onDeselect?: () => void;
}

const HIGHLIGHT_CSS = `
.gac-canvas [id^="p"]:not([data-grout]) { cursor: pointer; }
.gac-canvas .gac-hover { stroke: white; stroke-width: 4px; paint-order: stroke fill; vector-effect: non-scaling-stroke; filter: drop-shadow(0 0 2px #60a5fa); }
.gac-canvas .gac-selected { stroke: white; stroke-width: 5px; paint-order: stroke fill; vector-effect: non-scaling-stroke; filter: drop-shadow(0 0 3px #2563eb); }
`;

export function CustomizerCanvas(props: CustomizerCanvasProps) {
  let host!: HTMLDivElement;
  const [ready, setReady] = createSignal(false);
  const pieceEls = new Map<string, SVGElement>();
  const groutEls = new Map<string, SVGElement>();

  const panZoom = createPanZoom();

  let pressedPiece: string | null = null;

  onMount(() => {
    host.innerHTML = props.svgText;
    const svg = host.querySelector("svg");
    if (svg) {
      svg.setAttribute("width", "100%");
      svg.setAttribute("height", "100%");
      svg.style.display = "block";
    }
    pieceEls.clear();
    for (const id of props.pieceSource.keys()) {
      const el = host.querySelector<SVGElement>(`#${CSS.escape(id)}`);
      if (el) pieceEls.set(id, el);
    }
    groutEls.clear();
    for (const id of props.groutPieceIds) {
      const el = host.querySelector<SVGElement>(`#${CSS.escape(id)}`);
      if (el) {
        el.setAttribute("data-grout", "");
        groutEls.set(id, el);
      }
    }
    setReady(true);
  });

  // Apply resolved fills whenever overrides / hover-preview change.
  createEffect(() => {
    if (!ready()) return;
    for (const [id, groupKey] of props.pieceSource.entries()) {
      const el = pieceEls.get(id);
      if (el) el.style.fill = props.resolveHex(id, groupKey);
    }
  });

  // Apply grout color to grout shapes (the black back-shapes in the SVG).
  createEffect(() => {
    if (!ready()) return;
    const hex = props.groutHex ?? "#000000";
    for (const el of groutEls.values()) {
      el.style.fill = hex;
    }
  });

  // Selection + region highlight.
  createEffect(() => {
    if (!ready()) return;
    const selected = props.selectedPieceId;
    const region = props.highlightedRegion;
    for (const [id, groupKey] of props.pieceSource.entries()) {
      const el = pieceEls.get(id);
      if (!el) continue;
      el.classList.toggle("gac-selected", selected === id);
      el.classList.toggle(
        "gac-hover",
        selected !== id && region !== null && groupKey === region,
      );
    }
  });

  function pieceAt(target: EventTarget | null): string | null {
    const el = target as Element | null;
    if (!el || !el.id) return null;
    return props.pieceSource.has(el.id) ? el.id : null;
  }

  function onPointerDown(e: PointerEvent) {
    panZoom.beginPan(e);
    pressedPiece = pieceAt(e.target);
    host.setPointerCapture(e.pointerId);
  }

  function onPointerMove(e: PointerEvent) {
    if (panZoom.isPanning()) {
      panZoom.updatePan(e);
      return;
    }
    const id = pieceAt(e.target);
    props.onPieceHover(id, id ? props.pieceSource.get(id)! : null);
  }

  function onPointerUp(e: PointerEvent) {
    // A press that never moved is a click, not a pan: on a piece it selects it,
    // on empty backdrop it deselects.
    if (panZoom.isPanning() && !panZoom.didPan()) {
      if (pressedPiece) {
        props.onPieceClick(pressedPiece, props.pieceSource.get(pressedPiece)!);
      } else {
        props.onDeselect?.();
      }
    }
    panZoom.endPan();
    pressedPiece = null;
    host.releasePointerCapture?.(e.pointerId);
  }

  onCleanup(() => { pieceEls.clear(); groutEls.clear(); });

  return (
    <div class="relative flex h-full w-full flex-col">
      <style>{HIGHLIGHT_CSS}</style>
      <div
        ref={panZoom.setViewport}
        class="gac-canvas relative flex-1 overflow-hidden rounded-lg border border-gray-200"
        style={
          props.granite
            ? graniteBackgroundStyle(props.granite)
            : { "background-color": "#f3f4f6" }
        }
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerLeave={() => props.onPieceHover(null, null)}
        onWheel={(e) => panZoom.onWheel(e)}
      >
        <div
          class="absolute inset-0 flex items-center justify-center p-6"
          style={{
            transform: panZoom.transform(),
            "transform-origin": "center",
          }}
        >
          <div ref={host} class="h-full w-full" />
        </div>
      </div>

      <Show when={props.granite && props.onSelectGranite}>
        <GranitePill
          class="absolute left-3 top-3"
          selectedKey={props.graniteKey ?? ""}
          onSelect={(key) => props.onSelectGranite?.(key)}
        />
      </Show>

      <div class="pointer-events-none absolute bottom-3 left-1/2 flex -translate-x-1/2 items-center gap-1 rounded-full border border-gray-200 bg-white/90 px-2 py-1 shadow-sm">
        <Button
          variant="ghost"
          size="icon"
          class="pointer-events-auto h-7 w-7 rounded-full text-lg leading-none"
          onClick={panZoom.zoomOut}
          aria-label="Zoom out"
        >
          −
        </Button>
        <span class="w-12 text-center text-xs tabular-nums text-gray-500">
          {Math.round(panZoom.scale() * 100)}%
        </span>
        <Button
          variant="ghost"
          size="icon"
          class="pointer-events-auto h-7 w-7 rounded-full text-lg leading-none"
          onClick={panZoom.zoomIn}
          aria-label="Zoom in"
        >
          +
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="pointer-events-auto ml-1 rounded-full px-2 py-0.5 text-xs"
          onClick={panZoom.reset}
        >
          Reset
        </Button>
      </div>
    </div>
  );
}

export type { CustomizerCanvasProps };
