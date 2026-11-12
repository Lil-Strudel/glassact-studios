import {
  For,
  Show,
  createEffect,
  createMemo,
  createSignal,
  createUniqueId,
  on,
  onCleanup,
  onMount,
} from "solid-js";
import { createStore } from "solid-js/store";
import { createPanZoom, cn } from "@glassact/ui";
import type {
  StoneBackgroundElement,
  StoneElement,
  StoneShape,
  StoneTextElement,
  StoneTextFinish,
} from "@glassact/data";
import { graniteMeanHex, graniteTexture, type GraniteChoice } from "../granite/granite";
import type { BackgroundArt } from "./background-art";
import type { StoneEditor } from "./create-stone-editor";
import { STONE_FONTS } from "./stone-fonts";
import {
  angleDegrees,
  boundsOf,
  boxCorners,
  cloneJson,
  normalizeRotation,
  rotatePoint,
  type LocalBox,
  type Point,
} from "./stone-geometry";
import { stoneOutlinePath } from "./stone-shapes";

export interface PlacedInlayArt {
  name: string;
  // Null while a custom inlay has no proof to show yet.
  imageUrl: string | null;
  width: number;
  height: number;
  mirrorX?: boolean;
  mirrorY?: boolean;
}

export interface StoneSpec {
  shape: StoneShape;
  width: number;
  height: number;
  granite: GraniteChoice;
}

interface StoneCanvasProps {
  // Null draws a plain artboard of `artboard` size instead of a stone.
  stone: StoneSpec | null;
  artboard?: { width: number; height: number };
  elements: StoneElement[];
  resolveInlay: (inlayUuid: string) => PlacedInlayArt | undefined;
  resolveBackground: (catalogItemUuid: string) => BackgroundArt | undefined;
  // Without an editor the canvas is a read-only picture.
  editor?: StoneEditor;
  isZoomable?: boolean;
  class?: string;
}

type Gesture =
  | { kind: "none" }
  | { kind: "pan" }
  | { kind: "tap" }
  | {
      kind: "move" | "rotate" | "resize";
      start: Point;
      center: Point;
      started: boolean;
      originals: Map<string, StoneElement>;
    };

const PLAIN_STONE_HEX = "#d4d4d4";
const hitTestContext = document.createElement("canvas").getContext("2d");
const DRAG_THRESHOLD_PX = 3;
const ROTATE_SNAP_DEGREES = 15;

export function StoneCanvas(props: StoneCanvasProps) {
  const uid = createUniqueId();
  const ids = {
    granite: `granite-${uid}`,
    clip: `clip-${uid}`,
    sheen: `sheen-${uid}`,
    shadow: `shadow-${uid}`,
    frost: `frost-${uid}`,
    engrave: `engrave-${uid}`,
    grid: `grid-${uid}`,
  };

  const panZoom = createPanZoom({ minScale: 0.5, maxScale: 10 });
  let viewport!: HTMLDivElement;
  let svg!: SVGSVGElement;

  const [measured, setMeasured] = createStore<Record<string, LocalBox>>({});
  const [svgPixelSize, setSvgPixelSize] = createSignal({ width: 1, height: 1 });
  const [gesture, setGesture] = createSignal<Gesture>({ kind: "none" });

  const surface = createMemo(() =>
    props.stone
      ? { width: props.stone.width, height: props.stone.height }
      : (props.artboard ?? { width: 24, height: 24 }),
  );

  const padding = createMemo(
    () => Math.max(surface().width, surface().height) * 0.08,
  );

  const viewBox = createMemo(() => {
    const p = padding();
    const s = surface();
    return { x: -p, y: -p, width: s.width + 2 * p, height: s.height + 2 * p };
  });

  // Inches per screen pixel, so handles stay a constant on-screen size.
  const unitsPerPx = createMemo(() => {
    const vb = viewBox();
    const px = svgPixelSize();
    return Math.max(vb.width / px.width, vb.height / px.height) / panZoom.scale();
  });

  const texture = createMemo(() =>
    props.stone ? graniteTexture(props.stone.granite) : null,
  );

  // Cut lettering reads light on dark stone and dark on light stone. 0.18 is
  // the luminance where white and black give equal contrast.
  const engravedFill = createMemo(() => {
    const hex = props.stone ? graniteMeanHex(props.stone.granite) : PLAIN_STONE_HEX;
    return relativeLuminance(hex) > 0.18
      ? { color: "#1f1f1f", opacity: 0.85 }
      : { color: "#ffffff", opacity: 0.85 };
  });

  function textLook(finish: StoneTextFinish): { fill: TextFill; filter: string } {
    switch (finish) {
      case "engraved":
        return { fill: engravedFill(), filter: ids.engrave };
      case "frosted":
        return { fill: FROSTED_FILL, filter: ids.frost };
      case "black-fill":
        return { fill: BLACK_FILL, filter: ids.engrave };
    }
  }

  const outlinePath = createMemo(() =>
    props.stone
      ? stoneOutlinePath(props.stone.shape, props.stone.width, props.stone.height)
      : "",
  );

  function localBox(el: StoneElement): LocalBox | undefined {
    switch (el.type) {
      case "inlay": {
        const art = props.resolveInlay(el.inlay_uuid);
        if (!art) return undefined;
        return { x: -art.width / 2, y: -art.height / 2, width: art.width, height: art.height };
      }
      case "background": {
        const art = props.resolveBackground(el.background_uuid);
        const height = el.width * (art?.aspect ?? 0.5);
        return { x: -el.width / 2, y: -height / 2, width: el.width, height };
      }
      case "text":
        return measured[el.id];
    }
  }

  const selectedIds = createMemo(() => new Set(props.editor?.selection() ?? []));

  // A Path2D hit test, unlike SVGGeometryElement.isPointInFill, does not
  // depend on the outline element having been rendered yet.
  const outline2d = createMemo(() => (props.stone ? new Path2D(outlinePath()) : null));

  const outOfBounds = createMemo(() => {
    const flagged = new Set<string>();
    const path = outline2d();
    if (!path || !hitTestContext) return flagged;
    for (const el of props.elements) {
      if (el.type === "background") continue;
      const box = localBox(el);
      if (!box) continue;
      const inside = boxCorners({ x: el.x, y: el.y }, el.rotation, box).every(
        (corner) => hitTestContext.isPointInPath(path, corner.x, corner.y),
      );
      if (!inside) flagged.add(el.id);
    }
    return flagged;
  });

  const selectionBounds = createMemo(() => {
    const selected = props.elements.filter((el) => selectedIds().has(el.id));
    const corners = selected.flatMap((el) => {
      const box = localBox(el);
      return box ? boxCorners({ x: el.x, y: el.y }, el.rotation, box) : [];
    });
    if (corners.length === 0) return null;
    return { ...boundsOf(corners), elements: selected };
  });

  // Inlays are drawn at the size they will be made, so a selection holding
  // one cannot be scaled.
  const isResizable = createMemo(() => {
    const bounds = selectionBounds();
    return !!bounds && bounds.elements.every((el) => el.type === "text" || el.type === "background");
  });

  onMount(() => {
    panZoom.setViewport(viewport);
    const observer = new ResizeObserver(() => {
      setSvgPixelSize({
        width: svg.clientWidth || 1,
        height: svg.clientHeight || 1,
      });
    });
    observer.observe(svg);
    onCleanup(() => observer.disconnect());
  });

  function toSvgPoint(e: PointerEvent): Point {
    const matrix = svg.getScreenCTM();
    if (!matrix) return { x: 0, y: 0 };
    const p = new DOMPoint(e.clientX, e.clientY).matrixTransform(matrix.inverse());
    return { x: p.x, y: p.y };
  }

  function startElementGesture(
    e: PointerEvent,
    kind: "move" | "rotate" | "resize",
    center: Point,
  ) {
    const editor = props.editor;
    if (!editor) return;
    const originals = new Map(
      editor.selectedElements().map((el) => [el.id, cloneJson(el)]),
    );
    svg.setPointerCapture(e.pointerId);
    setGesture({ kind, start: toSvgPoint(e), center, started: false, originals });
  }

  function handleElementPointerDown(e: PointerEvent, el: StoneElement) {
    const editor = props.editor;
    if (!editor || e.button !== 0) return;
    e.stopPropagation();
    if (!selectedIds().has(el.id) || e.shiftKey) {
      editor.select(el.id, e.shiftKey);
    }
    startElementGesture(e, "move", { x: el.x, y: el.y });
  }

  function handleHandlePointerDown(e: PointerEvent, kind: "rotate" | "resize") {
    const bounds = selectionBounds();
    if (!bounds || e.button !== 0) return;
    e.stopPropagation();
    const only = bounds.elements.length === 1 ? bounds.elements[0] : null;
    const center = kind === "resize" && only ? { x: only.x, y: only.y } : bounds.center;
    startElementGesture(e, kind, center);
  }

  function handleSurfacePointerDown(e: PointerEvent) {
    if (e.button !== 0 && e.button !== 1) return;
    if (props.isZoomable) {
      svg.setPointerCapture(e.pointerId);
      panZoom.beginPan(e);
      setGesture({ kind: "pan" });
    } else if (props.editor) {
      setGesture({ kind: "tap" });
    }
  }

  function handlePointerMove(e: PointerEvent) {
    const current = gesture();
    if (current.kind === "none" || current.kind === "tap") return;
    if (current.kind === "pan") {
      panZoom.updatePan(e);
      return;
    }

    const editor = props.editor!;
    const point = toSvgPoint(e);
    if (!current.started) {
      const travelled = Math.hypot(point.x - current.start.x, point.y - current.start.y);
      if (travelled / unitsPerPx() < DRAG_THRESHOLD_PX) return;
      editor.beginGesture();
      setGesture({ ...current, started: true });
    }

    const scale = current.kind === "resize" ? resizeScale(current, point) : 1;
    editor.gesture((draft) => {
      for (const el of draft.layout.elements) {
        const original = current.originals.get(el.id);
        if (!original) continue;
        applyGesture(el, original, current, point, e.shiftKey, scale);
      }
    });
  }

  function handlePointerUp(e: PointerEvent) {
    const current = gesture();
    if (svg.hasPointerCapture(e.pointerId)) svg.releasePointerCapture(e.pointerId);
    if (current.kind === "pan") {
      const didPan = panZoom.didPan();
      panZoom.endPan();
      if (!didPan) props.editor?.clearSelection();
    }
    if (current.kind === "tap") props.editor?.clearSelection();
    setGesture({ kind: "none" });
  }

  function handleWheel(e: WheelEvent) {
    if (!props.isZoomable) return;
    panZoom.onWheel(e);
  }

  function reportMeasure(id: string, box: LocalBox) {
    setMeasured(id, box);
  }

  return (
    <div
      ref={viewport}
      class={cn("relative overflow-hidden select-none touch-none", props.class)}
      onWheel={handleWheel}
    >
      <div
        class="h-full w-full"
        style={{ transform: panZoom.transform(), "transform-origin": "center" }}
      >
        <svg
          ref={svg}
          class="h-full w-full"
          viewBox={`${viewBox().x} ${viewBox().y} ${viewBox().width} ${viewBox().height}`}
          preserveAspectRatio="xMidYMid meet"
          onPointerDown={handleSurfacePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
          onPointerCancel={handlePointerUp}
        >
          <defs>
            <style>
              {".stone-frost-art * { fill: #fff !important; stroke: none !important; }"}
            </style>
            <Show when={texture()}>
              {(tex) => (
                <pattern
                  id={ids.granite}
                  patternUnits="userSpaceOnUse"
                  width={tex().tileInches}
                  height={tex().tileInches}
                >
                  <rect width={tex().tileInches} height={tex().tileInches} fill={tex().baseHex} />
                  <image
                    href={tex().image}
                    width={tex().tileInches}
                    height={tex().tileInches}
                    preserveAspectRatio="none"
                    style={tex().isTinted ? { "mix-blend-mode": "overlay" } : undefined}
                  />
                </pattern>
              )}
            </Show>
            <pattern id={ids.grid} patternUnits="userSpaceOnUse" width="1" height="1">
              <path d="M1 0 L0 0 0 1" fill="none" stroke="#d1d5db" stroke-width="0.02" />
            </pattern>
            <clipPath id={ids.clip}>
              <path d={outlinePath()} />
            </clipPath>
            <linearGradient id={ids.sheen} x1="0" y1="0" x2="1" y2="1">
              <stop offset="0" stop-color="#fff" stop-opacity="0.16" />
              <stop offset="0.45" stop-color="#fff" stop-opacity="0" />
              <stop offset="1" stop-color="#000" stop-opacity="0.12" />
            </linearGradient>
            <filter id={ids.shadow} x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation={surface().width * 0.012} />
            </filter>
            <filter id={ids.frost} x="-2%" y="-2%" width="104%" height="104%">
              <feTurbulence type="fractalNoise" baseFrequency="14" numOctaves="2" seed="7" result="noise" />
              <feColorMatrix
                in="noise"
                type="matrix"
                values="0 0 0 0 1  0 0 0 0 1  0 0 0 0 1  0 0 0 -1.1 1.35"
                result="grain"
              />
              <feComposite in="SourceGraphic" in2="grain" operator="in" result="frosted" />
              <feOffset in="SourceAlpha" dx="0.015" dy="0.025" result="depthOffset" />
              <feComposite in="depthOffset" in2="SourceAlpha" operator="out" result="depthEdge" />
              <feFlood flood-color="#000" flood-opacity="0.35" />
              <feComposite in2="depthEdge" operator="in" result="depth" />
              <feMerge>
                <feMergeNode in="frosted" />
                <feMergeNode in="depth" />
              </feMerge>
            </filter>
            <filter id={ids.engrave} x="-5%" y="-10%" width="110%" height="120%">
              <feComponentTransfer in="SourceAlpha" result="inverse">
                <feFuncA type="table" tableValues="1 0" />
              </feComponentTransfer>
              <feOffset in="inverse" dx="0.035" dy="0.05" result="inverseOffset" />
              <feGaussianBlur in="inverseOffset" stdDeviation="0.02" result="inverseBlur" />
              <feFlood flood-color="#000" flood-opacity="0.8" />
              <feComposite in2="inverseBlur" operator="in" result="shade" />
              <feComposite in="shade" in2="SourceAlpha" operator="in" result="innerShadow" />
              <feMerge>
                <feMergeNode in="SourceGraphic" />
                <feMergeNode in="innerShadow" />
              </feMerge>
            </filter>
          </defs>

          <Show
            when={props.stone}
            fallback={
              <rect
                width={surface().width}
                height={surface().height}
                fill={`url(#${ids.grid})`}
                stroke="#9ca3af"
                stroke-width="0.04"
                stroke-dasharray="0.2 0.15"
              />
            }
          >
            <path
              d={outlinePath()}
              fill="#000"
              opacity="0.35"
              transform={`translate(${surface().width * 0.01} ${surface().width * 0.015})`}
              filter={`url(#${ids.shadow})`}
            />
            <path
              d={outlinePath()}
              fill={texture() ? `url(#${ids.granite})` : PLAIN_STONE_HEX}
            />
            <path d={outlinePath()} fill={`url(#${ids.sheen})`} />
            <path d={outlinePath()} fill="none" stroke="#000" stroke-opacity="0.35" stroke-width="0.12" />
          </Show>

          <g clip-path={props.stone ? `url(#${ids.clip})` : undefined}>
            <For
              each={props.elements.filter(
                (el): el is StoneBackgroundElement => el.type === "background",
              )}
            >
              {(el) => (
                <BackgroundView
                  el={el}
                  art={props.resolveBackground(el.background_uuid)}
                  frostFilter={ids.frost}
                  onPointerDown={(e) => handleElementPointerDown(e, el)}
                />
              )}
            </For>
          </g>

          <For each={props.elements.filter((el) => el.type !== "background")}>
            {(el) =>
              el.type === "inlay" ? (
                <InlayView
                  el={el}
                  art={props.resolveInlay(el.inlay_uuid)}
                  onPointerDown={(e) => handleElementPointerDown(e, el)}
                />
              ) : el.type === "text" ? (
                <TextView
                  el={el}
                  fill={textLook(el.finish).fill}
                  filter={textLook(el.finish).filter}
                  onMeasure={reportMeasure}
                  onPointerDown={(e) => handleElementPointerDown(e, el)}
                />
              ) : null
            }
          </For>

          <For each={props.elements}>
            {(el) => (
              <Show when={(selectedIds().has(el.id) || outOfBounds().has(el.id)) && localBox(el)}>
                {(box) => (
                  <rect
                    transform={`translate(${el.x} ${el.y}) rotate(${el.rotation})`}
                    x={box().x}
                    y={box().y}
                    width={box().width}
                    height={box().height}
                    fill="none"
                    stroke={outOfBounds().has(el.id) ? "#dc2626" : "#2563eb"}
                    stroke-dasharray={outOfBounds().has(el.id) ? "4 3" : undefined}
                    stroke-width="1.5"
                    vector-effect="non-scaling-stroke"
                    pointer-events="none"
                  />
                )}
              </Show>
            )}
          </For>

          <Show when={props.editor && selectionBounds()}>
            {(bounds) => {
              const handleRadius = () => unitsPerPx() * 7;
              const rotateHandle = () => ({
                x: bounds().center.x,
                y: bounds().y - unitsPerPx() * 28,
              });
              return (
                <g>
                  <rect
                    x={bounds().x}
                    y={bounds().y}
                    width={bounds().width}
                    height={bounds().height}
                    fill="none"
                    stroke="#2563eb"
                    stroke-opacity="0.5"
                    stroke-width="1"
                    vector-effect="non-scaling-stroke"
                    pointer-events="none"
                  />
                  <line
                    x1={bounds().center.x}
                    y1={bounds().y}
                    x2={rotateHandle().x}
                    y2={rotateHandle().y}
                    stroke="#2563eb"
                    stroke-width="1"
                    vector-effect="non-scaling-stroke"
                  />
                  <circle
                    cx={rotateHandle().x}
                    cy={rotateHandle().y}
                    r={handleRadius()}
                    fill="#fff"
                    stroke="#2563eb"
                    stroke-width="1.5"
                    vector-effect="non-scaling-stroke"
                    class="cursor-grab"
                    onPointerDown={(e) => handleHandlePointerDown(e, "rotate")}
                  />
                  <Show when={isResizable()}>
                    <rect
                      x={bounds().x + bounds().width - handleRadius()}
                      y={bounds().y + bounds().height - handleRadius()}
                      width={handleRadius() * 2}
                      height={handleRadius() * 2}
                      fill="#2563eb"
                      class="cursor-nwse-resize"
                      onPointerDown={(e) => handleHandlePointerDown(e, "resize")}
                    />
                  </Show>
                </g>
              );
            }}
          </Show>
        </svg>
      </div>

      <Show when={props.isZoomable}>
        <div class="absolute bottom-3 right-3 flex overflow-hidden rounded-md border border-gray-300 bg-white/90 text-sm shadow-sm">
          <button type="button" class="px-3 py-1 hover:bg-gray-100" onClick={panZoom.zoomOut} aria-label="Zoom out">
            −
          </button>
          <button type="button" class="border-x border-gray-300 px-3 py-1 hover:bg-gray-100" onClick={panZoom.reset}>
            Fit
          </button>
          <button type="button" class="px-3 py-1 hover:bg-gray-100" onClick={panZoom.zoomIn} aria-label="Zoom in">
            +
          </button>
        </div>
      </Show>
    </div>
  );
}

type TextFill = { color: string; opacity: number };

const FROSTED_FILL: TextFill = { color: "#ffffff", opacity: 0.92 };
const BLACK_FILL: TextFill = { color: "#0a0a0a", opacity: 1 };

function relativeLuminance(hex: string): number {
  const channel = (offset: number) => {
    const c = parseInt(hex.slice(offset, offset + 2), 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  };
  return 0.2126 * channel(1) + 0.7152 * channel(3) + 0.0722 * channel(5);
}

function applyGesture(
  el: StoneElement,
  original: StoneElement,
  gesture: Extract<Gesture, { originals: Map<string, StoneElement> }>,
  point: Point,
  isSnapping: boolean,
  scale: number,
) {
  switch (gesture.kind) {
    case "move":
      el.x = original.x + (point.x - gesture.start.x);
      el.y = original.y + (point.y - gesture.start.y);
      return;
    case "rotate": {
      let delta =
        angleDegrees(gesture.center, point) - angleDegrees(gesture.center, gesture.start);
      if (isSnapping) delta = Math.round(delta / ROTATE_SNAP_DEGREES) * ROTATE_SNAP_DEGREES;
      const moved = rotatePoint({ x: original.x, y: original.y }, gesture.center, delta);
      el.x = moved.x;
      el.y = moved.y;
      el.rotation = normalizeRotation(original.rotation + delta);
      return;
    }
    case "resize": {
      el.x = gesture.center.x + (original.x - gesture.center.x) * scale;
      el.y = gesture.center.y + (original.y - gesture.center.y) * scale;
      if (el.type === "text" && original.type === "text") {
        el.size = Number((original.size * scale).toFixed(3));
      } else if (el.type === "background" && original.type === "background") {
        el.width = Number((original.width * scale).toFixed(3));
      }
      return;
    }
  }
}

const TEXT_SIZE_LIMITS = { min: 0.1, max: 24 };
const BACKGROUND_WIDTH_LIMITS = { min: 0.25, max: 240 };

// One factor for the whole selection, clamped so the element nearest its
// limit stops everything; clamping each element alone would distort the
// arrangement.
function resizeScale(
  gesture: Extract<Gesture, { originals: Map<string, StoneElement> }>,
  point: Point,
): number {
  const startDistance = Math.hypot(
    gesture.start.x - gesture.center.x,
    gesture.start.y - gesture.center.y,
  );
  if (startDistance === 0) return 1;
  const scale = Math.hypot(point.x - gesture.center.x, point.y - gesture.center.y) / startDistance;
  let lowest = 0;
  let highest = Infinity;
  for (const original of gesture.originals.values()) {
    const [value, limits] =
      original.type === "text"
        ? [original.size, TEXT_SIZE_LIMITS]
        : original.type === "background"
          ? [original.width, BACKGROUND_WIDTH_LIMITS]
          : [null, null];
    if (value === null || value === 0) continue;
    lowest = Math.max(lowest, limits.min / value);
    highest = Math.min(highest, limits.max / value);
  }
  return Math.min(highest, Math.max(lowest, scale));
}

function InlayView(props: {
  el: StoneElement;
  art: PlacedInlayArt | undefined;
  onPointerDown: (e: PointerEvent) => void;
}) {
  return (
    <Show when={props.art}>
      {(art) => (
        <g
          transform={`translate(${props.el.x} ${props.el.y}) rotate(${props.el.rotation})`}
          class="cursor-move"
          onPointerDown={(e) => props.onPointerDown(e)}
        >
          <Show
            when={art().imageUrl}
            fallback={
              <g>
                <rect
                  x={-art().width / 2}
                  y={-art().height / 2}
                  width={art().width}
                  height={art().height}
                  fill="#ffffff"
                  fill-opacity="0.75"
                  stroke="#6b7280"
                  stroke-width="0.05"
                  stroke-dasharray="0.25 0.15"
                />
                <text
                  text-anchor="middle"
                  font-size={String(Math.min(art().width, art().height) * 0.12)}
                  fill="#374151"
                  font-family="sans-serif"
                >
                  <tspan x="0" dy="-0.2em">{art().name}</tspan>
                  <tspan x="0" dy="1.3em">Awaiting proof</tspan>
                </text>
              </g>
            }
          >
            {(url) => (
              <image
                href={url()}
                x={-art().width / 2}
                y={-art().height / 2}
                width={art().width}
                height={art().height}
                preserveAspectRatio="xMidYMid meet"
                transform={`scale(${art().mirrorX ? -1 : 1} ${art().mirrorY ? -1 : 1})`}
              />
            )}
          </Show>
        </g>
      )}
    </Show>
  );
}

function BackgroundView(props: {
  el: StoneBackgroundElement;
  art: BackgroundArt | undefined;
  frostFilter: string;
  onPointerDown: (e: PointerEvent) => void;
}) {
  const height = () => props.el.width * (props.art?.aspect ?? 0.5);
  return (
    <Show when={props.art}>
      {(art) => (
        <g
          transform={`translate(${props.el.x} ${props.el.y}) rotate(${props.el.rotation}) scale(${props.el.flip_x ? -1 : 1} 1)`}
          class="cursor-move"
          onPointerDown={(e) => props.onPointerDown(e)}
        >
          <g filter={`url(#${props.frostFilter})`} opacity="0.62">
            <svg
              x={-props.el.width / 2}
              y={-height() / 2}
              width={props.el.width}
              height={height()}
              viewBox={art().viewBox}
              preserveAspectRatio="xMidYMid meet"
              class="stone-frost-art"
              // eslint-disable-next-line solid/no-innerhtml -- sanitized by parseBackgroundArt
              innerHTML={art().markup}
            />
          </g>
        </g>
      )}
    </Show>
  );
}

function TextView(props: {
  el: StoneTextElement;
  fill: TextFill;
  filter: string;
  onMeasure: (id: string, box: LocalBox) => void;
  onPointerDown: (e: PointerEvent) => void;
}) {
  let textRef: SVGTextElement | undefined;
  const [shift, setShift] = createSignal({ x: 0, y: 0 });
  const [box, setBox] = createSignal<LocalBox | null>(null);

  const font = () => STONE_FONTS[props.el.font];
  const fontSize = () => props.el.size / font().capHeight;
  const lines = () => props.el.text.split("\n");
  const anchor = () =>
    props.el.align === "left" ? "start" : props.el.align === "right" ? "end" : "middle";

  function measure() {
    if (!textRef) return;
    const b = textRef.getBBox();
    const centered: LocalBox = {
      x: -b.width / 2,
      y: -b.height / 2,
      width: b.width,
      height: b.height,
    };
    setShift({ x: -(b.x + b.width / 2), y: -(b.y + b.height / 2) });
    setBox(centered);
    props.onMeasure(props.el.id, centered);
  }

  createEffect(
    on(
      () => [props.el.text, props.el.size, props.el.font, props.el.align],
      () => queueMicrotask(measure),
    ),
  );
  onMount(() => {
    document.fonts.ready.then(measure);
  });

  return (
    <g
      transform={`translate(${props.el.x} ${props.el.y}) rotate(${props.el.rotation})`}
      class="cursor-move"
      onPointerDown={(e) => props.onPointerDown(e)}
    >
      <Show when={box()}>
        {(b) => (
          <rect x={b().x} y={b().y} width={b().width} height={b().height} fill="transparent" />
        )}
      </Show>
      <g transform={`translate(${shift().x} ${shift().y})`}>
        <text
          ref={textRef}
          font-family={font().family}
          font-size={String(fontSize())}
          text-anchor={anchor()}
          fill={props.fill.color}
          fill-opacity={props.fill.opacity}
          filter={`url(#${props.filter})`}
        >
          <For each={lines()}>
            {(line, i) => (
              <tspan x="0" dy={i() === 0 ? 0 : fontSize() * 1.2}>
                {line || " "}
              </tspan>
            )}
          </For>
        </text>
      </g>
    </g>
  );
}
