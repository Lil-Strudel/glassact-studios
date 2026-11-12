import { For, Show, createMemo, type ParentProps } from "solid-js";
import type { GET, InlayWithInfo, SandblastBackground } from "@glassact/data";
import { Button } from "@glassact/ui";
import type { StoneEditor } from "./create-stone-editor";
import { newElementId, type Point } from "./stone-geometry";
import { CompoundPicker } from "./compound-picker";
import { CustomizeInlayLink } from "./customize-inlay-link";
import { RelatedInlays } from "./related-inlays";
import type { AddableCatalogItem } from "./use-add-catalog-inlay";
import { CatalogItemPicker } from "../catalog/catalog-item-picker";

interface InlayPanelProps {
  editor: StoneEditor;
  projectUuid: string;
  inlays: InlayWithInfo[];
  backgroundsByUuid: Map<string, GET<SandblastBackground>>;
  canAddInlays: boolean;
  isAddingInlay: boolean;
  onAddInlay: (item: AddableCatalogItem, anchor?: Point) => void;
  onAddCompound: (compoundUuid: string) => Promise<void>;
}

export function InlayPanel(props: InlayPanelProps) {
  const placedUuids = createMemo(
    () =>
      new Set(
        props.editor.draft.layout.elements.flatMap((el) =>
          el.type === "inlay" ? [el.inlay_uuid] : [],
        ),
      ),
  );

  const unplaced = createMemo(() =>
    props.inlays.filter((inlay) => !placedUuids().has(inlay.uuid)),
  );
  const placed = createMemo(() =>
    props.inlays.filter((inlay) => placedUuids().has(inlay.uuid)),
  );

  function place(inlay: InlayWithInfo) {
    const draft = props.editor.draft;
    const offset = (unplaced().indexOf(inlay) % 5) * 1.5;
    props.editor.addElements([
      {
        id: newElementId(),
        type: "inlay",
        inlay_uuid: inlay.uuid,
        x: draft.width / 2 + offset,
        y: draft.height / 2 + offset,
        rotation: 0,
        group_id: null,
      },
    ]);
  }

  function selectPlaced(inlay: InlayWithInfo) {
    const el = props.editor.draft.layout.elements.find(
      (e) => e.type === "inlay" && e.inlay_uuid === inlay.uuid,
    );
    if (el) props.editor.select(el.id, false);
  }

  return (
    <div class="flex flex-col gap-6">
      <Show when={props.canAddInlays}>
        <RelatedInlays
          editor={props.editor}
          backgroundsByUuid={props.backgroundsByUuid}
          isAdding={props.isAddingInlay}
          onAdd={props.onAddInlay}
        />
      </Show>

      <section>
        <h3 class="text-sm font-medium text-gray-900">Not on the stone yet</h3>
        <Show
          when={unplaced().length > 0}
          fallback={
            <p class="mt-2 text-xs text-gray-500">
              Every inlay on this project is placed.
            </p>
          }
        >
          <ul class="mt-2 flex flex-col gap-2">
            <For each={unplaced()}>
              {(inlay) => (
                <InlayRow inlay={inlay} projectUuid={props.projectUuid}>
                  <Show
                    when={inlay.width != null}
                    fallback={<span class="text-xs text-gray-400">No size yet</span>}
                  >
                    <Button size="sm" variant="outline" onClick={() => place(inlay)}>
                      Place
                    </Button>
                  </Show>
                </InlayRow>
              )}
            </For>
          </ul>
        </Show>
      </section>

      <Show when={placed().length > 0}>
        <section>
          <h3 class="text-sm font-medium text-gray-900">On the stone</h3>
          <ul class="mt-2 flex flex-col gap-2">
            <For each={placed()}>
              {(inlay) => (
                <InlayRow inlay={inlay} projectUuid={props.projectUuid}>
                  <Button size="sm" variant="ghost" onClick={() => selectPlaced(inlay)}>
                    Select
                  </Button>
                </InlayRow>
              )}
            </For>
          </ul>
        </section>
      </Show>

      <Show when={props.canAddInlays}>
        <CatalogItemPicker
          title="Add from the catalog"
          description="Click an inlay to add it to the project and place it on the stone."
          isDisabled={props.isAddingInlay}
          onPick={(item) => props.onAddInlay(item)}
        />
      </Show>

      <Show when={props.canAddInlays}>
        <section>
          <h3 class="text-sm font-medium text-gray-900">Compound inlays</h3>
          <p class="mt-1 text-xs text-gray-500">
            Arranged sets. Adding one puts every inlay in it on the project and
            groups them on the stone; ungroup to move them separately.
          </p>
          <CompoundPicker onAdd={props.onAddCompound} class="mt-2" />
        </section>
      </Show>
    </div>
  );
}

function InlayRow(props: ParentProps<{ inlay: InlayWithInfo; projectUuid: string }>) {
  return (
    <li class="flex items-center gap-3 rounded-md border border-gray-200 p-2">
      <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded bg-gray-100">
        <Show when={props.inlay.preview_url}>
          {(url) => <img src={url()} alt="" class="max-h-full max-w-full object-contain" />}
        </Show>
      </div>
      <div class="min-w-0 flex-1">
        <p class="truncate text-sm text-gray-900">{props.inlay.name}</p>
        <Show when={props.inlay.width != null}>
          <p class="text-xs text-gray-500">
            {props.inlay.width}″ × {props.inlay.height}″
          </p>
        </Show>
        <CustomizeInlayLink
          projectUuid={props.projectUuid}
          inlay={props.inlay}
          variant="link"
          class="h-auto px-0 py-0 text-xs"
        />
      </div>
      {props.children}
    </li>
  );
}
