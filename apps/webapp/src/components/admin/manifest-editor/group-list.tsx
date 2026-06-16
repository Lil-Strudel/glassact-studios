import { createMemo, For, Show } from "solid-js";
import type { GlassColor, Grout, Manifest, GET } from "@glassact/data";
import { Badge, Button } from "@glassact/ui";
import { SwatchPicker, type Swatch, type Selection } from "../../customizer/shared";

interface GroupListProps {
  manifest: Manifest;
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
  selection: Selection | null;
  // The group key currently open for color editing (group selection).
  activeGroupKey: string | null;
  // Set of pieces tagged for a move operation.
  selectedPieceIds: string[];
  onSelectGroup: (groupKey: string) => void;
  onHoverGroup: (groupKey: string | null) => void;
  onAssignGroupColor: (groupKey: string, glassColorId: number) => void;
  onAssignGroutColor: (groutId: number) => void;
  onMarkGroupAsGrout: (groupKey: string) => void;
  onMergeInto: (targetKey: string) => void;
  onMovePiecesToGroup: (targetKey: string) => void;
  onMovePiecesToGrout: () => void;
  onSplitSelected: () => void;
  onClearPieceSelection: () => void;
}

export function GroupList(props: GroupListProps) {
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

  const glassById = createMemo(
    () => new Map(props.glassColors.map((g) => [g.id, g])),
  );

  const groups = createMemo(() =>
    Object.entries(props.manifest.glass_regions).sort((a, b) =>
      a[0].localeCompare(b[0]),
    ),
  );

  const hasPieceSelection = createMemo(() => props.selectedPieceIds.length > 0);

  return (
    <div class="flex flex-col gap-4">
      <Show when={hasPieceSelection()}>
        <div class="rounded-lg border border-blue-200 bg-blue-50 p-3">
          <div class="mb-2 flex items-center justify-between">
            <p class="text-sm font-medium text-blue-900">
              {props.selectedPieceIds.length} piece
              {props.selectedPieceIds.length === 1 ? "" : "s"} selected
            </p>
            <Button
              variant="ghost"
              size="sm"
              onClick={props.onClearPieceSelection}
            >
              Clear
            </Button>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button size="sm" variant="outline" onClick={props.onSplitSelected}>
              Split to new group
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={props.onMovePiecesToGrout}
            >
              Move to grout
            </Button>
          </div>
          <p class="mt-2 text-xs text-blue-700">
            Or pick a group below, then use its "Move here" action.
          </p>
        </div>
      </Show>

      <div class="flex flex-col gap-2">
        <h3 class="text-sm font-semibold text-gray-900">
          Glass groups ({groups().length})
        </h3>
        <For each={groups()}>
          {([groupKey, region]) => {
            const isActive = () => props.activeGroupKey === groupKey;
            const assigned = () =>
              region.glass_color_id != null
                ? glassById().get(region.glass_color_id)
                : undefined;
            return (
              <div
                class="rounded-lg border transition"
                classList={{
                  "border-blue-500 ring-1 ring-blue-400": isActive(),
                  "border-gray-200": !isActive(),
                }}
                onMouseEnter={() => props.onHoverGroup(groupKey)}
                onMouseLeave={() => props.onHoverGroup(null)}
              >
                <button
                  type="button"
                  class="flex w-full items-center gap-3 p-3 text-left"
                  onClick={() => props.onSelectGroup(groupKey)}
                >
                  <Show
                    when={assigned()}
                    fallback={
                      <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded border-2 border-dashed border-amber-400 bg-amber-50 text-amber-600">
                        !
                      </span>
                    }
                  >
                    {(g) => (
                      <span
                        class="h-7 w-7 shrink-0 rounded border border-black/10"
                        style={{ "background-color": g().hex }}
                      />
                    )}
                  </Show>
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-sm font-medium text-gray-900">
                      {assigned()?.name ?? "Unassigned"}
                    </p>
                    <p class="text-xs text-gray-500">
                      {region.count} piece{region.count === 1 ? "" : "s"}
                      <Show when={region.source_class}>
                        {" · "}
                        {region.source_class}
                      </Show>
                    </p>
                  </div>
                  <Show when={region.glass_color_id == null}>
                    <Badge variant="warning" class="rounded-full">
                      Unassigned
                    </Badge>
                  </Show>
                </button>

                <Show when={isActive()}>
                  <div class="border-t border-gray-100 p-3">
                    <SwatchPicker
                      swatches={glassSwatches()}
                      selectedId={region.glass_color_id}
                      onSelect={(id) =>
                        props.onAssignGroupColor(groupKey, id)
                      }
                      searchPlaceholder="Search glass colors..."
                    />
                    <div class="mt-3 flex flex-wrap gap-2 border-t border-gray-100 pt-3">
                      <Show when={hasPieceSelection()}>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => props.onMovePiecesToGroup(groupKey)}
                        >
                          Move selected here
                        </Button>
                      </Show>
                      <For
                        each={groups().filter(([k]) => k !== groupKey)}
                      >
                        {([otherKey, otherRegion]) => (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => props.onMergeInto(otherKey)}
                            title={`Merge "${groupKey}" into another group`}
                          >
                            Merge into {otherRegion.source_class ?? otherKey}
                          </Button>
                        )}
                      </For>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => props.onMarkGroupAsGrout(groupKey)}
                      >
                        Mark as grout
                      </Button>
                    </div>
                  </div>
                </Show>
              </div>
            );
          }}
        </For>
      </div>

      <div class="flex flex-col gap-2 border-t border-gray-200 pt-4">
        <h3 class="text-sm font-semibold text-gray-900">Grout</h3>
        <div class="flex items-center gap-3 px-1">
          <Show
            when={
              props.manifest.grout_region.grout_id != null
                ? props.grouts.find(
                    (g) => g.id === props.manifest.grout_region.grout_id,
                  )
                : undefined
            }
            fallback={
              <Badge variant="warning" class="rounded-full">
                Unassigned
              </Badge>
            }
          >
            {(g) => (
              <span class="flex items-center gap-2 text-sm text-gray-700">
                <span
                  class="h-6 w-6 rounded border border-black/10"
                  style={{ "background-color": g().hex }}
                />
                {g().name}
              </span>
            )}
          </Show>
          <span class="text-xs text-gray-500">
            {props.manifest.grout_region.count} piece
            {props.manifest.grout_region.count === 1 ? "" : "s"}
          </span>
        </div>
        <SwatchPicker
          swatches={groutSwatches()}
          selectedId={props.manifest.grout_region.grout_id}
          onSelect={props.onAssignGroutColor}
          searchPlaceholder="Search grouts..."
        />
      </div>
    </div>
  );
}
