import { For, Show, createMemo } from "solid-js";
import {
  MAX_STONE_TEXT_LENGTH,
  STONE_TEXT_FINISHES,
  STONE_TEXT_FONTS,
  type StoneTextAlign,
  type StoneTextElement,
  type StoneTextFinish,
} from "@glassact/data";
import { Button, cn } from "@glassact/ui";
import type { StoneDraft, StoneEditor } from "./create-stone-editor";
import { STONE_FONTS } from "./stone-fonts";
import { newTextElement } from "./stone-text";
import { NameAndDatesForm } from "./name-and-dates-form";

interface TextPanelProps {
  editor: StoneEditor;
}

const ALIGNS: StoneTextAlign[] = ["left", "center", "right"];
const FINISH_LABELS: Record<StoneTextFinish, string> = {
  engraved: "Engraved",
  frosted: "Frosted",
  "black-fill": "Black fill",
};

const inputClass = "w-full rounded-md border border-gray-300 px-2 py-1 text-sm";

export function TextPanel(props: TextPanelProps) {
  const selectedTexts = createMemo(() =>
    props.editor
      .selectedElements()
      .filter((el): el is StoneTextElement => el.type === "text"),
  );

  // The value every selected line agrees on, or undefined when they differ.
  function sharedValue<K extends keyof StoneTextElement>(key: K) {
    const [first, ...rest] = selectedTexts();
    if (!first) return undefined;
    return rest.every((el) => el[key] === first[key]) ? first[key] : undefined;
  }

  function addLine() {
    const draft = props.editor.draft;
    props.editor.addElements([
      newTextElement({
        text: "In Loving Memory",
        x: draft.width / 2,
        y: draft.height * 0.3,
        size: Math.max(0.5, draft.width * 0.03),
        font: "serif",
      }),
    ]);
  }

  function updateSelected(mutate: (el: StoneTextElement) => void, isGesture = false) {
    const ids = new Set(selectedTexts().map((el) => el.id));
    if (ids.size === 0) return;
    const apply = (d: StoneDraft) => {
      for (const el of d.layout.elements) {
        if (el.type === "text" && ids.has(el.id)) mutate(el);
      }
    };
    if (isGesture) props.editor.gesture(apply);
    else props.editor.update(apply);
  }

  return (
    <div class="flex flex-col gap-6">
      <Show when={selectedTexts().length > 0}>
        <section class="flex flex-col gap-3">
          <h3 class="text-sm font-medium text-gray-900">
            {selectedTexts().length === 1
              ? "Selected text"
              : `${selectedTexts().length} lines of text selected`}
          </h3>
          <Show when={selectedTexts().length === 1 ? selectedTexts()[0] : null}>
            {(el) => (
              <textarea
                rows={3}
                maxLength={MAX_STONE_TEXT_LENGTH}
                value={el().text}
                onFocus={() => props.editor.beginGesture()}
                onInput={(e) => {
                  const value = e.currentTarget.value;
                  updateSelected((t) => (t.text = value), true);
                }}
                class={inputClass}
              />
            )}
          </Show>
          <label class="flex flex-col gap-1 text-xs text-gray-600">
            Font
            <select
              value={sharedValue("font") ?? ""}
              onChange={(e) => {
                const font = e.currentTarget.value as StoneTextElement["font"];
                updateSelected((t) => (t.font = font));
              }}
              class={inputClass}
            >
              <Show when={sharedValue("font") === undefined}>
                <option value="" disabled>
                  Mixed
                </option>
              </Show>
              <For each={STONE_TEXT_FONTS}>
                {(font) => <option value={font}>{STONE_FONTS[font].label}</option>}
              </For>
            </select>
          </label>
          <label class="flex items-center justify-between gap-2 text-xs text-gray-600">
            Letter height (inches)
            <input
              type="number"
              min="0.1"
              max="24"
              step="0.125"
              placeholder="Mixed"
              value={sharedValue("size") ?? ""}
              onChange={(e) => {
                const value = Number(e.currentTarget.value);
                if (value >= 0.1 && value <= 24) updateSelected((t) => (t.size = value));
              }}
              class="w-24 rounded-md border border-gray-300 px-2 py-1 text-sm"
            />
          </label>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="flex overflow-hidden rounded-md border border-gray-300 text-xs">
              <For each={ALIGNS}>
                {(align) => (
                  <button
                    type="button"
                    class={cn("px-2.5 py-1 capitalize", sharedValue("align") === align ? "bg-blue-50 text-blue-700" : "hover:bg-gray-50")}
                    onClick={() => updateSelected((t) => (t.align = align))}
                  >
                    {align}
                  </button>
                )}
              </For>
            </div>
            <div class="flex overflow-hidden rounded-md border border-gray-300 text-xs">
              <For each={STONE_TEXT_FINISHES}>
                {(finish) => (
                  <button
                    type="button"
                    class={cn("px-2.5 py-1", sharedValue("finish") === finish ? "bg-blue-50 text-blue-700" : "hover:bg-gray-50")}
                    onClick={() => updateSelected((t) => (t.finish = finish))}
                  >
                    {FINISH_LABELS[finish]}
                  </button>
                )}
              </For>
            </div>
          </div>
        </section>
      </Show>

      <section>
        <Button variant="outline" size="sm" onClick={addLine}>
          Add a line of text
        </Button>
      </section>

      <NameAndDatesForm editor={props.editor} />
    </div>
  );
}
