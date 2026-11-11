import { onCleanup, onMount, type Accessor } from "solid-js";
import type { StoneEditor } from "./create-stone-editor";

const NUDGE_INCHES = 0.125;
const LARGE_NUDGE_INCHES = 1;

function isTyping(target: EventTarget | null) {
  return (
    target instanceof HTMLElement &&
    (target.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName))
  );
}

export function useDesignerShortcuts(editor: StoneEditor, isEnabled: Accessor<boolean>) {
  function onKeyDown(e: KeyboardEvent) {
    if (!isEnabled() || isTyping(e.target)) return;
    const mod = e.metaKey || e.ctrlKey;
    const step = e.shiftKey ? LARGE_NUDGE_INCHES : NUDGE_INCHES;

    if (mod && e.key.toLowerCase() === "z") {
      e.preventDefault();
      if (e.shiftKey) editor.redo();
      else editor.undo();
      return;
    }
    if (mod && e.key.toLowerCase() === "y") {
      e.preventDefault();
      editor.redo();
      return;
    }

    const nudges: Record<string, [number, number]> = {
      ArrowLeft: [-step, 0],
      ArrowRight: [step, 0],
      ArrowUp: [0, -step],
      ArrowDown: [0, step],
    };
    if (e.key in nudges && editor.selection().length > 0) {
      e.preventDefault();
      editor.nudgeSelected(...nudges[e.key]);
      return;
    }
    if ((e.key === "Delete" || e.key === "Backspace") && editor.selection().length > 0) {
      e.preventDefault();
      editor.removeSelected();
      return;
    }
    if (e.key === "Escape") editor.clearSelection();
  }

  onMount(() => window.addEventListener("keydown", onKeyDown));
  onCleanup(() => window.removeEventListener("keydown", onKeyDown));
}
