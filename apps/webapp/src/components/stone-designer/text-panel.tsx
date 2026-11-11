import { For, Show, createMemo, createSignal } from "solid-js";
import {
  MAX_STONE_TEXT_LENGTH,
  STONE_TEXT_FONTS,
  type StoneTextAlign,
  type StoneTextElement,
  type StoneTextFinish,
} from "@glassact/data";
import { Button, cn } from "@glassact/ui";
import type { StoneEditor } from "./create-stone-editor";
import { STONE_FONTS } from "./stone-fonts";
import { newElementId } from "./stone-geometry";

interface TextPanelProps {
  editor: StoneEditor;
}

const ALIGNS: StoneTextAlign[] = ["left", "center", "right"];
const FINISH_LABELS: Record<StoneTextFinish, string> = {
  engraved: "Engraved",
  frosted: "Frosted",
};

const inputClass = "w-full rounded-md border border-gray-300 px-2 py-1 text-sm";

export function TextPanel(props: TextPanelProps) {
  const selectedText = createMemo(() => {
    const selected = props.editor.selectedElements();
    return selected.length === 1 && selected[0].type === "text" ? selected[0] : null;
  });

  function textElement(
    text: string,
    y: number,
    size: number,
    font: StoneTextElement["font"],
    groupId: string | null = null,
  ): StoneTextElement {
    return {
      id: newElementId(),
      type: "text",
      text,
      font,
      size,
      align: "center",
      finish: "engraved",
      x: props.editor.draft.width / 2,
      y,
      rotation: 0,
      group_id: groupId,
    };
  }

  function addLine() {
    const draft = props.editor.draft;
    props.editor.addElements([
      textElement("In Loving Memory", draft.height * 0.3, Math.max(0.5, draft.width * 0.03), "serif"),
    ]);
  }

  function updateSelected(mutate: (el: StoneTextElement) => void, isGesture = false) {
    const target = selectedText();
    if (!target) return;
    const apply = (d: typeof props.editor.draft) => {
      const el = d.layout.elements.find((e) => e.id === target.id);
      if (el?.type === "text") mutate(el);
    };
    if (isGesture) props.editor.gesture(apply);
    else props.editor.update(apply);
  }

  return (
    <div class="flex flex-col gap-6">
      <Show when={selectedText()}>
        {(el) => (
          <section class="flex flex-col gap-3">
            <h3 class="text-sm font-medium text-gray-900">Selected text</h3>
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
            <label class="flex flex-col gap-1 text-xs text-gray-600">
              Font
              <select
                value={el().font}
                onChange={(e) => {
                  const font = e.currentTarget.value as StoneTextElement["font"];
                  updateSelected((t) => (t.font = font));
                }}
                class={inputClass}
              >
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
                value={el().size}
                onChange={(e) => {
                  const value = Number(e.currentTarget.value);
                  if (value >= 0.1 && value <= 24) updateSelected((t) => (t.size = value));
                }}
                class="w-24 rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
            <div class="flex items-center justify-between gap-2">
              <div class="flex overflow-hidden rounded-md border border-gray-300 text-xs">
                <For each={ALIGNS}>
                  {(align) => (
                    <button
                      type="button"
                      class={cn("px-2.5 py-1 capitalize", el().align === align ? "bg-blue-50 text-blue-700" : "hover:bg-gray-50")}
                      onClick={() => updateSelected((t) => (t.align = align))}
                    >
                      {align}
                    </button>
                  )}
                </For>
              </div>
              <div class="flex overflow-hidden rounded-md border border-gray-300 text-xs">
                <For each={Object.keys(FINISH_LABELS) as StoneTextFinish[]}>
                  {(finish) => (
                    <button
                      type="button"
                      class={cn("px-2.5 py-1", el().finish === finish ? "bg-blue-50 text-blue-700" : "hover:bg-gray-50")}
                      onClick={() => updateSelected((t) => (t.finish = finish))}
                    >
                      {FINISH_LABELS[finish]}
                    </button>
                  )}
                </For>
              </div>
            </div>
          </section>
        )}
      </Show>

      <section>
        <Button variant="outline" size="sm" onClick={addLine}>
          Add a line of text
        </Button>
      </section>

      <MemorialTemplateForm editor={props.editor} makeText={textElement} />
    </div>
  );
}

function MemorialTemplateForm(props: {
  editor: StoneEditor;
  makeText: (
    text: string,
    y: number,
    size: number,
    font: StoneTextElement["font"],
    groupId: string | null,
  ) => StoneTextElement;
}) {
  const [name, setName] = createSignal("");
  const [born, setBorn] = createSignal("");
  const [died, setDied] = createSignal("");
  const [epitaph, setEpitaph] = createSignal("");

  const canInsert = () => name().trim() !== "" || born().trim() !== "" || died().trim() !== "";

  // Sizes follow the stone so the block reads the same on a marker as on a
  // large upright.
  function insert() {
    const draft = props.editor.draft;
    const nameSize = Math.min(3, Math.max(0.6, draft.width * 0.05));
    const datesSize = nameSize * 0.55;
    const epitaphSize = nameSize * 0.55;
    const groupId = newElementId();

    const elements: StoneTextElement[] = [];
    let y = draft.height * 0.62;
    if (name().trim()) {
      elements.push(props.makeText(name().trim().toUpperCase(), y, nameSize, "roman", groupId));
      y += nameSize * 1.9;
    }
    const dates = [born().trim(), died().trim()].filter(Boolean).join(" – ");
    if (dates) {
      elements.push(props.makeText(dates, y, datesSize, "serif", groupId));
      y += datesSize * 2.6;
    }
    if (epitaph().trim()) {
      elements.push(props.makeText(epitaph().trim(), y, epitaphSize, "script", groupId));
    }

    props.editor.addElements(elements, {
      id: groupId,
      name: "Memorial text",
      compound_uuid: null,
    });
    setName("");
    setBorn("");
    setDied("");
    setEpitaph("");
  }

  return (
    <section class="flex flex-col gap-2 border-t border-gray-200 pt-4">
      <h3 class="text-sm font-medium text-gray-900">Name and dates</h3>
      <p class="text-xs text-gray-500">
        Fills in a laid-out block of lettering you can then move or edit.
      </p>
      <input placeholder="Full name" value={name()} onInput={(e) => setName(e.currentTarget.value)} class={inputClass} />
      <div class="flex gap-2">
        <input placeholder="Born" value={born()} onInput={(e) => setBorn(e.currentTarget.value)} class={inputClass} />
        <input placeholder="Died" value={died()} onInput={(e) => setDied(e.currentTarget.value)} class={inputClass} />
      </div>
      <input
        placeholder="Epitaph (optional)"
        value={epitaph()}
        onInput={(e) => setEpitaph(e.currentTarget.value)}
        class={inputClass}
      />
      <Button size="sm" disabled={!canInsert()} onClick={insert} class="self-start">
        Add to stone
      </Button>
    </section>
  );
}
