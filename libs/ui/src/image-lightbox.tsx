import { Dialog as DialogPrimitive } from "@kobalte/core/dialog";
import {
  For,
  Show,
  children,
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  untrack,
  type JSX,
} from "solid-js";
import { cn } from "./cn";
import { createPanZoom } from "./create-pan-zoom";

export interface LightboxImage {
  src: string;
  alt: string;
}

export interface ImageLightboxProps {
  /** Images in the set. A single-entry array hides the gallery navigation. */
  images: LightboxImage[];
  /** Index to show when opened. */
  index?: number;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  /** Rendered as the clickable trigger. Omit when driving `open` yourself. */
  children?: JSX.Element;
  /** Class applied to the trigger wrapper. */
  triggerClass?: string;
  /**
   * Background for the viewer surface, replacing the default dark backdrop —
   * e.g. the granite slab an inlay is previewed against.
   */
  backdropStyle?: JSX.CSSProperties;
  /** Overlaid at the top-left of the viewer, e.g. a backdrop picker. */
  controls?: JSX.Element;
}

/**
 * Fullscreen image viewer with pan/zoom and optional gallery navigation.
 *
 * Built on the Kobalte dialog primitive rather than the shared DialogContent,
 * which hardcodes a centered `max-w-lg` panel and a light overlay — neither of
 * which suits an edge-to-edge dark lightbox.
 */
export function ImageLightbox(props: ImageLightboxProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = createSignal(false);
  const [current, setCurrent] = createSignal(
    untrack(() => props.index ?? 0),
  );

  const isOpen = () => props.open ?? uncontrolledOpen();

  function setOpen(open: boolean) {
    setUncontrolledOpen(open);
    props.onOpenChange?.(open);
  }

  const panZoom = createPanZoom({ minScale: 0.5, maxScale: 8 });

  // Memoized so testing for and rendering the slot doesn't build it twice.
  const controls = children(() => props.controls);

  const count = createMemo(() => props.images.length);
  const hasGallery = createMemo(() => count() > 1);
  const image = createMemo(() => props.images[current()] ?? props.images[0]);

  // Opening (or stepping to another image) starts from an untransformed view.
  createEffect(() => {
    if (isOpen()) {
      setCurrent(props.index ?? 0);
    }
  });

  function step(delta: number) {
    if (!hasGallery()) return;
    setCurrent((current() + delta + count()) % count());
    panZoom.reset();
  }

  function onKeyDown(e: KeyboardEvent) {
    switch (e.key) {
      case "ArrowRight":
        e.preventDefault();
        step(1);
        break;
      case "ArrowLeft":
        e.preventDefault();
        step(-1);
        break;
      case "+":
      case "=":
        e.preventDefault();
        panZoom.zoomIn();
        break;
      case "-":
        e.preventDefault();
        panZoom.zoomOut();
        break;
      case "0":
        e.preventDefault();
        panZoom.reset();
        break;
    }
  }

  let viewport: HTMLDivElement | undefined;
  let activePointer: number | null = null;

  function onPointerDown(e: PointerEvent) {
    activePointer = e.pointerId;
    panZoom.beginPan(e);
    viewport?.setPointerCapture(e.pointerId);
  }

  function onPointerMove(e: PointerEvent) {
    if (activePointer !== e.pointerId) return;
    panZoom.updatePan(e);
  }

  function onPointerUp(e: PointerEvent) {
    if (activePointer !== e.pointerId) return;
    panZoom.endPan();
    viewport?.releasePointerCapture?.(e.pointerId);
    activePointer = null;
  }

  onCleanup(() => {
    activePointer = null;
  });

  return (
    <DialogPrimitive
      open={isOpen()}
      onOpenChange={(open) => {
        if (open) panZoom.reset();
        setOpen(open);
      }}
    >
      <Show when={props.children}>
        <DialogPrimitive.Trigger
          as="button"
          type="button"
          class={cn("cursor-zoom-in", props.triggerClass)}
          aria-label={`View ${image()?.alt ?? "image"} full screen`}
        >
          {props.children}
        </DialogPrimitive.Trigger>
      </Show>

      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay class="fixed inset-0 z-50 bg-black/90 data-[expanded]:animate-in data-[closed]:animate-out data-[closed]:fade-out-0 data-[expanded]:fade-in-0" />
        <DialogPrimitive.Content
          class="fixed inset-0 z-50 flex flex-col focus:outline-none data-[expanded]:animate-in data-[closed]:animate-out data-[closed]:fade-out-0 data-[expanded]:fade-in-0"
          onKeyDown={onKeyDown}
        >
          <DialogPrimitive.Title class="sr-only">
            {image()?.alt ?? "Image viewer"}
          </DialogPrimitive.Title>

          <div
            ref={(el) => {
              viewport = el;
              panZoom.setViewport(el);
            }}
            class={cn(
              "relative flex-1 touch-none overflow-hidden",
              panZoom.isPanning() ? "cursor-grabbing" : "cursor-grab",
            )}
            style={props.backdropStyle}
            onPointerDown={onPointerDown}
            onPointerMove={onPointerMove}
            onPointerUp={onPointerUp}
            onPointerCancel={onPointerUp}
            onWheel={(e) => panZoom.onWheel(e)}
          >
            <div
              class="absolute inset-0 flex items-center justify-center p-8 sm:p-16"
              style={{
                transform: panZoom.transform(),
                "transform-origin": "center",
              }}
            >
              <Show when={image()}>
                {(img) => (
                  <img
                    src={img().src}
                    alt={img().alt}
                    draggable={false}
                    class="max-h-full max-w-full select-none object-contain"
                  />
                )}
              </Show>
            </div>
          </div>

          <Show when={controls()}>
            {/* Sibling of the viewport, not a child, so interacting with the
                controls never reads as the start of a pan. */}
            <div class="absolute left-3 top-3">{controls()}</div>
          </Show>

          <Show when={hasGallery()}>
            <button
              type="button"
              aria-label="Previous image"
              onClick={() => step(-1)}
              class="absolute left-3 top-1/2 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-2xl leading-none text-white transition-colors hover:bg-white/20 focus:outline-none focus-visible:ring-2 focus-visible:ring-white"
            >
              &#8249;
            </button>
            <button
              type="button"
              aria-label="Next image"
              onClick={() => step(1)}
              class="absolute right-3 top-1/2 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-2xl leading-none text-white transition-colors hover:bg-white/20 focus:outline-none focus-visible:ring-2 focus-visible:ring-white"
            >
              &#8250;
            </button>
          </Show>

          <div class="pointer-events-none absolute bottom-4 left-1/2 flex -translate-x-1/2 items-center gap-1 rounded-full bg-white/10 px-2 py-1 backdrop-blur">
            <button
              type="button"
              aria-label="Zoom out"
              onClick={panZoom.zoomOut}
              class="pointer-events-auto flex h-7 w-7 items-center justify-center rounded-full text-lg leading-none text-white hover:bg-white/20"
            >
              &minus;
            </button>
            <span class="w-12 text-center text-xs tabular-nums text-white/80">
              {Math.round(panZoom.scale() * 100)}%
            </span>
            <button
              type="button"
              aria-label="Zoom in"
              onClick={panZoom.zoomIn}
              class="pointer-events-auto flex h-7 w-7 items-center justify-center rounded-full text-lg leading-none text-white hover:bg-white/20"
            >
              +
            </button>
            <button
              type="button"
              onClick={panZoom.reset}
              class="pointer-events-auto ml-1 rounded-full px-2 py-0.5 text-xs text-white/80 hover:bg-white/20"
            >
              Reset
            </button>
            <Show when={hasGallery()}>
              <span class="ml-2 border-l border-white/20 pl-3 text-xs tabular-nums text-white/80">
                {current() + 1} / {count()}
              </span>
            </Show>
          </div>

          <DialogPrimitive.CloseButton
            class="absolute right-3 top-3 flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20 focus:outline-none focus-visible:ring-2 focus-visible:ring-white"
            aria-label="Close"
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" class="h-5 w-5">
              <path
                fill="none"
                stroke="currentColor"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M18 6L6 18M6 6l12 12"
              />
            </svg>
          </DialogPrimitive.CloseButton>

          <Show when={hasGallery()}>
            {/* Preload neighbours so stepping through the gallery is instant. */}
            <div class="hidden">
              <For each={[current() - 1, current() + 1]}>
                {(i) => (
                  <Show when={props.images[(i + count()) % count()]}>
                    {(neighbour) => (
                      <img src={neighbour().src} alt="" aria-hidden="true" />
                    )}
                  </Show>
                )}
              </For>
            </div>
          </Show>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive>
  );
}
