import { createSignal } from "solid-js";

export interface PanZoomOptions {
  minScale?: number;
  maxScale?: number;
  // Multiplier applied per zoom step (wheel notch or button press).
  step?: number;
  // Pointer travel in px before a press is treated as a drag rather than a click.
  dragThreshold?: number;
}

export interface PanZoom {
  scale: () => number;
  tx: () => number;
  ty: () => number;
  /** CSS transform string; pair with `transform-origin: center`. */
  transform: () => string;
  /** The element wheel-zoom anchors the cursor against. */
  setViewport: (el: HTMLElement) => void;
  onWheel: (e: WheelEvent) => void;
  beginPan: (e: PointerEvent) => void;
  /** Returns true once the pointer has travelled past the drag threshold. */
  updatePan: (e: PointerEvent) => boolean;
  endPan: () => void;
  isPanning: () => boolean;
  /** True if the in-flight (or just-finished) press moved — use to suppress clicks. */
  didPan: () => boolean;
  zoomIn: () => void;
  zoomOut: () => void;
  reset: () => void;
}

/**
 * Pan and cursor-anchored wheel zoom over a transformed child element.
 *
 * Shared by the customizer canvas and the image lightbox so both behave
 * identically. It owns only the transform; hit-testing and rendering stay with
 * the caller.
 */
export function createPanZoom(options: PanZoomOptions = {}): PanZoom {
  const minScale = options.minScale ?? 0.4;
  const maxScale = options.maxScale ?? 8;
  const step = options.step ?? 1.12;
  const dragThreshold = options.dragThreshold ?? 4;

  const [scale, setScale] = createSignal(1);
  const [tx, setTx] = createSignal(0);
  const [ty, setTy] = createSignal(0);
  const [panning, setPanning] = createSignal(false);
  const [moved, setMoved] = createSignal(false);

  let viewport: HTMLElement | null = null;
  let startX = 0;
  let startY = 0;

  const clamp = (value: number) => Math.min(maxScale, Math.max(minScale, value));

  function zoomAround(newScale: number, mx: number, my: number) {
    const current = scale();
    setTx(mx - (mx - tx()) * (newScale / current));
    setTy(my - (my - ty()) * (newScale / current));
    setScale(newScale);
  }

  function zoomToCenter(factor: number) {
    setScale(clamp(scale() * factor));
  }

  return {
    scale,
    tx,
    ty,
    transform: () => `translate(${tx()}px, ${ty()}px) scale(${scale()})`,
    setViewport: (el) => {
      viewport = el;
    },
    onWheel(e) {
      e.preventDefault();
      const factor = e.deltaY < 0 ? step : 1 / step;
      const newScale = clamp(scale() * factor);
      if (!viewport) {
        setScale(newScale);
        return;
      }
      const rect = viewport.getBoundingClientRect();
      // Cursor position relative to the transform-origin (viewport center).
      zoomAround(
        newScale,
        e.clientX - rect.left - rect.width / 2,
        e.clientY - rect.top - rect.height / 2,
      );
    },
    beginPan(e) {
      setPanning(true);
      setMoved(false);
      startX = e.clientX;
      startY = e.clientY;
    },
    updatePan(e) {
      if (!panning()) return false;
      if (
        Math.abs(e.clientX - startX) + Math.abs(e.clientY - startY) >
        dragThreshold
      ) {
        setMoved(true);
      }
      if (moved()) {
        setTx(tx() + e.movementX);
        setTy(ty() + e.movementY);
      }
      return moved();
    },
    endPan() {
      setPanning(false);
    },
    isPanning: panning,
    didPan: moved,
    zoomIn: () => zoomToCenter(step),
    zoomOut: () => zoomToCenter(1 / step),
    reset() {
      setScale(1);
      setTx(0);
      setTy(0);
    },
  };
}
