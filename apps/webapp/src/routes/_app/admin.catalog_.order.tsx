import { createFileRoute, Link } from "@tanstack/solid-router";
import {
  For,
  Show,
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  onMount,
  untrack,
} from "solid-js";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  Button,
  TextField,
  TextFieldRoot,
  showToast,
  cn,
} from "@glassact/ui";
import {
  draggable,
  dropTargetForElements,
  monitorForElements,
} from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { combine } from "@atlaskit/pragmatic-drag-and-drop/combine";
import { CleanupFn } from "@atlaskit/pragmatic-drag-and-drop/dist/types/internal-types";
import {
  IoArrowDownOutline,
  IoArrowUpOutline,
  IoCloseOutline,
  IoReorderTwoOutline,
} from "solid-icons/io";
import { z } from "zod";
import type { CatalogItem, GET } from "@glassact/data";
import {
  getRankedCatalogOpts,
  putCatalogDisplayOrderOpts,
} from "../../queries/catalog";
import { browseCatalogOpts } from "../../queries/catalog-browse";
import { useDebounce } from "../../hooks/use-debounce";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/catalog_/order")({
  component: RouteComponent,
});

interface RankedRowProps {
  item: GET<CatalogItem>;
  index: number;
  total: number;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onRemove: () => void;
}

function RankedRow(props: RankedRowProps) {
  let ref!: HTMLLIElement;
  const [dragging, setDragging] = createSignal(false);
  const [draggedOver, setDraggedOver] = createSignal(false);

  let cleanup: CleanupFn;
  onMount(() => {
    if (!ref) return;
    const c1 = draggable({
      element: ref,
      getInitialData: () => ({ rowIndex: props.index }),
      onDragStart: () => setDragging(true),
      onDrop: () => setDragging(false),
    });
    const c2 = dropTargetForElements({
      element: ref,
      getData: () => ({ rowIndex: props.index }),
      onDragEnter: () => setDraggedOver(true),
      onDragLeave: () => setDraggedOver(false),
      onDrop: () => setDraggedOver(false),
    });
    cleanup = combine(c1, c2);
  });

  onCleanup(() => {
    if (cleanup) cleanup();
  });

  return (
    <li
      ref={ref}
      class={cn(
        "flex touch-none items-center gap-3 bg-white px-3 py-2",
        dragging() && "opacity-30",
        draggedOver() && !dragging() && "bg-primary/5 ring-1 ring-inset ring-primary",
      )}
    >
      <IoReorderTwoOutline
        size={20}
        class="flex-shrink-0 cursor-grab text-gray-400"
        aria-hidden="true"
      />

      <span class="w-8 flex-shrink-0 text-sm font-medium tabular-nums text-gray-500">
        #{props.index + 1}
      </span>

      <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center overflow-hidden rounded bg-gray-50">
        <img
          src={props.item.svg_url}
          alt=""
          draggable={false}
          class="max-h-full max-w-full object-contain"
        />
      </div>

      <div class="min-w-0 flex-1">
        <p class="truncate text-sm font-medium text-gray-900">
          {props.item.name}
        </p>
        <p class="truncate text-xs text-gray-500">
          {props.item.catalog_code} &middot; {props.item.category}
        </p>
      </div>

      <div class="flex flex-shrink-0 items-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          aria-label={`Move ${props.item.name} up`}
          disabled={props.index === 0}
          onClick={props.onMoveUp}
        >
          <IoArrowUpOutline size={16} />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          aria-label={`Move ${props.item.name} down`}
          disabled={props.index === props.total - 1}
          onClick={props.onMoveDown}
        >
          <IoArrowDownOutline size={16} />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          aria-label={`Remove ${props.item.name} from best sellers`}
          onClick={props.onRemove}
        >
          <IoCloseOutline size={18} />
        </Button>
      </div>
    </li>
  );
}

function RouteComponent() {
  const queryClient = useQueryClient();
  const rankedQuery = useQuery(() => getRankedCatalogOpts());
  const saveOrder = useMutation(() => putCatalogDisplayOrderOpts());

  // Working copy of the ranking. Nothing is written until "Save order", so a
  // half-finished drag session never reaches the database.
  const [ranked, setRanked] = createSignal<GET<CatalogItem>[]>([]);
  const [isDirty, setIsDirty] = createSignal(false);

  // Sync from the server only when fresh data arrives, and never over unsaved
  // edits. `isDirty` is untracked deliberately: clearing it after a save must
  // not re-run this and stomp the just-saved order with the pre-save cache.
  createEffect(() => {
    const data = rankedQuery.data;
    if (!data) return;
    if (untrack(isDirty)) return;
    setRanked(data);
  });

  const [search, setSearch] = createSignal("");
  const debouncedSearch = useDebounce(search, 300);

  const pickerQuery = useQuery(() =>
    browseCatalogOpts({ search: debouncedSearch(), limit: 20, offset: 0 }),
  );

  const rankedUUIDs = createMemo(() => new Set(ranked().map((i) => i.uuid)));

  const pickerResults = createMemo(() =>
    (pickerQuery.data?.items ?? []).filter(
      (item) => !rankedUUIDs().has(item.uuid),
    ),
  );

  function move(from: number, to: number) {
    const items = [...ranked()];
    if (to < 0 || to >= items.length || from === to) return;
    const [moved] = items.splice(from, 1);
    items.splice(to, 0, moved);
    setRanked(items);
    setIsDirty(true);
  }

  function addItem(item: GET<CatalogItem>) {
    setRanked([...ranked(), item]);
    setIsDirty(true);
  }

  function removeItem(uuid: string) {
    setRanked(ranked().filter((i) => i.uuid !== uuid));
    setIsDirty(true);
  }

  let cleanup: CleanupFn;
  onMount(() => {
    cleanup = monitorForElements({
      onDrop({ source, location }) {
        const dest = location.current.dropTargets[0];
        if (!dest) return;

        const { success, data } = z
          .object({ from: z.number().int(), to: z.number().int() })
          .safeParse({ from: source.data.rowIndex, to: dest.data.rowIndex });

        if (!success) return;
        move(data.from, data.to);
      },
    });
  });

  onCleanup(() => {
    if (cleanup) cleanup();
  });

  function handleSave() {
    saveOrder.mutate(
      { ordered_uuids: ranked().map((i) => i.uuid) },
      {
        onSuccess(items) {
          setIsDirty(false);
          setRanked(items);
          queryClient.invalidateQueries({ queryKey: ["catalog"] });
          queryClient.invalidateQueries({ queryKey: ["catalog-browse"] });
          showToast({
            title: "Best sellers updated",
            description:
              items.length === 0
                ? "No items are ranked; the catalog will sort alphabetically."
                : `${items.length} item${items.length === 1 ? "" : "s"} will show first in the catalog.`,
            variant: "success",
          });
        },
        onError(error) {
          showToast({
            title: "Failed to save order",
            description: isApiError(error)
              ? (error?.data?.error ?? "Unknown error")
              : "Unknown error",
            variant: "error",
          });
        },
      },
    );
  }

  function handleReset() {
    setRanked(rankedQuery.data ?? []);
    setIsDirty(false);
  }

  return (
    <div class="flex flex-col gap-6">
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">Best Sellers</h1>
          <p class="text-gray-600 mt-1">
            Items listed here show first when dealerships browse the catalog, in
            the order shown. Everything else follows alphabetically.
          </p>
        </div>
        <Button variant="outline" as={Link} to="/admin/catalog">
          Back to catalog
        </Button>
      </div>

      <div class="flex items-center justify-between rounded-md border bg-gray-50 px-4 py-3">
        <p class="text-sm text-gray-600">
          <Show when={isDirty()} fallback="No unsaved changes.">
            You have unsaved changes.
          </Show>
        </p>
        <div class="flex gap-2">
          <Button
            variant="outline"
            onClick={handleReset}
            disabled={!isDirty() || saveOrder.isPending}
          >
            Discard changes
          </Button>
          <Button
            onClick={handleSave}
            disabled={!isDirty() || saveOrder.isPending}
          >
            {saveOrder.isPending ? "Saving..." : "Save order"}
          </Button>
        </div>
      </div>

      <div>
        <h2 class="mb-2 text-sm font-medium text-gray-900">
          Ranked items ({ranked().length})
        </h2>

        <Show
          when={ranked().length > 0}
          fallback={
            <div class="rounded-md border border-dashed p-8 text-center text-sm text-gray-500">
              Nothing is ranked yet. Search below to add your best sellers.
            </div>
          }
        >
          <ul class="flex flex-col rounded-md border divide-y">
            <For each={ranked()}>
              {(item, index) => (
                <RankedRow
                  item={item}
                  index={index()}
                  total={ranked().length}
                  onMoveUp={() => move(index(), index() - 1)}
                  onMoveDown={() => move(index(), index() + 1)}
                  onRemove={() => removeItem(item.uuid)}
                />
              )}
            </For>
          </ul>
        </Show>
      </div>

      <div class="border-t pt-6">
        <h2 class="mb-2 text-sm font-medium text-gray-900">
          Add an item to the ranking
        </h2>

        <TextFieldRoot value={search()} onChange={setSearch}>
          <TextField
            placeholder="Search by name or code..."
            class="max-w-sm"
          />
        </TextFieldRoot>

        <Show when={debouncedSearch().length > 0}>
          <Show
            when={pickerResults().length > 0}
            fallback={
              <p class="mt-4 text-sm text-gray-500">
                No unranked items match that search.
              </p>
            }
          >
            <ul class="mt-4 flex flex-col rounded-md border divide-y">
              <For each={pickerResults()}>
                {(item) => (
                  <li class="flex items-center gap-3 px-3 py-2">
                    <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded bg-gray-50">
                      <img
                        src={item.svg_url}
                        alt=""
                        class="max-h-full max-w-full object-contain"
                      />
                    </div>
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-sm font-medium text-gray-900">
                        {item.name}
                      </p>
                      <p class="truncate text-xs text-gray-500">
                        {item.catalog_code} &middot; {item.category}
                      </p>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => addItem(item)}
                    >
                      Add
                    </Button>
                  </li>
                )}
              </For>
            </ul>
          </Show>
        </Show>
      </div>
    </div>
  );
}
