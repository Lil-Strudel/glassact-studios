import { For } from "solid-js";
import { cn } from "@glassact/ui";
import {
  CUSTOM_GRANITE_KEY,
  GRANITE_SAMPLES,
  PLAIN_GRANITE_KEY,
  graniteChoiceName,
  graniteSwatchStyle,
  type GraniteChoice,
} from "./granite";
import type { GranitePreference } from "../../hooks/use-granite-preference";

interface GranitePickerProps {
  preference: GranitePreference;
  class?: string;
}

const SWATCH_CLASS =
  "pointer-events-auto aspect-square w-full rounded-full border transition";

export function GranitePicker(props: GranitePickerProps) {
  const isSelected = (key: string) => props.preference.graniteKey() === key;

  return (
    <div class={cn("flex flex-col gap-2", props.class)}>
      <div class="grid grid-cols-[repeat(auto-fill,minmax(1.5rem,1fr))] gap-1.5">
        <SwatchButton
          choice={{ kind: "plain" }}
          isSelected={isSelected(PLAIN_GRANITE_KEY)}
          onSelect={() => props.preference.setGraniteKey(PLAIN_GRANITE_KEY)}
        />
        <For each={GRANITE_SAMPLES}>
          {(sample) => (
            <SwatchButton
              choice={{ kind: "sample", sample }}
              isSelected={isSelected(sample.key)}
              onSelect={() => props.preference.setGraniteKey(sample.key)}
            />
          )}
        </For>
        <CustomColorSwatch
          hex={props.preference.customHex()}
          isSelected={isSelected(CUSTOM_GRANITE_KEY)}
          onSelect={() => props.preference.setGraniteKey(CUSTOM_GRANITE_KEY)}
          onPick={props.preference.setCustomHex}
        />
      </div>
      <p class="text-xs text-gray-600">
        {graniteChoiceName(props.preference.granite())}
      </p>
    </div>
  );
}

function swatchStateClass(isSelected: boolean) {
  return isSelected
    ? "border-blue-600 ring-2 ring-blue-500/40"
    : "border-black/15 hover:border-gray-500";
}

function SwatchButton(props: {
  choice: GraniteChoice;
  isSelected: boolean;
  onSelect: () => void;
}) {
  const name = () => graniteChoiceName(props.choice);
  return (
    <button
      type="button"
      title={name()}
      aria-label={`${name()} background`}
      aria-pressed={props.isSelected}
      onClick={() => props.onSelect()}
      class={cn(SWATCH_CLASS, swatchStateClass(props.isSelected))}
      style={graniteSwatchStyle(props.choice)}
    />
  );
}

// The native color input sits invisibly over the swatch, so clicking the
// swatch both selects the custom backdrop and opens the browser's color
// picker anchored right there.
function CustomColorSwatch(props: {
  hex: string;
  isSelected: boolean;
  onSelect: () => void;
  onPick: (hex: string) => void;
}) {
  return (
    <span
      class={cn(
        SWATCH_CLASS,
        "relative bg-[conic-gradient(#f87171,#facc15,#4ade80,#22d3ee,#818cf8,#e879f9,#f87171)] p-0.5",
        swatchStateClass(props.isSelected),
      )}
      title="Custom color — for a granite not listed"
    >
      <span
        class="block h-full w-full rounded-full"
        style={graniteSwatchStyle({ kind: "custom", hex: props.hex })}
      />
      <input
        type="color"
        aria-label="Custom granite color"
        value={props.hex}
        onClick={() => props.onSelect()}
        onInput={(e) => props.onPick(e.currentTarget.value)}
        class="absolute inset-0 h-full w-full cursor-pointer appearance-none rounded-full opacity-0"
      />
    </span>
  );
}
