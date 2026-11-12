import { createSignal } from "solid-js";
import type { StoneTextElement } from "@glassact/data";
import { Button } from "@glassact/ui";
import type { StoneEditor } from "./create-stone-editor";
import { newTextElement } from "./stone-text";

const inputClass = "w-full rounded-md border border-gray-300 px-2 py-1 text-sm";

export function NameAndDatesForm(props: { editor: StoneEditor }) {
  const [name, setName] = createSignal("");
  const [born, setBorn] = createSignal("");
  const [died, setDied] = createSignal("");
  const [epitaph, setEpitaph] = createSignal("");

  const canInsert = () => name().trim() !== "" || born().trim() !== "" || died().trim() !== "";

  // Sizes follow the stone so the lines read the same on a marker as on a
  // large upright.
  function insert() {
    const draft = props.editor.draft;
    const x = draft.width / 2;
    const nameSize = Math.min(3, Math.max(0.6, draft.width * 0.05));
    const smallSize = nameSize * 0.55;

    const lines: StoneTextElement[] = [];
    let y = draft.height * 0.62;
    if (name().trim()) {
      lines.push(newTextElement({ text: name().trim().toUpperCase(), x, y, size: nameSize, font: "roman" }));
      y += nameSize * 1.9;
    }
    const dates = [born().trim(), died().trim()].filter(Boolean).join(" – ");
    if (dates) {
      lines.push(newTextElement({ text: dates, x, y, size: smallSize, font: "serif" }));
      y += smallSize * 2.6;
    }
    if (epitaph().trim()) {
      lines.push(newTextElement({ text: epitaph().trim(), x, y, size: smallSize, font: "script" }));
    }

    props.editor.addElements(lines);
    setName("");
    setBorn("");
    setDied("");
    setEpitaph("");
  }

  return (
    <section class="flex flex-col gap-2 border-t border-gray-200 pt-4">
      <h3 class="text-sm font-medium text-gray-900">Name and dates</h3>
      <p class="text-xs text-gray-500">
        Lays out the usual name, dates and epitaph lines as ordinary text you
        can move and restyle like any other.
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
