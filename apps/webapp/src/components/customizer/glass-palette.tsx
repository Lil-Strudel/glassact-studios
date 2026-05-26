import { createMemo, createSignal, For, Show } from "solid-js";

export interface Swatch {
  id: number;
  name: string;
  hex: string;
  family?: string | null;
}

interface SwatchPickerProps {
  swatches: Swatch[];
  selectedId: number | null;
  usedIds?: number[];
  onSelect: (id: number) => void;
  onHoverChange?: (id: number | null) => void;
  searchPlaceholder?: string;
  disabled?: boolean;
}

const FAMILY_ORDER = [
  "neutral",
  "blue",
  "green",
  "purple",
  "brown",
  "amber",
  "red",
];

function familyLabel(family: string): string {
  return family.charAt(0).toUpperCase() + family.slice(1);
}

export function SwatchPicker(props: SwatchPickerProps) {
  const [search, setSearch] = createSignal("");

  const filtered = createMemo(() => {
    const q = search().trim().toLowerCase();
    if (!q) return props.swatches;
    return props.swatches.filter((s) => s.name.toLowerCase().includes(q));
  });

  const grouped = createMemo(() => {
    const groups = new Map<string, Swatch[]>();
    for (const s of filtered()) {
      const key = s.family ?? "";
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key)!.push(s);
    }
    return [...groups.entries()].sort((a, b) => {
      const ia = FAMILY_ORDER.indexOf(a[0]);
      const ib = FAMILY_ORDER.indexOf(b[0]);
      return (ia === -1 ? 999 : ia) - (ib === -1 ? 999 : ib);
    });
  });

  const usedSwatches = createMemo(() => {
    const used = new Set(props.usedIds ?? []);
    if (used.size === 0) return [];
    return props.swatches.filter((s) => used.has(s.id));
  });

  return (
    <div class="flex flex-col gap-3">
      <input
        type="search"
        value={search()}
        onInput={(e) => setSearch(e.currentTarget.value)}
        placeholder={props.searchPlaceholder ?? "Search colors..."}
        class="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      />

      <Show when={usedSwatches().length > 0 && !search()}>
        <div class="flex flex-col gap-1.5">
          <p class="text-xs font-medium text-gray-500">Used in this design</p>
          <div class="flex flex-wrap gap-1.5">
            <For each={usedSwatches()}>
              {(s) => (
                <SwatchButton
                  swatch={s}
                  selected={props.selectedId === s.id}
                  onSelect={props.onSelect}
                  onHoverChange={props.onHoverChange}
                  disabled={props.disabled}
                  compact
                />
              )}
            </For>
          </div>
        </div>
      </Show>

      <div class="flex max-h-72 flex-col gap-3 overflow-y-auto pr-1">
        <For each={grouped()}>
          {([family, swatches]) => (
            <div class="flex flex-col gap-1.5">
              <Show when={family}>
                <p class="text-xs font-medium text-gray-500">
                  {familyLabel(family)}
                </p>
              </Show>
              <div class="grid grid-cols-2 gap-1.5">
                <For each={swatches}>
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
              </div>
            </div>
          )}
        </For>
        <Show when={filtered().length === 0}>
          <p class="py-4 text-center text-sm text-gray-400">No colors match.</p>
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
  compact?: boolean;
}) {
  return (
    <button
      type="button"
      disabled={props.disabled}
      title={props.swatch.name}
      onClick={() => props.onSelect(props.swatch.id)}
      onMouseEnter={() => props.onHoverChange?.(props.swatch.id)}
      onMouseLeave={() => props.onHoverChange?.(null)}
      class="flex items-center gap-2 rounded-md border p-1 text-left transition disabled:cursor-not-allowed disabled:opacity-50"
      classList={{
        "border-blue-600 ring-2 ring-blue-500/40": props.selected,
        "border-gray-200 hover:border-gray-400": !props.selected,
      }}
    >
      <span
        class="h-6 w-6 shrink-0 rounded border border-black/10"
        style={{ "background-color": props.swatch.hex }}
      />
      <Show when={!props.compact}>
        <span class="truncate text-xs text-gray-700">{props.swatch.name}</span>
      </Show>
    </button>
  );
}
