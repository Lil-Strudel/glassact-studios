import { createMemo, createSignal, For, Show } from "solid-js";
import { TextFieldRoot, TextField } from "@glassact/ui";

export interface Swatch {
  id: number;
  name: string;
  hex: string;
  family?: string | null;
}

interface SwatchPickerProps {
  swatches: Swatch[];
  selectedId: number | null;
  onSelect: (id: number) => void;
  onHoverChange?: (id: number | null) => void;
  // When provided, a search box is shown (the color name is otherwise only a
  // hover tooltip). The consumer customizer omits this for a pure swatch grid;
  // the admin manifest editor passes it to search a long palette by name.
  searchPlaceholder?: string;
  disabled?: boolean;
}

export function SwatchPicker(props: SwatchPickerProps) {
  const [search, setSearch] = createSignal("");

  const filtered = createMemo(() => {
    const q = search().trim().toLowerCase();
    if (!q) return props.swatches;
    return props.swatches.filter((s) => s.name.toLowerCase().includes(q));
  });

  return (
    <div class="flex flex-col gap-3">
      <Show when={props.searchPlaceholder}>
        <TextFieldRoot class="w-full">
          <TextField
            type="search"
            value={search()}
            onInput={(e) => setSearch(e.currentTarget.value)}
            placeholder={props.searchPlaceholder}
          />
        </TextFieldRoot>
      </Show>

      <div class="grid grid-cols-10 gap-1">
        <For each={filtered()}>
          {(s) => (
            <SwatchButton
              swatch={s}
              selected={props.selectedId === s.id}
              onSelect={props.onSelect}
              onHoverChange={props.onHoverChange}
              disabled={props.disabled}
            />
          )}
        </For>
        <Show when={filtered().length === 0}>
          <p class="col-span-full py-4 text-center text-sm text-gray-400">
            No colors match.
          </p>
        </Show>
      </div>
    </div>
  );
}

function SwatchButton(props: {
  swatch: Swatch;
  selected: boolean;
  onSelect: (id: number) => void;
  onHoverChange?: (id: number | null) => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      disabled={props.disabled}
      title={props.swatch.name}
      aria-label={props.swatch.name}
      onClick={() => props.onSelect(props.swatch.id)}
      onMouseEnter={() => props.onHoverChange?.(props.swatch.id)}
      onMouseLeave={() => props.onHoverChange?.(null)}
      class="aspect-square w-full rounded border transition disabled:cursor-not-allowed disabled:opacity-50"
      classList={{
        "border-blue-600 ring-2 ring-blue-500/40": props.selected,
        "border-black/10 hover:border-gray-500": !props.selected,
      }}
      style={{ "background-color": props.swatch.hex }}
    />
  );
}
