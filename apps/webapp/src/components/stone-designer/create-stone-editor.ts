import { createMemo, createSignal, type Accessor } from "solid-js";
import { createStore, produce, reconcile, unwrap } from "solid-js/store";
import type {
  StoneElement,
  StoneGroup,
  StoneLayout,
  StoneShape,
} from "@glassact/data";
import { cloneJson, newElementId } from "./stone-geometry";

export interface StoneDraft {
  shape: StoneShape;
  width: number;
  height: number;
  graniteKey: string;
  customHex: string;
  layout: StoneLayout;
}

const HISTORY_LIMIT = 100;

export interface StoneEditor {
  draft: StoneDraft;
  selection: Accessor<string[]>;
  selectedElements: Accessor<StoneElement[]>;
  // The group every selected element belongs to, when they are exactly one
  // whole group.
  selectedGroup: Accessor<StoneGroup | null>;
  // Bumped on every change, so autosave can watch one signal.
  revision: Accessor<number>;
  canUndo: Accessor<boolean>;
  canRedo: Accessor<boolean>;

  load: (draft: StoneDraft) => void;
  update: (mutate: (draft: StoneDraft) => void) => void;
  // A drag records one undo step at its start, then mutates freely.
  beginGesture: () => void;
  gesture: (mutate: (draft: StoneDraft) => void) => void;
  undo: () => void;
  redo: () => void;

  select: (id: string, additive: boolean) => void;
  selectIds: (ids: string[]) => void;
  clearSelection: () => void;

  addElements: (elements: StoneElement[], group?: StoneGroup) => void;
  removeSelected: () => void;
  groupSelected: () => void;
  ungroupSelected: () => void;
  nudgeSelected: (dx: number, dy: number) => void;
  moveSelectedInStack: (direction: "forward" | "backward") => void;
}

function snapshot(draft: StoneDraft): StoneDraft {
  return cloneJson(unwrap(draft));
}

export function createStoneEditor(initial: StoneDraft): StoneEditor {
  const [draft, setDraft] = createStore<StoneDraft>(cloneJson(initial));
  const [selection, setSelection] = createSignal<string[]>([]);
  const [revision, setRevision] = createSignal(0);
  const [past, setPast] = createSignal<StoneDraft[]>([]);
  const [future, setFuture] = createSignal<StoneDraft[]>([]);

  const bump = () => setRevision((r) => r + 1);

  function pushHistory() {
    setPast((p) => [...p.slice(-(HISTORY_LIMIT - 1)), snapshot(draft)]);
    setFuture([]);
  }

  function pruneSelection() {
    const ids = new Set(draft.layout.elements.map((el) => el.id));
    setSelection((sel) => sel.filter((id) => ids.has(id)));
  }

  function pruneEmptyGroups(d: StoneDraft) {
    const used = new Set(
      d.layout.elements.map((el) => el.group_id).filter(Boolean),
    );
    d.layout.groups = d.layout.groups.filter((g) => used.has(g.id));
  }

  function update(mutate: (d: StoneDraft) => void) {
    pushHistory();
    setDraft(produce(mutate));
    pruneSelection();
    bump();
  }

  function gesture(mutate: (d: StoneDraft) => void) {
    setDraft(produce(mutate));
    bump();
  }

  function restore(target: StoneDraft) {
    setDraft(reconcile(target));
    pruneSelection();
    bump();
  }

  const selectedElements = createMemo(() => {
    const ids = new Set(selection());
    return draft.layout.elements.filter((el) => ids.has(el.id));
  });

  const selectedGroup = createMemo(() => {
    const selected = selectedElements();
    const groupId = selected[0]?.group_id;
    if (!groupId || selected.some((el) => el.group_id !== groupId)) return null;
    const members = draft.layout.elements.filter(
      (el) => el.group_id === groupId,
    );
    if (members.length !== selected.length) return null;
    return draft.layout.groups.find((g) => g.id === groupId) ?? null;
  });

  function idsWithGroups(ids: string[]): string[] {
    const groupIds = new Set(
      draft.layout.elements
        .filter((el) => ids.includes(el.id) && el.group_id)
        .map((el) => el.group_id),
    );
    return draft.layout.elements
      .filter((el) => ids.includes(el.id) || groupIds.has(el.group_id))
      .map((el) => el.id);
  }

  return {
    draft,
    selection,
    selectedElements,
    selectedGroup,
    revision,
    canUndo: () => past().length > 0,
    canRedo: () => future().length > 0,

    load(next) {
      setDraft(reconcile(cloneJson(next)));
      setPast([]);
      setFuture([]);
      pruneSelection();
    },
    update,
    beginGesture: pushHistory,
    gesture,
    undo() {
      const previous = past().at(-1);
      if (!previous) return;
      setFuture((f) => [...f, snapshot(draft)]);
      setPast((p) => p.slice(0, -1));
      restore(previous);
    },
    redo() {
      const next = future().at(-1);
      if (!next) return;
      setPast((p) => [...p, snapshot(draft)]);
      setFuture((f) => f.slice(0, -1));
      restore(next);
    },

    select(id, additive) {
      const ids = idsWithGroups([id]);
      if (!additive) {
        setSelection(ids);
        return;
      }
      const current = new Set(selection());
      const isSelected = ids.every((i) => current.has(i));
      ids.forEach((i) => (isSelected ? current.delete(i) : current.add(i)));
      setSelection([...current]);
    },
    selectIds: (ids) => setSelection(idsWithGroups(ids)),
    clearSelection: () => setSelection([]),

    addElements(elements, group) {
      update((d) => {
        if (group) d.layout.groups.push(group);
        d.layout.elements.push(...elements);
      });
      setSelection(elements.map((el) => el.id));
    },
    removeSelected() {
      const ids = new Set(selection());
      if (ids.size === 0) return;
      update((d) => {
        d.layout.elements = d.layout.elements.filter((el) => !ids.has(el.id));
        pruneEmptyGroups(d);
      });
      setSelection([]);
    },
    groupSelected() {
      const ids = new Set(selection());
      if (ids.size < 2) return;
      const group: StoneGroup = {
        id: newElementId(),
        name: "Group",
        compound_uuid: null,
      };
      update((d) => {
        d.layout.groups.push(group);
        d.layout.elements.forEach((el) => {
          if (ids.has(el.id)) el.group_id = group.id;
        });
        pruneEmptyGroups(d);
      });
    },
    ungroupSelected() {
      const ids = new Set(selection());
      update((d) => {
        d.layout.elements.forEach((el) => {
          if (ids.has(el.id)) el.group_id = null;
        });
        pruneEmptyGroups(d);
      });
    },
    nudgeSelected(dx, dy) {
      const ids = new Set(selection());
      if (ids.size === 0) return;
      update((d) => {
        d.layout.elements.forEach((el) => {
          if (!ids.has(el.id)) return;
          el.x += dx;
          el.y += dy;
        });
      });
    },
    moveSelectedInStack(direction) {
      const ids = new Set(selection());
      if (ids.size === 0) return;
      update((d) => {
        const moving = d.layout.elements.filter((el) => ids.has(el.id));
        const rest = d.layout.elements.filter((el) => !ids.has(el.id));
        d.layout.elements =
          direction === "forward" ? [...rest, ...moving] : [...moving, ...rest];
      });
    },
  };
}
