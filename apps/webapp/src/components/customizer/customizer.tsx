import {
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  Show,
  untrack,
} from "solid-js";
import { useMutation } from "@tanstack/solid-query";
import type {
  BakeResult,
  CatalogItem,
  ColorOverrides,
  GlassColor,
  Grout,
  GET,
} from "@glassact/data";
import { Alert, AlertDescription, Badge, Button } from "@glassact/ui";
import { postBakeOpts } from "../../queries/customize";
import { useGranitePreference } from "../../hooks/use-granite-preference";
import { ControlPanel } from "./control-panel";
import { PricingWarningDialog } from "./pricing-warning-dialog";
import {
  CustomizerCanvas,
  buildGroutPieceIds,
  buildPieceSourceMap,
  groupGlassId,
  resolvePieceHex,
  totalCustomPieces,
  type GlassById,
  type Selection,
} from "./shared";

interface CustomizerProps {
  item: GET<CatalogItem>;
  svgText: string;
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
  onBakeComplete?: (result: BakeResult) => void;
  // Where to start. Re-customizing an existing inlay seeds from that inlay's
  // current coloring; otherwise the catalog defaults are used.
  initialState?: InitialState;
}

interface InitialState {
  overrides: ColorOverrides;
  width: number;
}

export function Customizer(props: CustomizerProps) {
  const aspect = untrack(
    () => props.item.default_height / props.item.default_width,
  );
  const minWidth = Math.max(
    untrack(() => props.item.min_width),
    untrack(() => props.item.min_height) / aspect,
  );

  // Where to start: a re-customize session seeds from the inlay's current
  // coloring (props.initialState); otherwise from catalog defaults. Nothing is
  // persisted between sessions.
  const initial: InitialState = untrack(
    () =>
      props.initialState ?? { overrides: {}, width: props.item.default_width },
  );

  const [overrides, setOverrides] = createSignal<ColorOverrides>(
    initial.overrides ?? {},
  );
  const [past, setPast] = createSignal<ColorOverrides[]>([]);
  const [future, setFuture] = createSignal<ColorOverrides[]>([]);

  const [mode, setMode] = createSignal<"group" | "piece">("group");
  const [pricingOpen, setPricingOpen] = createSignal(false);
  const [pieceModeAcknowledged, setPieceModeAcknowledged] = createSignal(false);

  const [selection, setSelection] = createSignal<Selection | null>(null);
  const [hoverGlassId, setHoverGlassId] = createSignal<number | null>(null);
  const [hoveredRegion, setHoveredRegion] = createSignal<string | null>(null);

  const [width, setWidth] = createSignal(
    initial.width ?? untrack(() => props.item.default_width),
  );
  const height = createMemo(() => width() * aspect);

  // The granite backdrop (the "stone" the inlay sits on) is a viewing
  // preference, independent of the saved/baked coloring. It is shared with the
  // catalog so the stone a dealership browsed on carries into the customizer.
  const { graniteKey, granite, setGraniteKey } = useGranitePreference();

  const manifest = createMemo(
    () =>
      props.item.manifest ?? {
        view_box: "0 0 0 0",
        grout_region: { grout_id: null, piece_ids: [], count: 0 },
        glass_regions: {},
      },
  );
  const glassById = createMemo<GlassById>(
    () => new Map(props.glassColors.map((g) => [g.id, g])),
  );
  const pieceSource = createMemo(() => buildPieceSourceMap(manifest()));
  const groutPieceIds = createMemo(() => buildGroutPieceIds(manifest()));

  const groutHex = createMemo(() => {
    const id = overrides().background?.grout_id;
    if (id == null) return null;
    return props.grouts.find((g) => g.id === id)?.hex ?? null;
  });

  const selectedGlassId = createMemo<number | null>(() => {
    const sel = selection();
    if (!sel) return null;
    const o = overrides();
    if (sel.type === "group") return groupGlassId(sel.groupKey, o, manifest());
    return (
      o.pieces?.[sel.pieceId]?.glass_color_id ??
      groupGlassId(sel.groupKey, o, manifest())
    );
  });

  // Resolution with live hover-preview layered on top of committed overrides.
  function resolveHex(pieceId: string, groupKey: string): string {
    const hov = hoverGlassId();
    const sel = selection();
    const o = overrides();
    if (hov != null && sel) {
      const previewHex = glassById().get(hov)?.hex;
      if (previewHex) {
        if (sel.type === "piece" && sel.pieceId === pieceId) return previewHex;
        if (
          sel.type === "group" &&
          sel.groupKey === groupKey &&
          !o.pieces?.[pieceId]
        ) {
          return previewHex;
        }
      }
    }
    return resolvePieceHex(pieceId, groupKey, o, manifest(), glassById());
  }

  function commit(next: ColorOverrides) {
    setPast([...past(), overrides()]);
    setFuture([]);
    setOverrides(next);
  }

  function undo() {
    const p = past();
    if (!p.length) return;
    setFuture([overrides(), ...future()]);
    setOverrides(p[p.length - 1]);
    setPast(p.slice(0, -1));
  }

  function redo() {
    const f = future();
    if (!f.length) return;
    setPast([...past(), overrides()]);
    setOverrides(f[0]);
    setFuture(f.slice(1));
  }

  function requestMode(next: "group" | "piece") {
    if (next === "piece" && !pieceModeAcknowledged()) {
      setPricingOpen(true);
      return;
    }
    setMode(next);
  }

  // Selecting the already-selected group/piece toggles it back off so the color
  // grid collapses ("clicking no color deselects").
  function selectGroup(groupKey: string) {
    const sel = selection();
    if (sel?.type === "group" && sel.groupKey === groupKey) {
      setSelection(null);
      return;
    }
    setSelection({ type: "group", groupKey });
  }

  function onPieceClick(pieceId: string, groupKey: string) {
    if (mode() === "piece") {
      const sel = selection();
      if (sel?.type === "piece" && sel.pieceId === pieceId) {
        setSelection(null);
        return;
      }
      setSelection({ type: "piece", pieceId, groupKey });
    } else {
      selectGroup(groupKey);
    }
  }

  function deselect() {
    setSelection(null);
  }

  function assignGlass(glassId: number) {
    const sel = selection();
    if (!sel) return;
    const o = overrides();
    if (sel.type === "group") {
      commit({
        ...o,
        groups: { ...(o.groups ?? {}), [sel.groupKey]: { glass_color_id: glassId } },
      });
    } else {
      commit({
        ...o,
        pieces: { ...(o.pieces ?? {}), [sel.pieceId]: { glass_color_id: glassId } },
      });
    }
  }

  function resetPiece(pieceId: string) {
    const o = overrides();
    const pieces = { ...(o.pieces ?? {}) };
    delete pieces[pieceId];
    commit({ ...o, pieces });
  }

  function selectGrout(groutId: number) {
    commit({ ...overrides(), background: { grout_id: groutId } });
  }

  function resetAll() {
    commit({});
    setSelection(null);
  }

  function setWidthClamped(w: number) {
    if (Number.isNaN(w)) return;
    setWidth(Math.max(minWidth, w));
  }

  const isDirty = createMemo(
    () =>
      past().length > 0 ||
      Object.keys(overrides()).length > 0 ||
      width() !== props.item.default_width,
  );

  // Warn before leaving with unsaved changes (work is kept only in memory).
  createEffect(() => {
    const dirty = isDirty();
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };
    if (dirty) window.addEventListener("beforeunload", handler);
    onCleanup(() => window.removeEventListener("beforeunload", handler));
  });

  const bake = useMutation(() => postBakeOpts());

  function onSave() {
    bake.mutate(
      {
        uuid: props.item.uuid,
        body: {
          scale_factor: width() / props.item.default_width,
          width: width(),
          height: height(),
          color_overrides: overrides(),
        },
      },
      {
        onSuccess(result) {
          props.onBakeComplete?.(result);
        },
      },
    );
  }

  const customPieces = createMemo(() => totalCustomPieces(overrides()));

  return (
    <div class="flex h-[calc(100vh-8rem)] flex-col gap-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-xl font-bold text-gray-900">
            Customize: {props.item.name}
          </h1>
          <p class="text-xs text-gray-500">{props.item.catalog_code}</p>
        </div>

        <div class="flex items-center gap-2">
          <Show when={customPieces() > 0}>
            <Badge variant="warning" class="rounded-full">
              {customPieces()} custom piece{customPieces() === 1 ? "" : "s"} · may
              affect price
            </Badge>
          </Show>
          <Button
            variant="outline"
            size="sm"
            disabled={past().length === 0}
            onClick={undo}
          >
            Undo
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={future().length === 0}
            onClick={redo}
          >
            Redo
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!isDirty()}
            onClick={resetAll}
          >
            Reset
          </Button>
          <Button size="sm" disabled={bake.isPending} onClick={onSave}>
            {bake.isPending ? "Saving..." : "Save"}
          </Button>
        </div>
      </div>

      <Show when={!props.onBakeComplete && bake.isSuccess && bake.data}>
        {(result) => (
          <Alert variant="success">
            <AlertDescription class="flex items-center justify-between">
              <span>Design saved.</span>
              <Button
                as="a"
                variant="link"
                size="sm"
                href={result().design_asset_url}
                target="_blank"
                rel="noreferrer"
              >
                View baked SVG
              </Button>
            </AlertDescription>
          </Alert>
        )}
      </Show>
      <Show when={bake.isError}>
        <Alert variant="destructive">
          <AlertDescription class="flex items-center justify-between">
            <span>
              {bake.error instanceof Error ? bake.error.message : "Failed to save."}
            </span>
            <Button size="sm" variant="outline" onClick={onSave}>
              Retry
            </Button>
          </AlertDescription>
        </Alert>
      </Show>

      <div class="flex min-h-0 flex-1 flex-col gap-4 lg:flex-row">
        <div class="min-h-[24rem] flex-1">
          <CustomizerCanvas
            svgText={props.svgText}
            pieceSource={pieceSource()}
            groutPieceIds={groutPieceIds()}
            resolveHex={resolveHex}
            groutHex={groutHex()}
            selectedPieceIds={
              selection()?.type === "piece"
                ? [(selection() as { pieceId: string }).pieceId]
                : []
            }
            highlightedRegion={hoveredRegion()}
            granite={granite()}
            graniteKey={graniteKey()}
            onSelectGranite={setGraniteKey}
            onPieceClick={onPieceClick}
            onPieceHover={(_, groupKey) => setHoveredRegion(groupKey)}
            onDeselect={deselect}
          />
        </div>

        <div class="w-full overflow-y-auto lg:w-96 lg:shrink-0">
          <ControlPanel
            mode={mode()}
            onRequestMode={requestMode}
            manifest={manifest()}
            glassColors={props.glassColors}
            glassById={glassById()}
            grouts={props.grouts}
            overrides={overrides()}
            selection={selection()}
            selectedGlassId={selectedGlassId()}
            width={width()}
            height={height()}
            minWidth={minWidth}
            minHeight={props.item.min_height}
            onSelectGroup={selectGroup}
            onRegionHover={setHoveredRegion}
            onAssignGlass={assignGlass}
            onHoverGlass={setHoverGlassId}
            onResetPiece={resetPiece}
            onSelectGrout={selectGrout}
            onWidthChange={setWidthClamped}
          />
        </div>
      </div>

      <PricingWarningDialog
        open={pricingOpen()}
        onContinue={(dontRemind) => {
          if (dontRemind) setPieceModeAcknowledged(true);
          setPricingOpen(false);
          setMode("piece");
        }}
        onCancel={() => setPricingOpen(false)}
      />
    </div>
  );
}
