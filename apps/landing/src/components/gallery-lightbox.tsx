import { createSignal, onCleanup, onMount } from "solid-js";
import { ImageLightbox, type LightboxImage } from "@glassact/ui";

interface GalleryLightboxProps {
  images: LightboxImage[];
  /** Id of the statically rendered grid whose tiles open the viewer. */
  gridId: string;
}

/**
 * Opens the gallery images full screen.
 *
 * The grid itself stays static, build-time-optimized Astro HTML — this island
 * renders nothing until something is clicked, and finds its targets with a
 * single delegated listener rather than hydrating 54 tiles.
 */
export function GalleryLightbox(props: GalleryLightboxProps) {
  const [open, setOpen] = createSignal(false);
  const [index, setIndex] = createSignal(0);

  onMount(() => {
    const grid = document.getElementById(props.gridId);
    if (!grid) return;

    function handleClick(event: MouseEvent) {
      const tile = (event.target as HTMLElement | null)?.closest<HTMLElement>(
        "[data-gallery-index]",
      );
      if (!tile) return;

      const parsed = Number(tile.dataset.galleryIndex);
      if (!Number.isInteger(parsed)) return;

      event.preventDefault();
      setIndex(parsed);
      setOpen(true);
    }

    grid.addEventListener("click", handleClick);
    onCleanup(() => grid.removeEventListener("click", handleClick));
  });

  return (
    <ImageLightbox
      images={props.images}
      index={index()}
      open={open()}
      onOpenChange={setOpen}
    />
  );
}

export default GalleryLightbox;
