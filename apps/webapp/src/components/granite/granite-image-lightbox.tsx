import { ImageLightbox, type LightboxImage } from "@glassact/ui";
import type { ParentProps } from "solid-js";
import { graniteBackgroundStyle, PLAIN_GRANITE_KEY } from "./granite";
import { GranitePill } from "./granite-pill";
import { useGranitePreference } from "../../hooks/use-granite-preference";

interface GraniteImageLightboxProps {
  images: LightboxImage[];
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
  const { graniteKey, granite, setGraniteKey } =
    useGranitePreference(PLAIN_GRANITE_KEY);

  return (
    <ImageLightbox
      images={props.images}
      triggerClass={props.triggerClass}
      backdropStyle={graniteBackgroundStyle(granite())}
      controls={
        <GranitePill selectedKey={graniteKey()} onSelect={setGraniteKey} />
      }
    >
      <div class={props.previewClass} style={graniteBackgroundStyle(granite())}>
        {props.children}
      </div>
    </ImageLightbox>
  );
}
