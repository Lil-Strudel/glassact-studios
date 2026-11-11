import { Show } from "solid-js";
import { Button } from "@glassact/ui";
import type { StoneEditor } from "./create-stone-editor";
import type { SaveStatus } from "./create-stone-autosave";

interface DesignerToolbarProps {
  editor: StoneEditor;
  saveStatus?: SaveStatus;
}

const SAVE_LABELS: Record<SaveStatus, string> = {
  saved: "All changes saved",
  pending: "Unsaved changes",
  saving: "Saving...",
  error: "Couldn't save — retrying on your next change",
  conflict: "Reloaded the latest version",
};

export function DesignerToolbar(props: DesignerToolbarProps) {
  const selectionCount = () => props.editor.selection().length;

  return (
    <div class="flex flex-wrap items-center gap-2">
      <Button size="sm" variant="outline" disabled={!props.editor.canUndo()} onClick={() => props.editor.undo()}>
        Undo
      </Button>
      <Button size="sm" variant="outline" disabled={!props.editor.canRedo()} onClick={() => props.editor.redo()}>
        Redo
      </Button>
      <div class="mx-1 h-5 w-px bg-gray-300" />
      <Show
        when={props.editor.selectedGroup()}
        fallback={
          <Button
            size="sm"
            variant="outline"
            disabled={selectionCount() < 2}
            onClick={() => props.editor.groupSelected()}
          >
            Group
          </Button>
        }
      >
        <Button size="sm" variant="outline" onClick={() => props.editor.ungroupSelected()}>
          Ungroup
        </Button>
      </Show>
      <Button
        size="sm"
        variant="outline"
        disabled={selectionCount() === 0}
        onClick={() => props.editor.removeSelected()}
      >
        Remove from stone
      </Button>
      <Show when={props.saveStatus}>
        {(status) => (
          <span
            class={
              status() === "error" ? "ml-auto text-xs text-red-600" : "ml-auto text-xs text-gray-500"
            }
          >
            {SAVE_LABELS[status()]}
          </span>
        )}
      </Show>
    </div>
  );
}
