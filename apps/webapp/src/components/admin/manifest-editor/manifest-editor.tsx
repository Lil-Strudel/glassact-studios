import { createMemo, createSignal, For, Show } from "solid-js";
import type { GlassColor, Grout, Manifest, GET } from "@glassact/data";
import { Alert, AlertDescription, Badge, Button } from "@glassact/ui";
import {
  CustomizerCanvas,
  GROUT_REGION_KEY,
  buildGroutPieceIds,
  buildPieceSourceMap,
  resolvePieceHex,
  type GlassById,
} from "../../customizer/shared";
import { GroupList } from "./group-list";
import {
  assignGroupColor,
  assignGroutColor,
  markGroupAsGrout,
  mergeGroups,
  movePiecesToGrout,
  movePiecesToGroup,
  splitGroup,
  unmarkGroutPieces,
} from "./manifest-ops";

interface ManifestEditorProps {
  structureSvg: string;
  manifest: Manifest;
  warnings: string[];
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
  onManifestChange: (manifest: Manifest) => void;
}

type EditMode = "group" | "piece";

// The in-editor manifest IS the source of truth for colors (no separate
// overrides layer), so colors resolve via resolvePieceHex with an empty override
// object. Unassigned groups fall back to the neutral "#cccccc" and are visually
// flagged in the panel and canvas.
export function ManifestEditor(props: ManifestEditorProps) {
  const [mode, setMode] = createSignal<EditMode>("group");
  // A glass group key, or GROUT_REGION_KEY when the grout region is open.
  const [activeRegionKey, setActiveRegionKey] = createSignal<string | null>(
    null,
  );
  const [selectedPieceIds, setSelectedPieceIds] = createSignal<string[]>([]);
  const [hoveredRegion, setHoveredRegion] = createSignal<string | null>(null);

  const glassById = createMemo<GlassById>(
    () => new Map(props.glassColors.map((g) => [g.id, g])),
  );

  const pieceSource = createMemo(() => buildPieceSourceMap(props.manifest));
  const groutPieceIds = createMemo(() => buildGroutPieceIds(props.manifest));

  const groutHex = createMemo(() => {
    const id = props.manifest.grout_region.grout_id;
    if (id == null) return null;
    return props.grouts.find((g) => g.id === id)?.hex ?? null;
  });

  // Whether the current piece selection reaches into the grout region, which is
  // what enables the "move back to glass" action.
  const selectionHasGroutPieces = createMemo(() => {
    const grout = new Set(groutPieceIds());
    return selectedPieceIds().some((id) => grout.has(id));
  });

  function resolveHex(pieceId: string, groupKey: string): string {
    return resolvePieceHex(pieceId, groupKey, {}, props.manifest, glassById());
  }

  function update(next: Manifest) {
    props.onManifestChange(next);
  }

  function onPieceClick(pieceId: string, regionKey: string) {
    if (mode() === "piece") {
      const current = selectedPieceIds();
      setSelectedPieceIds(
        current.includes(pieceId)
          ? current.filter((id) => id !== pieceId)
          : [...current, pieceId],
      );
    } else {
      setActiveRegionKey(regionKey);
    }
  }

  function selectRegion(regionKey: string) {
    setActiveRegionKey((prev) => (prev === regionKey ? null : regionKey));
  }

  function assignGroup(groupKey: string, glassColorId: number) {
    update(assignGroupColor(props.manifest, groupKey, glassColorId));
  }

  function assignGrout(groutId: number) {
    update(assignGroutColor(props.manifest, groutId));
  }

  function markGrout(groupKey: string) {
    update(markGroupAsGrout(props.manifest, groupKey));
    if (activeRegionKey() === groupKey) setActiveRegionKey(GROUT_REGION_KEY);
  }

  function mergeInto(targetKey: string) {
    const source = activeRegionKey();
    if (!source || source === GROUT_REGION_KEY) return;
    update(mergeGroups(props.manifest, source, targetKey));
    setActiveRegionKey(targetKey);
  }

  function moveSelectedToGroup(targetKey: string) {
    const ids = selectedPieceIds();
    if (ids.length === 0) return;
    update(movePiecesToGroup(props.manifest, ids, targetKey));
    setSelectedPieceIds([]);
  }

  function moveSelectedToGrout() {
    const ids = selectedPieceIds();
    if (ids.length === 0) return;
    update(movePiecesToGrout(props.manifest, ids));
    setSelectedPieceIds([]);
  }

  function splitSelected() {
    const ids = selectedPieceIds();
    if (ids.length === 0) return;
    const { manifest, newKey } = splitGroup(props.manifest, ids);
    update(manifest);
    setSelectedPieceIds([]);
    setActiveRegionKey(newKey);
  }

  // The recovery path when ingest seeds a glass detail into the grout region.
  function moveSelectedToGlass() {
    const ids = selectedPieceIds();
    if (ids.length === 0) return;
    const { manifest, newKey } = unmarkGroutPieces(props.manifest, ids);
    update(manifest);
    setSelectedPieceIds([]);
    setActiveRegionKey(newKey);
  }

  // Highlight the active region, the hovered one, or the in-progress selection.
  const highlightedRegion = createMemo(
    () => hoveredRegion() ?? activeRegionKey(),
  );

  const unassignedCount = createMemo(
    () =>
      Object.values(props.manifest.glass_regions).filter(
        (r) => r.glass_color_id == null,
      ).length + (props.manifest.grout_region.grout_id == null ? 1 : 0),
  );

  return (
    <div class="flex flex-col gap-3">
      <Show when={props.warnings.length > 0}>
        <Alert>
          <AlertDescription>
            <p class="mb-1 font-medium">Analysis notes</p>
            <ul class="list-inside list-disc text-sm">
              <For each={props.warnings}>{(w) => <li>{w}</li>}</For>
            </ul>
          </AlertDescription>
        </Alert>
      </Show>

      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <Button
            type="button"
            size="sm"
            variant={mode() === "group" ? "default" : "outline"}
            onClick={() => {
              setMode("group");
              setSelectedPieceIds([]);
              setHoveredRegion(null);
            }}
          >
            Group mode
          </Button>
          <Button
            type="button"
            size="sm"
            variant={mode() === "piece" ? "default" : "outline"}
            onClick={() => {
              setMode("piece");
              setActiveRegionKey(null);
              setHoveredRegion(null);
            }}
          >
            Piece mode
          </Button>
        </div>
        <Show
          when={unassignedCount() === 0}
          fallback={
            <Badge variant="warning" class="rounded-full">
              {unassignedCount()} region
              {unassignedCount() === 1 ? "" : "s"} unassigned
            </Badge>
          }
        >
          <Badge variant="secondary" class="rounded-full">
            All regions assigned
          </Badge>
        </Show>
      </div>

      <div class="flex min-h-[28rem] flex-col gap-4 lg:flex-row">
        <div class="min-h-[24rem] flex-1">
          <CustomizerCanvas
            svgText={props.structureSvg}
            pieceSource={pieceSource()}
            groutPieceIds={groutPieceIds()}
            resolveHex={resolveHex}
            groutHex={groutHex()}
            selectedPieceIds={selectedPieceIds()}
            highlightedRegion={highlightedRegion()}
            groutSelectable
            onPieceClick={onPieceClick}
            onPieceHover={(_, src) => {
              if (mode() === "group") setHoveredRegion(src);
            }}
          />
        </div>

        <div class="w-full overflow-y-auto lg:w-96 lg:shrink-0">
          <GroupList
            manifest={props.manifest}
            glassColors={props.glassColors}
            grouts={props.grouts}
            activeRegionKey={activeRegionKey()}
            selectedPieceIds={selectedPieceIds()}
            selectionHasGroutPieces={selectionHasGroutPieces()}
            onSelectRegion={selectRegion}
            onHoverRegion={setHoveredRegion}
            onAssignGroupColor={assignGroup}
            onAssignGroutColor={assignGrout}
            onMarkGroupAsGrout={markGrout}
            onMergeInto={mergeInto}
            onMovePiecesToGroup={moveSelectedToGroup}
            onMovePiecesToGrout={moveSelectedToGrout}
            onMovePiecesToGlass={moveSelectedToGlass}
            onSplitSelected={splitSelected}
            onClearPieceSelection={() => setSelectedPieceIds([])}
          />
        </div>
      </div>
    </div>
  );
}
