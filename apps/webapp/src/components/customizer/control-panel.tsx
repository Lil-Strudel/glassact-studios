import { createMemo, For, Show } from "solid-js";
import type { Manifest, ColorOverrides, GlassColor, Grout, GET } from "@glassact/data";
import {
  Badge,
  Button,
  NumberFieldRoot,
  NumberField,
  NumberFieldLabel,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsIndicator,
} from "@glassact/ui";
import {
  SwatchPicker,
  type Swatch,
  type GlassById,
  type Selection,
  customPieceCount,
  groupGlassId,
} from "./shared";

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
  width: number;
  height: number;
  minWidth: number;
  minHeight: number;
  onSelectGroup: (groupKey: string) => void;
  onRegionHover: (groupKey: string | null) => void;
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

  const groupEntries = createMemo(() =>
    Object.entries(props.manifest.glass_regions ?? {}),
  );

  const selectionLabel = createMemo(() => {
    const sel = props.selection;
    if (!sel) return null;
    if (sel.type === "group") {
      const count = props.manifest.glass_regions?.[sel.groupKey]?.count ?? 0;
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
        <Tabs
          value={props.mode}
          onChange={(v) => props.onRequestMode(v as "group" | "piece")}
        >
          <TabsList>
            <TabsTrigger value="group" class="data-[selected]:text-primary-foreground">Color groups</TabsTrigger>
            <TabsTrigger value="piece" class="data-[selected]:text-primary-foreground">Individual pieces</TabsTrigger>
            <TabsIndicator class="bg-primary" />
          </TabsList>
        </Tabs>
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
            <Button
              variant="link"
              size="sm"
              class="h-auto p-0 text-xs"
              onClick={() => {
                const sel = props.selection;
                if (sel?.type === "piece") props.onResetPiece(sel.pieceId);
              }}
            >
              Reset to group
            </Button>
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
          <For each={groupEntries()}>
            {([groupKey, region]) => {
              const glassId = () =>
                groupGlassId(groupKey, props.overrides, props.manifest);
              const glass = () => {
                const id = glassId();
                return id != null ? props.glassById.get(id) : undefined;
              };
              const custom = () =>
                customPieceCount(region.piece_ids, props.overrides);
              const isSelected = () =>
                props.selection?.type === "group" &&
                props.selection.groupKey === groupKey;
              return (
                <button
                  type="button"
                  onClick={() => props.onSelectGroup(groupKey)}
                  onMouseEnter={() => props.onRegionHover(groupKey)}
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
                      style={{
                        "background-color":
                          glass()?.hex ?? region.source_hex ?? "#cccccc",
                      }}
                    />
                  </span>
                  <span class="min-w-0 flex-1">
                    <span class="block truncate text-xs font-medium text-gray-800">
                      {glass()?.name ?? "Original color"}
                    </span>
                    <span class="block text-[11px] text-gray-500">
                      {region.count} piece
                      {region.count === 1 ? "" : "s"}
                    </span>
                  </span>
                  <Show when={custom() > 0}>
                    <Badge variant="warning" class="shrink-0 rounded-full px-1.5 py-0.5 text-[10px]">
                      {custom()} custom
                    </Badge>
                  </Show>
                </button>
              );
            }}
          </For>
        </div>
      </div>

      <div>
        <p class="mb-1.5 text-xs font-medium uppercase tracking-wide text-gray-500">
          Grout
        </p>
        <SwatchPicker
          swatches={groutSwatches()}
          selectedId={
            props.overrides.background?.grout_id ??
            props.manifest.grout_region.grout_id
          }
          onSelect={props.onSelectGrout}
          searchPlaceholder="Search grout..."
        />
      </div>

      <div>
        <p class="mb-1.5 text-xs font-medium uppercase tracking-wide text-gray-500">
          Size
        </p>
        <div class="flex items-end gap-3">
          <NumberFieldRoot class="flex flex-col gap-1 text-xs text-gray-600">
            <NumberFieldLabel>Width (in)</NumberFieldLabel>
            <NumberField
              decimalPlaces={1}
              class="w-24"
              value={String(props.width)}
              onChange={(v) => props.onWidthChange(Number(v))}
            />
          </NumberFieldRoot>
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
