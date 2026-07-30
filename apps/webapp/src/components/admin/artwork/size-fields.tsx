import { createEffect, createSignal } from "solid-js";
import { NumberField, NumberFieldLabel, NumberFieldRoot } from "@glassact/ui";

interface SizeFieldsProps {
  width: number | null;
  height: number | null;
  // Height per unit width of the trimmed artwork.
  aspect: number;
  onChange: (width: number | null, height: number | null) => void;
  class?: string;
}

function toText(value: number | null): string {
  return value == null ? "" : String(value);
}

function toNumber(text: string): number | null {
  if (text.trim() === "") return null;
  const parsed = parseFloat(text);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function round2(value: number): number {
  return Math.round(value * 100) / 100;
}

// Width/height pair locked to the artwork's aspect ratio: editing either
// dimension recomputes the other. The raw text is held locally so partially
// typed values ("3.") survive the round trip through the numeric props.
export function SizeFields(props: SizeFieldsProps) {
  const [widthText, setWidthText] = createSignal("");
  const [heightText, setHeightText] = createSignal("");

  createEffect(() => {
    if (toNumber(widthText()) !== props.width) setWidthText(toText(props.width));
  });

  createEffect(() => {
    if (toNumber(heightText()) !== props.height) {
      setHeightText(toText(props.height));
    }
  });

  function handleWidth(text: string) {
    setWidthText(text);
    const width = toNumber(text);
    props.onChange(width, width == null ? null : round2(width * props.aspect));
  }

  function handleHeight(text: string) {
    setHeightText(text);
    const height = toNumber(text);
    props.onChange(height == null ? null : round2(height / props.aspect), height);
  }

  return (
    <div class={props.class}>
      <div class="flex items-end gap-3">
        <NumberFieldRoot class="flex flex-col gap-1">
          <NumberFieldLabel>Default Width (in)</NumberFieldLabel>
          <NumberField
            class="w-28"
            decimalPlaces={2}
            placeholder="e.g., 3.00"
            value={widthText()}
            onChange={handleWidth}
          />
        </NumberFieldRoot>
        <span class="pb-2 text-gray-400">×</span>
        <NumberFieldRoot class="flex flex-col gap-1">
          <NumberFieldLabel>Default Height (in)</NumberFieldLabel>
          <NumberField
            class="w-28"
            decimalPlaces={2}
            placeholder="e.g., 4.25"
            value={heightText()}
            onChange={handleHeight}
          />
        </NumberFieldRoot>
      </div>
      <p class="mt-1.5 text-xs text-gray-500">
        Aspect ratio is locked to the artwork — editing either dimension updates
        the other.
      </p>
    </div>
  );
}
