import { createMemo, For, Show } from "solid-js";
import type { Manifest, ColorOverrides, GlassColor, Grout, GET } from "@glassact/data";
import { SwatchPicker, type Swatch } from "./glass-palette";
import {
  type GlassById,
  type Selection,
  customPieceCount,
  regionGlassId,
} from "./resolution";

interface ControlPanelProps {
  mode: "group" | "piece";
  onRequestMode: (mode: "group" | "piece") => void;
  manifest: Manifest;
  glassColors: GET<GlassColor>[];
  glassById: GlassById;
  grouts: GET<Grout>[];
  overrides: ColorOverrides;
  selection: Selection | null;
  selectedGlassId: number | null;
  usedGlassIds: number[];
  backgroundGroutId: number | null;
  width: number;
  height: number;
  minWidth: number;
  minHeight: number;
  onSelectRegion: (sourceHex: string) => void;
  onRegionHover: (sourceHex: string | null) => void;
  onAssignGlass: (glassId: number) => void;
  onHoverGlass: (glassId: number | null) => void;
  onResetPiece: (pieceId: string) => void;
  onSelectGrout: (groutId: number) => void;
  onWidthChange: (width: number) => void;
}

export function ControlPanel(props: ControlPanelProps) {
  const glassSwatches = createMemo<Swatch[]>(() =>
    props.glassColors.map((g) => ({
      id: g.id,
      name: g.name,
      hex: g.hex,
      family: g.family,
    })),
  );
  const groutSwatches = createMemo<Swatch[]>(() =>
    props.grouts.map((g) => ({ id: g.id, name: g.name, hex: g.hex })),
  );

  const regionEntries = createMemo(() =>
    Object.entries(props.manifest.regions ?? {}),
  );

  const selectionLabel = createMemo(() => {
    const sel = props.selection;
    if (!sel) return null;
    if (sel.type === "region") {
      const region = props.manifest.regions?.[sel.sourceHex];
      const count = region?.piece_ids.length ?? 0;
      return `Color group · ${count} piece${count === 1 ? "" : "s"}`;
    }
    return "Single piece";
  });

  return (
    <div class="flex flex-col gap-5">
      <div>
        <p class="mb-1.5 text-xs font-medium uppercase tracking-wide text-gray-500">
          Edit mode
        </p>
        <div class="inline-flex rounded-md border border-gray-300 p-0.5">
          <ModeButton
            active={props.mode === "group"}
            onClick={() => props.onRequestMode("group")}
          >
            Color groups
          </ModeButton>
          <ModeButton
            active={props.mode === "piece"}
            onClick={() => props.onRequestMode("piece")}
          >
            Individual pieces
          </ModeButton>
        </div>
        <p class="mt-1.5 text-xs text-gray-500">
          {props.mode === "group"
            ? "Click a piece or a group below to recolor the whole color."
            : "Click a single piece to recolor just that piece."}
        </p>
      </div>

      <div class="rounded-lg border border-gray-200 p-3">
        <div class="mb-2 flex items-center justify-between">
          <p class="text-sm font-semibold text-gray-900">
            {selectionLabel() ?? "Choose glass color"}
          </p>
          <Show when={props.selection?.type === "piece"}>
            <button
              type="button"
              class="text-xs text-blue-600 hover:underline"
              onClick={() => {
                const sel = props.selection;
                if (sel?.type === "piece") props.onResetPiece(sel.pieceId);
              }}
            >
              Reset to group
            </button>
          </Show>
        </div>

        <Show
          when={props.selection}
          fallback={
            <p class="py-3 text-center text-sm text-gray-400">
              Select a color group or click a piece to recolor it.
            </p>
          }
        >
          <SwatchPicker
            swatches={glassSwatches()}
            selectedId={props.selectedGlassId}
            usedIds={props.usedGlassIds}
            onSelect={props.onAssignGlass}
            onHoverChange={props.onHoverGlass}
            searchPlaceholder="Search glass colors..."
          />
        </Show>
      </div>

      <div>
        <p class="mb-1.5 text-xs font-medium uppercase tracking-wide text-gray-500">
          Color groups
        </p>
        <div class="flex flex-col gap-1">
          <For each={regionEntries()}>
            {([sourceHex, region]) => {
              const glassId = () => regionGlassId(sourceHex, props.overrides);
              const glass = () => {
                const id = glassId();
                return id != null ? props.glassById.get(id) : undefined;
              };
              const custom = () =>
                customPieceCount(region.piece_ids, props.overrides);
              const isSelected = () =>
                props.selection?.type === "region" &&
                props.selection.sourceHex === sourceHex;
              return (
                <button
                  type="button"
                  onClick={() => props.onSelectRegion(sourceHex)}
                  onMouseEnter={() => props.onRegionHover(sourceHex)}
                  onMouseLeave={() => props.onRegionHover(null)}
                  class="flex items-center gap-2 rounded-md border p-1.5 text-left transition"
                  classList={{
                    "border-blue-600 bg-blue-50": isSelected(),
                    "border-gray-200 hover:border-gray-400": !isSelected(),
                  }}
                >
                  <span class="flex shrink-0 items-center">
                    <span
                      class="h-6 w-6 rounded border border-black/10"
                      style={{ "background-color": glass()?.hex ?? sourceHex }}
                    />
                  </span>
                  <span class="min-w-0 flex-1">
                    <span class="block truncate text-xs font-medium text-gray-800">
                      {glass()?.name ?? "Original color"}
                    </span>
                    <span class="block text-[11px] text-gray-500">
                      {region.piece_ids.length} piece
                      {region.piece_ids.length === 1 ? "" : "s"}
                    </span>
                  </span>
                  <Show when={custom() > 0}>
                    <span class="shrink-0 rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700">
                      {custom()} custom
                    </span>
                  </Show>
                </button>
              );
            }}
          </For>
        </div>
      </div>

      <div>
        <p class="mb-1.5 text-xs font-medium uppercase tracking-wide text-gray-500">
          Background (grout)
        </p>
        <SwatchPicker
          swatches={groutSwatches()}
          selectedId={props.backgroundGroutId}
          onSelect={props.onSelectGrout}
          searchPlaceholder="Search grout..."
        />
      </div>

      <div>
        <p class="mb-1.5 text-xs font-medium uppercase tracking-wide text-gray-500">
          Size
        </p>
        <div class="flex items-end gap-3">
          <label class="flex flex-col gap-1 text-xs text-gray-600">
            Width (in)
            <input
              type="number"
              min={props.minWidth}
              step="0.5"
              value={props.width}
              onChange={(e) =>
                props.onWidthChange(Number(e.currentTarget.value))
              }
              class="w-24 rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none"
            />
          </label>
          <span class="pb-1.5 text-gray-400">×</span>
          <div class="flex flex-col gap-1 text-xs text-gray-600">
            Height (in)
            <div class="w-24 rounded-md border border-gray-200 bg-gray-50 px-2 py-1 text-sm text-gray-500">
              {props.height.toFixed(1)}
            </div>
          </div>
        </div>
        <p class="mt-1 text-[11px] text-gray-400">
          Aspect ratio locked. Minimum {props.minWidth}" × {props.minHeight}".
        </p>
      </div>
    </div>
  );
}

function ModeButton(props: {
  active: boolean;
  onClick: () => void;
  children: string;
}) {
  return (
    <button
      type="button"
      onClick={() => props.onClick()}
      class="rounded px-3 py-1 text-sm transition"
      classList={{
        "bg-blue-600 text-white": props.active,
        "text-gray-600 hover:bg-gray-100": !props.active,
      }}
    >
      {props.children}
    </button>
  );
}
