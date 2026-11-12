import { For, Show, createMemo } from "solid-js";
import type {
  GET,
  RelatedCatalogItem,
  SandblastBackground,
  StoneBackgroundElement,
} from "@glassact/data";
import type { StoneEditor } from "./create-stone-editor";
import type { Point } from "./stone-geometry";

interface RelatedInlaysProps {
  editor: StoneEditor;
  backgroundsByUuid: Map<string, GET<SandblastBackground>>;
  isAdding: boolean;
  onAdd: (item: RelatedCatalogItem, anchor: Point) => void;
}

interface Suggestion {
  background: GET<SandblastBackground>;
  anchor: Point;
}

// Suggests the selected background's related inlays, or those of every
// background on the stone when none is selected. A suggestion is dropped onto
// the background it came from.
export function RelatedInlays(props: RelatedInlaysProps) {
  const suggestions = createMemo<Suggestion[]>(() => {
    const selected = props.editor.selectedElements();
    const sources =
      selected.length === 1 && selected[0].type === "background"
        ? [selected[0]]
        : props.editor.draft.layout.elements.filter(
            (el): el is StoneBackgroundElement => el.type === "background",
          );

    const seen = new Set<string>();
    return sources.flatMap((el) => {
      const background = props.backgroundsByUuid.get(el.background_uuid);
      if (!background || background.related_items.length === 0 || seen.has(background.uuid)) {
        return [];
      }
      seen.add(background.uuid);
      return [{ background, anchor: { x: el.x, y: el.y } }];
    });
  });

  return (
    <Show when={suggestions().length > 0}>
      <section class="flex flex-col gap-3">
        <div>
          <h3 class="text-sm font-medium text-gray-900">Goes with your background</h3>
          <p class="mt-1 text-xs text-gray-500">
            Click one to add it to the project and place it on the background.
          </p>
        </div>
        <For each={suggestions()}>
          {(suggestion) => (
            <div class="flex flex-col gap-1.5">
              <Show when={suggestions().length > 1}>
                <p class="text-xs font-medium text-gray-600">{suggestion.background.name}</p>
              </Show>
              <div class="grid grid-cols-4 gap-2">
                <For each={suggestion.background.related_items}>
                  {(item) => (
                    <button
                      type="button"
                      title={`Add ${item.name} (${item.catalog_code})`}
                      disabled={props.isAdding}
                      class="flex aspect-square items-center justify-center rounded-md border border-gray-200 bg-white p-1.5 hover:border-blue-500 disabled:cursor-wait disabled:opacity-50"
                      onClick={() => props.onAdd(item, suggestion.anchor)}
                    >
                      <img
                        src={item.svg_url}
                        alt={item.name}
                        loading="lazy"
                        class="max-h-full max-w-full object-contain"
                      />
                    </button>
                  )}
                </For>
              </div>
            </div>
          )}
        </For>
      </section>
    </Show>
  );
}
