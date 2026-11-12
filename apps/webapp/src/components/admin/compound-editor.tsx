import { Show, createSignal } from "solid-js";
import { createStore } from "solid-js/store";
import type {
  CatalogCompound,
  CatalogCompoundWrite,
  CatalogItem,
  GET,
} from "@glassact/data";
import { Button } from "@glassact/ui";
import { CatalogItemPicker } from "../catalog/catalog-item-picker";
import { createStoneEditor } from "../stone-designer/create-stone-editor";
import { StoneCanvas, type PlacedInlayArt } from "../stone-designer/stone-canvas";
import { boundsOf, boxCorners, newElementId } from "../stone-designer/stone-geometry";
import { useDesignerShortcuts } from "../stone-designer/use-designer-shortcuts";

interface CompoundEditorProps {
  initial?: GET<CatalogCompound>;
  isSaving: boolean;
  onSave: (body: CatalogCompoundWrite) => void;
}

// Members are arranged on a fixed working area; on save their positions are
// re-expressed relative to the arrangement's own center.
const ARTBOARD = { width: 30, height: 20 };

interface MemberItem {
  id: number;
  name: string;
  svgUrl: string;
  width: number;
  height: number;
}

export function CompoundEditor(props: CompoundEditorProps) {
  const [name, setName] = createSignal(props.initial?.name ?? "");
  const [description, setDescription] = createSignal(props.initial?.description ?? "");
  const [isActive, setIsActive] = createSignal(props.initial?.is_active ?? true);

  // Keyed by catalog item uuid, which doubles as each element's inlay_uuid.
  const [items, setItems] = createStore<Record<string, MemberItem>>(
    Object.fromEntries(
      (props.initial?.members ?? []).map((m) => [
        m.catalog_item.uuid,
        {
          id: m.catalog_item_id,
          name: m.catalog_item.name,
          svgUrl: m.catalog_item.svg_url,
          width: m.catalog_item.default_width,
          height: m.catalog_item.default_height,
        },
      ]),
    ),
  );

  const editor = createStoneEditor({
    shape: "flat-top",
    ...ARTBOARD,
    graniteKey: "plain",
    customHex: "#000000",
    layout: {
      groups: [],
      elements: (props.initial?.members ?? []).map((m) => ({
        id: newElementId(),
        type: "inlay" as const,
        inlay_uuid: m.catalog_item.uuid,
        x: ARTBOARD.width / 2 + m.center_x,
        y: ARTBOARD.height / 2 + m.center_y,
        rotation: m.rotation_deg,
        group_id: null,
      })),
    },
  });
  useDesignerShortcuts(editor, () => true);

  const resolveInlay = (uuid: string): PlacedInlayArt | undefined => {
    const item = items[uuid];
    return item
      ? { name: item.name, imageUrl: item.svgUrl, width: item.width, height: item.height }
      : undefined;
  };

  function addItem(item: GET<CatalogItem>) {
    setItems(item.uuid, {
      id: item.id,
      name: item.name,
      svgUrl: item.svg_url,
      width: item.default_width,
      height: item.default_height,
    });
    editor.addElements([
      {
        id: newElementId(),
        type: "inlay",
        inlay_uuid: item.uuid,
        x: ARTBOARD.width / 2,
        y: ARTBOARD.height / 2,
        rotation: 0,
        group_id: null,
      },
    ]);
  }

  const memberCount = () => editor.draft.layout.elements.length;
  const canSave = () => name().trim() !== "" && memberCount() >= 2 && !props.isSaving;

  function save() {
    const elements = editor.draft.layout.elements.flatMap((el) =>
      el.type === "inlay" && items[el.inlay_uuid] ? [{ el, item: items[el.inlay_uuid] }] : [],
    );
    const corners = elements.flatMap(({ el, item }) =>
      boxCorners({ x: el.x, y: el.y }, el.rotation, {
        x: -item.width / 2,
        y: -item.height / 2,
        width: item.width,
        height: item.height,
      }),
    );
    const center = boundsOf(corners).center;

    props.onSave({
      name: name().trim(),
      description: description().trim() || null,
      is_active: isActive(),
      display_order: props.initial?.display_order ?? null,
      members: elements.map(({ el, item }) => ({
        catalog_item_id: item.id,
        center_x: Number((el.x - center.x).toFixed(3)),
        center_y: Number((el.y - center.y).toFixed(3)),
        rotation_deg: el.rotation,
      })),
    });
  }

  return (
    <div class="flex flex-col gap-6 lg:flex-row">
      <div class="flex min-w-0 flex-1 flex-col gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <Button size="sm" variant="outline" disabled={!editor.canUndo()} onClick={() => editor.undo()}>
            Undo
          </Button>
          <Button size="sm" variant="outline" disabled={!editor.canRedo()} onClick={() => editor.redo()}>
            Redo
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={editor.selection().length === 0}
            onClick={() => editor.removeSelected()}
          >
            Remove
          </Button>
          <span class="ml-auto text-xs text-gray-500">
            {memberCount()} inlays · 1 square = 1 inch, shown at real size
          </span>
        </div>
        <StoneCanvas
          stone={null}
          artboard={ARTBOARD}
          elements={editor.draft.layout.elements}
          resolveInlay={resolveInlay}
          resolveBackground={() => undefined}
          editor={editor}
          isZoomable
          class="h-[480px] rounded-lg border border-gray-200 bg-white"
        />
      </div>

      <aside class="flex w-full flex-col gap-5 lg:w-96">
        <label class="flex flex-col gap-1 text-sm font-medium text-gray-900">
          Name
          <input
            value={name()}
            onInput={(e) => setName(e.currentTarget.value)}
            class="rounded-md border border-gray-300 px-3 py-2 text-sm font-normal"
          />
        </label>
        <label class="flex flex-col gap-1 text-sm font-medium text-gray-900">
          Description
          <textarea
            rows={2}
            value={description()}
            onInput={(e) => setDescription(e.currentTarget.value)}
            class="rounded-md border border-gray-300 px-3 py-2 text-sm font-normal"
          />
        </label>
        <label class="flex items-center gap-2 text-sm text-gray-900">
          <input
            type="checkbox"
            checked={isActive()}
            onChange={(e) => setIsActive(e.currentTarget.checked)}
            class="h-4 w-4 rounded border-gray-300"
          />
          Offered to dealerships
        </label>
        <CatalogItemPicker title="Add catalog inlays" onPick={addItem} />
        <Button disabled={!canSave()} onClick={save}>
          {props.isSaving ? "Saving..." : "Save compound"}
        </Button>
        <Show when={memberCount() < 2}>
          <p class="text-xs text-gray-500">A compound needs at least two inlays.</p>
        </Show>
      </aside>
    </div>
  );
}
