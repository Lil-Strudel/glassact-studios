import { For, Show, createMemo } from "solid-js";
import type { CatalogItem, GET, StoneBackgroundElement } from "@glassact/data";
import { Button, cn } from "@glassact/ui";
import { useBackgroundItems } from "./use-background-items";
import { DownloadSandblastButton } from "../catalog/download-sandblast-button";
import type { StoneEditor } from "./create-stone-editor";
import { newElementId } from "./stone-geometry";

interface BackgroundPanelProps {
  editor: StoneEditor;
  isEditable: boolean;
}

export function BackgroundPanel(props: BackgroundPanelProps) {
  const backgrounds = useBackgroundItems();

  const placed = createMemo(() =>
    props.editor.draft.layout.elements.filter(
      (el): el is StoneBackgroundElement => el.type === "background",
    ),
  );

  const selectedBackground = createMemo(() => {
    const selected = props.editor.selectedElements();
    return selected.length === 1 && selected[0].type === "background"
      ? selected[0]
      : null;
  });

  // The usual case is one background framing the whole face, so picking one
  // replaces the current background unless the user is editing a second one.
  function choose(item: GET<CatalogItem>) {
    const draft = props.editor.draft;
    const replacing = selectedBackground() ?? placed()[0];
    const width = replacing?.width ?? draft.width * 0.9;
    const element: StoneBackgroundElement = {
      id: newElementId(),
      type: "background",
      catalog_item_uuid: item.uuid,
      x: replacing?.x ?? draft.width / 2,
      y: replacing?.y ?? draft.height / 2,
      rotation: replacing?.rotation ?? 0,
      width,
      flip_x: replacing?.flip_x,
      group_id: replacing?.group_id ?? null,
    };
    props.editor.update((d) => {
      const index = replacing
        ? d.layout.elements.findIndex((el) => el.id === replacing.id)
        : -1;
      if (index >= 0) d.layout.elements[index] = element;
      else d.layout.elements.unshift(element);
    });
    props.editor.selectIds([element.id]);
  }

  function addAnother(item: GET<CatalogItem>) {
    const draft = props.editor.draft;
    props.editor.addElements([
      {
        id: newElementId(),
        type: "background",
        catalog_item_uuid: item.uuid,
        x: draft.width / 2,
        y: draft.height / 2,
        rotation: 0,
        width: draft.width * 0.5,
        group_id: null,
      },
    ]);
  }

  function updateSelected(mutate: (el: StoneBackgroundElement) => void) {
    const target = selectedBackground();
    if (!target) return;
    props.editor.update((d) => {
      const el = d.layout.elements.find((e) => e.id === target.id);
      if (el?.type === "background") mutate(el);
    });
  }

  return (
    <div class="flex flex-col gap-6">
      <Show when={placed().length > 0}>
        <section class="flex flex-col gap-2">
          <h3 class="text-sm font-medium text-gray-900">On the stone</h3>
          <For each={placed()}>
            {(el) => {
              const item = () => backgrounds.byUuid().get(el.catalog_item_uuid);
              return (
                <div
                  class={cn(
                    "flex items-center gap-2 rounded-md border p-2",
                    selectedBackground()?.id === el.id ? "border-blue-500" : "border-gray-200",
                  )}
                >
                  <button
                    type="button"
                    class="min-w-0 flex-1 truncate text-left text-sm text-gray-900"
                    onClick={() => props.editor.selectIds([el.id])}
                  >
                    {item()?.name ?? "Background"}
                    <span class="block text-xs text-gray-500">{item()?.catalog_code}</span>
                  </button>
                  <DownloadSandblastButton catalogItemUuid={el.catalog_item_uuid} />
                </div>
              );
            }}
          </For>
        </section>
      </Show>

      <Show when={props.isEditable && selectedBackground()}>
        {(el) => (
          <section class="flex flex-col gap-2">
            <h3 class="text-sm font-medium text-gray-900">Selected background</h3>
            <label class="flex items-center justify-between text-xs text-gray-600">
              Width (inches)
              <input
                type="number"
                min="0.25"
                max="240"
                step="0.25"
                value={el().width}
                onChange={(e) => {
                  const value = Number(e.currentTarget.value);
                  if (value > 0) updateSelected((b) => (b.width = Math.min(240, value)));
                }}
                class="w-24 rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
            <div class="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => updateSelected((b) => (b.flip_x = !b.flip_x))}>
                Flip
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  const { width, height } = props.editor.draft;
                  updateSelected((b) => {
                    b.x = width / 2;
                    b.y = height / 2;
                    b.rotation = 0;
                    b.width = width * 0.9;
                  });
                }}
              >
                Fit to stone
              </Button>
              <Button size="sm" variant="outline" onClick={() => props.editor.removeSelected()}>
                Remove
              </Button>
            </div>
          </section>
        )}
      </Show>

      <Show when={props.isEditable}>
        <section>
          <h3 class="text-sm font-medium text-gray-900">Sandblast backgrounds</h3>
          <p class="mt-1 text-xs text-gray-500">
            Click to use as the stone's background. Shift-click to add another
            one alongside it.
          </p>
          <Show when={backgrounds.query.isLoading}>
            <div class="mt-2 h-32 animate-pulse rounded-md bg-gray-100" />
          </Show>
          <div class="mt-2 grid grid-cols-3 gap-2">
            <For each={backgrounds.query.data?.items ?? []}>
              {(item) => (
                <button
                  type="button"
                  title={`${item.name} (${item.catalog_code})`}
                  class="flex aspect-square items-center justify-center rounded-md border border-gray-200 bg-white p-1.5 hover:border-blue-500"
                  onClick={(e) => (e.shiftKey ? addAnother(item) : choose(item))}
                >
                  <img
                    src={item.svg_url}
                    alt={item.name}
                    loading="lazy"
                    class="max-h-full max-w-full object-contain brightness-0"
                  />
                </button>
              )}
            </For>
          </div>
        </section>
      </Show>
    </div>
  );
}
