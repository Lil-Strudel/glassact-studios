import { cn } from "@glassact/ui";
import { GranitePicker } from "./granite-picker";

interface GranitePillProps {
  selectedKey: string;
  onSelect: (key: string) => void;
  class?: string;
}

// The picker as a floating chip, for overlaying on a preview surface (the
// customizer canvas, the fullscreen lightbox). The wrapper stays
// pointer-transparent so only the swatches themselves take clicks.
export function GranitePill(props: GranitePillProps) {
  return (
    <div
      class={cn(
        "pointer-events-none flex flex-nowrap items-center gap-1 rounded-full border border-gray-200 bg-white/90 px-2 py-1 shadow-sm",
        props.class,
      )}
    >
      <span class="whitespace-nowrap px-1 text-[11px] font-medium text-gray-500">
        Granite
      </span>
      <GranitePicker
        selectedKey={props.selectedKey}
        onSelect={props.onSelect}
      />
    </div>
  );
}
