import { cn, Popover, PopoverContent, PopoverTrigger } from "@glassact/ui";
import { GranitePicker } from "./granite-picker";
import { graniteChoiceName, graniteSwatchStyle } from "./granite";
import type { GranitePreference } from "../../hooks/use-granite-preference";

interface GranitePillProps {
  preference: GranitePreference;
  class?: string;
}

// The picker behind a floating chip, for overlaying on a preview surface (the
// customizer canvas, the fullscreen lightbox) without covering the inlay.
export function GranitePill(props: GranitePillProps) {
  return (
    <Popover placement="bottom-start">
      <PopoverTrigger
        class={cn(
          "flex items-center gap-1.5 rounded-full border border-gray-200 bg-white/90 py-1 pl-1 pr-2.5 text-[11px] font-medium text-gray-600 shadow-sm hover:bg-white",
          props.class,
        )}
      >
        <span
          class="h-5 w-5 shrink-0 rounded-full border border-black/15"
          style={graniteSwatchStyle(props.preference.granite())}
        />
        <span class="whitespace-nowrap">
          Granite: {graniteChoiceName(props.preference.granite())}
        </span>
        <span aria-hidden="true" class="text-gray-400">
          &#9662;
        </span>
      </PopoverTrigger>
      <PopoverContent class="w-64">
        <GranitePicker preference={props.preference} />
      </PopoverContent>
    </Popover>
  );
}
