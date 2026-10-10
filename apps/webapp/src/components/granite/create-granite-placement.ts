import {
  createEffect,
  createSignal,
  onCleanup,
  type Accessor,
} from "solid-js";

export interface GranitePlacement {
  pxPerInch: number;
  // The inlay artwork's top-left corner, relative to the backdrop's padding box.
  originX: number;
  originY: number;
}

export interface InlayInches {
  width: number;
  height: number;
}

interface GranitePlacementOptions {
  backdrop: Accessor<HTMLElement | undefined>;
  // The element the inlay artwork is drawn in (an <svg> or <img>). Its box may
  // be larger than the artwork; the artwork is assumed letterboxed inside it
  // (object-fit: contain / preserveAspectRatio meet), at the inches' aspect.
  inlay: Accessor<Element | undefined>;
  inches: Accessor<InlayInches | undefined>;
  // A reactive value that moves the inlay without resizing anything, e.g. a
  // pan/zoom transform, which no ResizeObserver would notice.
  track?: Accessor<unknown>;
}

// Measures how big an inch of the inlay is on screen and where the inlay sits,
// so the granite behind it can be drawn at true scale and stay locked to it.
export function createGranitePlacement(
  options: GranitePlacementOptions,
): Accessor<GranitePlacement | null> {
  const [placement, setPlacement] = createSignal<GranitePlacement | null>(null);

  function measure() {
    const backdrop = options.backdrop();
    const inlay = options.inlay();
    const inches = options.inches();
    if (
      !backdrop ||
      !inlay ||
      !inches ||
      inches.width <= 0 ||
      inches.height <= 0
    ) {
      setPlacement(null);
      return;
    }

    const box = inlay.getBoundingClientRect();
    const pxPerInch = Math.min(
      box.width / inches.width,
      box.height / inches.height,
    );
    if (!(pxPerInch > 0)) {
      setPlacement(null);
      return;
    }

    const backdropBox = backdrop.getBoundingClientRect();
    setPlacement({
      pxPerInch,
      originX:
        box.left + (box.width - inches.width * pxPerInch) / 2 -
        backdropBox.left - backdrop.clientLeft,
      originY:
        box.top + (box.height - inches.height * pxPerInch) / 2 -
        backdropBox.top - backdrop.clientTop,
    });
  }

  createEffect(() => {
    const backdrop = options.backdrop();
    const inlay = options.inlay();
    if (!backdrop || !inlay) return;

    const observer = new ResizeObserver(measure);
    observer.observe(backdrop);
    observer.observe(inlay);
    onCleanup(() => observer.disconnect());
  });

  createEffect(() => {
    options.inches();
    options.track?.();
    measure();
  });

  return placement;
}
