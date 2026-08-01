import { For } from "solid-js";
import { cn } from "@glassact/ui";
import { GRANITE_PRESETS, graniteSwatchStyle } from "./granite";

interface GranitePickerProps {
  selectedKey: string;
  onSelect: (key: string) => void;
  class?: string;
}

// Swatches stay on one row: `flex-nowrap`, plus `min-w-0` so they compress
// rather than wrap if the row ever outgrows its container (a narrower sidebar,
// more presets).
export function GranitePicker(props: GranitePickerProps) {
  return (
    <div class={cn("flex flex-nowrap items-center gap-1", props.class)}>
      <For each={GRANITE_PRESETS}>
        {(preset) => (
          <button
            type="button"
            title={preset.name}
            aria-label={`${preset.name} background`}
            aria-pressed={props.selectedKey === preset.key}
            onClick={() => props.onSelect(preset.key)}
            class="pointer-events-auto h-6 w-6 min-w-0 shrink rounded-full border transition"
            classList={{
              "border-blue-600 ring-2 ring-blue-500/40":
                props.selectedKey === preset.key,
              "border-black/15 hover:border-gray-500":
                props.selectedKey !== preset.key,
            }}
            style={graniteSwatchStyle(preset)}
          />
        )}
      </For>
    </div>
  );
}
