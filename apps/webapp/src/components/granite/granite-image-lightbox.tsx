import {
  ImageLightbox,
  type LightboxImage,
  type LightboxView,
} from "@glassact/ui";
import { createSignal, type ParentProps } from "solid-js";
import { graniteBackdropStyle, PLAIN_GRANITE_KEY } from "./granite";
import { GranitePill } from "./granite-pill";
import {
  createGranitePlacement,
  type InlayInches,
} from "./create-granite-placement";
import { useGranitePreference } from "../../hooks/use-granite-preference";

interface GraniteImageLightboxProps {
  images: LightboxImage[];
  // The inlay's physical size, so the granite is drawn at true scale.
  inches: InlayInches;
  /** Class for the clickable trigger box. */
  triggerClass?: string;
  /** Class for the granite-backed panel the thumbnail sits on. */
  previewClass?: string;
}

// An ImageLightbox whose thumbnail and fullscreen view both sit on the shared
// granite backdrop, with the picker available in the fullscreen view.
export function GraniteImageLightbox(
  props: ParentProps<GraniteImageLightboxProps>,
) {
  const preference = useGranitePreference(PLAIN_GRANITE_KEY);
  const [panel, setPanel] = createSignal<HTMLDivElement>();

  const panelPlacement = createGranitePlacement({
    backdrop: panel,
    inlay: () => panel()?.querySelector("img") ?? undefined,
    inches: () => props.inches,
  });

  function viewerBackdrop(view: LightboxView) {
    const viewerPlacement = createGranitePlacement({
      backdrop: view.viewport,
      inlay: view.image,
      inches: () => props.inches,
      track: view.transform,
    });
    return () => graniteBackdropStyle(preference.granite(), viewerPlacement());
  }

  return (
    <ImageLightbox
      images={props.images}
      triggerClass={props.triggerClass}
      backdropStyle={viewerBackdrop}
      controls={<GranitePill preference={preference} />}
    >
      <div
        ref={setPanel}
        class={props.previewClass}
        style={graniteBackdropStyle(preference.granite(), panelPlacement())}
      >
        {props.children}
      </div>
    </ImageLightbox>
  );
}
