import { createSignal } from "solid-js";
import { useQueryClient } from "@tanstack/solid-query";
import { showToast } from "@glassact/ui";
import { postCatalogInlay } from "../../queries/inlay";
import { isApiError } from "../../utils/is-api-error";
import type { StoneEditor } from "./create-stone-editor";
import { newElementId, type Point } from "./stone-geometry";

export interface AddableCatalogItem {
  id: number;
  name: string;
}

// Adds a stock catalog inlay to the project and places it on the stone. The
// placement rides on the designer's autosave like any other edit, so a lost
// save only leaves the inlay waiting in "Not on the stone yet".
export function useAddCatalogInlay(options: {
  editor: StoneEditor;
  projectUuid: () => string;
}) {
  const queryClient = useQueryClient();
  const [isAdding, setIsAdding] = createSignal(false);
  let addedCount = 0;

  function defaultAnchor(): Point {
    const selected = options.editor.selectedElements();
    if (selected.length === 1 && selected[0].type === "background") {
      return { x: selected[0].x, y: selected[0].y };
    }
    const { width, height } = options.editor.draft;
    return { x: width / 2, y: height / 2 };
  }

  async function add(item: AddableCatalogItem, anchor: Point = defaultAnchor()) {
    if (isAdding()) return;
    setIsAdding(true);
    try {
      const projectUuid = options.projectUuid();
      const inlay = await postCatalogInlay({
        projectUuid,
        body: { name: item.name, catalog_item_id: item.id },
      });
      // The canvas sizes an inlay from the project's inlay list, so it has to
      // hold the new inlay before the element is placed.
      await queryClient.invalidateQueries({ queryKey: ["project", projectUuid, "inlays"] });

      const offset = (addedCount++ % 5) * 1.5;
      options.editor.addElements([
        {
          id: newElementId(),
          type: "inlay",
          inlay_uuid: inlay.uuid,
          x: anchor.x + offset,
          y: anchor.y + offset,
          rotation: 0,
          group_id: null,
        },
      ]);
    } catch (error) {
      showToast({
        title: `Failed to add ${item.name}`,
        description:
          error instanceof Error && isApiError(error)
            ? (error.data?.error ?? "Unknown error")
            : "Unknown error",
        variant: "error",
      });
    } finally {
      setIsAdding(false);
    }
  }

  return { add, isAdding };
}
