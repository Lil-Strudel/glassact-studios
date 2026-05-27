import { createEffect, createSignal, onCleanup, onMount } from "solid-js";
import { Button } from "@glassact/ui";

interface CustomizerCanvasProps {
  svgText: string;
  pieceSource: Map<string, string>;
  groutPieceIds: string[];
  // Resolves the current fill for a piece. Reads reactive state, so calling it
  // inside an effect re-applies fills whenever overrides/hover-preview change.
  resolveHex: (pieceId: string, sourceHex: string) => string;
  groutHex: string | null;
  selectedPieceId: string | null;
  highlightedRegion: string | null;
  onPieceClick: (pieceId: string, sourceHex: string) => void;
  onPieceHover: (pieceId: string | null, sourceHex: string | null) => void;
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

  const [scale, setScale] = createSignal(1);
  const [tx, setTx] = createSignal(0);
  const [ty, setTy] = createSignal(0);

  let dragging = false;
  let moved = false;
  let startX = 0;
  let startY = 0;
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
    for (const [id, sourceHex] of props.pieceSource.entries()) {
      const el = pieceEls.get(id);
      if (el) el.style.fill = props.resolveHex(id, sourceHex);
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
    for (const [id, sourceHex] of props.pieceSource.entries()) {
      const el = pieceEls.get(id);
      if (!el) continue;
      el.classList.toggle("gac-selected", selected === id);
      el.classList.toggle(
        "gac-hover",
        selected !== id && region !== null && sourceHex === region,
      );
    }
  });

  const transform = () =>
    `translate(${tx()}px, ${ty()}px) scale(${scale()})`;

  function pieceAt(target: EventTarget | null): string | null {
    const el = target as Element | null;
    if (!el || !el.id) return null;
    return props.pieceSource.has(el.id) ? el.id : null;
  }

  function onPointerDown(e: PointerEvent) {
    dragging = true;
    moved = false;
    startX = e.clientX;
    startY = e.clientY;
    pressedPiece = pieceAt(e.target);
    host.setPointerCapture(e.pointerId);
  }

  function onPointerMove(e: PointerEvent) {
    if (dragging) {
      if (Math.abs(e.clientX - startX) + Math.abs(e.clientY - startY) > 4) {
        moved = true;
      }
      if (moved) {
        setTx(tx() + e.movementX);
        setTy(ty() + e.movementY);
      }
      return;
    }
    const id = pieceAt(e.target);
    props.onPieceHover(id, id ? props.pieceSource.get(id)! : null);
  }

  function onPointerUp(e: PointerEvent) {
    if (dragging && !moved && pressedPiece) {
      props.onPieceClick(pressedPiece, props.pieceSource.get(pressedPiece)!);
    }
    dragging = false;
    pressedPiece = null;
    host.releasePointerCapture?.(e.pointerId);
  }

  function onWheel(e: WheelEvent) {
    e.preventDefault();
    const factor = e.deltaY < 0 ? 1.12 : 0.89;
    setScale(Math.min(8, Math.max(0.4, scale() * factor)));
  }

  function reset() {
    setScale(1);
    setTx(0);
    setTy(0);
  }

  onCleanup(() => { pieceEls.clear(); groutEls.clear(); });

  return (
    <div class="relative flex h-full w-full flex-col">
      <style>{HIGHLIGHT_CSS}</style>
      <div
        class="gac-canvas relative flex-1 overflow-hidden rounded-lg border border-gray-200"
        style={{ "background-color": "#f3f4f6" }}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerLeave={() => props.onPieceHover(null, null)}
        onWheel={onWheel}
      >
        <div
          class="absolute inset-0 flex items-center justify-center p-6"
          style={{ transform: transform(), "transform-origin": "center" }}
        >
          <div ref={host} class="h-full w-full" />
        </div>
      </div>

      <div class="pointer-events-none absolute bottom-3 left-1/2 flex -translate-x-1/2 items-center gap-1 rounded-full border border-gray-200 bg-white/90 px-2 py-1 shadow-sm">
        <Button
          variant="ghost"
          size="icon"
          class="pointer-events-auto h-7 w-7 rounded-full text-lg leading-none"
          onClick={() => setScale(Math.max(0.4, scale() * 0.89))}
          aria-label="Zoom out"
        >
          −
        </Button>
        <span class="w-12 text-center text-xs tabular-nums text-gray-500">
          {Math.round(scale() * 100)}%
        </span>
        <Button
          variant="ghost"
          size="icon"
          class="pointer-events-auto h-7 w-7 rounded-full text-lg leading-none"
          onClick={() => setScale(Math.min(8, scale() * 1.12))}
          aria-label="Zoom in"
        >
          +
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="pointer-events-auto ml-1 rounded-full px-2 py-0.5 text-xs"
          onClick={reset}
        >
          Reset
        </Button>
      </div>
    </div>
  );
}
