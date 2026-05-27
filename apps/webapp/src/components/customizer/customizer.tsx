import {
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  Show,
} from "solid-js";
import { useMutation } from "@tanstack/solid-query";
import type {
  CatalogItem,
  ColorOverrides,
  GlassColor,
  Grout,
  GET,
} from "@glassact/data";
import { Alert, AlertDescription, Badge, Button } from "@glassact/ui";
import { postBakeOpts } from "../../queries/customize";
import { CustomizerCanvas } from "./customizer-canvas";
import { ControlPanel } from "./control-panel";
import { PricingWarningDialog } from "./pricing-warning-dialog";
import {
  buildPieceSourceMap,
  resolvePieceHex,
  totalCustomPieces,
  type GlassById,
  type Selection,
} from "./resolution";

interface CustomizerProps {
  item: GET<CatalogItem>;
  svgText: string;
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
}

interface PersistedState {
  overrides: ColorOverrides;
  width: number;
}

export function Customizer(props: CustomizerProps) {
  const storageKey = `gac:customizer:${props.item.uuid}`;
  const aspect = props.item.default_height / props.item.default_width;
  const minWidth = Math.max(
    props.item.min_width,
    props.item.min_height / aspect,
  );

  const loadPersisted = (): PersistedState => {
    try {
      const raw = localStorage.getItem(storageKey);
      if (raw) return JSON.parse(raw) as PersistedState;
    } catch {
      /* ignore */
    }
    return { overrides: {}, width: props.item.default_width };
  };
  const persisted = loadPersisted();

  const [overrides, setOverrides] = createSignal<ColorOverrides>(
    persisted.overrides ?? {},
  );
  const [past, setPast] = createSignal<ColorOverrides[]>([]);
  const [future, setFuture] = createSignal<ColorOverrides[]>([]);

  const [mode, setMode] = createSignal<"group" | "piece">("group");
  const [pricingOpen, setPricingOpen] = createSignal(false);
  const [pieceModeAcknowledged, setPieceModeAcknowledged] = createSignal(false);

  const [selection, setSelection] = createSignal<Selection | null>(null);
  const [hoverGlassId, setHoverGlassId] = createSignal<number | null>(null);
  const [hoveredRegion, setHoveredRegion] = createSignal<string | null>(null);

  const [width, setWidth] = createSignal(persisted.width ?? props.item.default_width);
  const height = createMemo(() => width() * aspect);

  const manifest = createMemo(
    () => props.item.manifest ?? { view_box: "0 0 0 0", regions: {} },
  );
  const glassById = createMemo<GlassById>(
    () => new Map(props.glassColors.map((g) => [g.id, g])),
  );
  const pieceSource = createMemo(() => buildPieceSourceMap(manifest()));

  const groutHex = createMemo(() => {
    const id = overrides().background?.grout_id;
    if (id == null) return null;
    return props.grouts.find((g) => g.id === id)?.hex ?? null;
  });

  const usedGlassIds = createMemo(() => {
    const o = overrides();
    const ids = new Set<number>();
    for (const r of Object.values(o.regions ?? {})) ids.add(r.glass_color_id);
    for (const p of Object.values(o.pieces ?? {})) ids.add(p.glass_color_id);
    return [...ids];
  });

  const selectedGlassId = createMemo<number | null>(() => {
    const sel = selection();
    if (!sel) return null;
    const o = overrides();
    if (sel.type === "region") return o.regions?.[sel.sourceHex]?.glass_color_id ?? null;
    return o.pieces?.[sel.pieceId]?.glass_color_id ?? null;
  });

  // Resolution with live hover-preview layered on top of committed overrides.
  function resolveHex(pieceId: string, sourceHex: string): string {
    const hov = hoverGlassId();
    const sel = selection();
    const o = overrides();
    if (hov != null && sel) {
      const previewHex = glassById().get(hov)?.hex;
      if (previewHex) {
        if (sel.type === "piece" && sel.pieceId === pieceId) return previewHex;
        if (
          sel.type === "region" &&
          sel.sourceHex === sourceHex &&
          !o.pieces?.[pieceId]
        ) {
          return previewHex;
        }
      }
    }
    return resolvePieceHex(pieceId, sourceHex, o, glassById());
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

  function selectRegion(sourceHex: string) {
    setSelection({ type: "region", sourceHex });
  }

  function onPieceClick(pieceId: string, sourceHex: string) {
    if (mode() === "piece") {
      setSelection({ type: "piece", pieceId, sourceHex });
    } else {
      setSelection({ type: "region", sourceHex });
    }
  }

  function assignGlass(glassId: number) {
    const sel = selection();
    if (!sel) return;
    const o = overrides();
    if (sel.type === "region") {
      commit({
        ...o,
        regions: { ...(o.regions ?? {}), [sel.sourceHex]: { glass_color_id: glassId } },
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

  // Autosave to localStorage so a refresh/navigation doesn't lose work.
  createEffect(() => {
    const state: PersistedState = { overrides: overrides(), width: width() };
    try {
      localStorage.setItem(storageKey, JSON.stringify(state));
    } catch {
      /* ignore quota errors */
    }
  });

  // Warn before leaving with unsaved changes.
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
    bake.mutate({
      uuid: props.item.uuid,
      body: {
        scale_factor: width() / props.item.default_width,
        width: width(),
        height: height(),
        color_overrides: overrides(),
      },
    });
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

      <Show when={bake.isSuccess && bake.data}>
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
            resolveHex={resolveHex}
            groutHex={groutHex()}
            selectedPieceId={
              selection()?.type === "piece"
                ? (selection() as { pieceId: string }).pieceId
                : null
            }
            highlightedRegion={hoveredRegion()}
            onPieceClick={onPieceClick}
            onPieceHover={(_, src) => setHoveredRegion(src)}
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
            usedGlassIds={usedGlassIds()}
            backgroundGroutId={overrides().background?.grout_id ?? null}
            width={width()}
            height={height()}
            minWidth={minWidth}
            minHeight={props.item.min_height}
            onSelectRegion={selectRegion}
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
