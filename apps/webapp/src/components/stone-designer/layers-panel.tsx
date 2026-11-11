import { For, Show, createMemo } from "solid-js";
import { Button, cn } from "@glassact/ui";
import type { StoneEditor } from "./create-stone-editor";
import { describeElement } from "./stone-draft";

interface LayersPanelProps {
  editor: StoneEditor;
  inlayName: (uuid: string) => string | undefined;
  backgroundName: (uuid: string) => string | undefined;
}

export function LayersPanel(props: LayersPanelProps) {
  const topFirst = createMemo(() => [...props.editor.draft.layout.elements].reverse());
  const selected = createMemo(() => new Set(props.editor.selection()));
  const groupNames = createMemo(
    () => new Map(props.editor.draft.layout.groups.map((g) => [g.id, g.name])),
  );

  return (
    <div class="flex flex-col gap-3">
      <div class="flex gap-2">
        <Button
          size="sm"
          variant="outline"
          disabled={selected().size === 0}
          onClick={() => props.editor.moveSelectedInStack("forward")}
        >
          Bring to front
        </Button>
        <Button
          size="sm"
          variant="outline"
          disabled={selected().size === 0}
          onClick={() => props.editor.moveSelectedInStack("backward")}
        >
          Send to back
        </Button>
      </div>
      <Show
        when={topFirst().length > 0}
        fallback={<p class="text-xs text-gray-500">Nothing on the stone yet.</p>}
      >
        <ul class="flex flex-col gap-1">
          <For each={topFirst()}>
            {(el) => (
              <li>
                <button
                  type="button"
                  class={cn(
                    "flex w-full items-center justify-between gap-2 rounded-md px-2 py-1.5 text-left text-sm",
                    selected().has(el.id) ? "bg-blue-50 text-blue-800" : "text-gray-800 hover:bg-gray-50",
                  )}
                  onClick={(e) => props.editor.select(el.id, e.shiftKey)}
                >
                  <span class="truncate">
                    {describeElement(el, props.inlayName, props.backgroundName)}
                  </span>
                  <Show when={el.group_id && groupNames().get(el.group_id)}>
                    {(name) => (
                      <span class="flex-shrink-0 rounded bg-gray-100 px-1.5 py-0.5 text-[10px] text-gray-600">
                        {name()}
                      </span>
                    )}
                  </Show>
                </button>
              </li>
            )}
          </For>
        </ul>
      </Show>
    </div>
  );
}
