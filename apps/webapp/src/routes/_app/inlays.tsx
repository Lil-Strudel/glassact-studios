import {
  cn,
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  showToast,
} from "@glassact/ui";
import { createFileRoute } from "@tanstack/solid-router";
import {
  createMemo,
  createSignal,
  For,
  Match,
  onCleanup,
  onMount,
  ParentComponent,
  Show,
  Switch,
} from "solid-js";
import {
  draggable,
  dropTargetForElements,
  monitorForElements,
} from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { unsafeOverflowAutoScrollForElements } from "@atlaskit/pragmatic-drag-and-drop-auto-scroll/unsafe-overflow/element";
import { autoScrollForElements } from "@atlaskit/pragmatic-drag-and-drop-auto-scroll/element";
import { combine } from "@atlaskit/pragmatic-drag-and-drop/combine";
import { CleanupFn } from "@atlaskit/pragmatic-drag-and-drop/dist/types/internal-types";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import type { ManufacturingStep } from "@glassact/data";
import {
  getKanbanInlaysOpts,
  patchInlayStep,
  type KanbanInlay,
} from "../../queries/manufacturing";
import { isApiError } from "../../utils/is-api-error";
import { Can } from "../../components/Can";
import { AddInlayUpdateDialog } from "../../components/manufacturing/add-inlay-update-dialog";

export const Route = createFileRoute("/_app/inlays")({
  component: RouteComponent,
});

interface StepColumn {
  id: ManufacturingStep;
  title: string;
}

const STEP_COLUMNS: StepColumn[] = [
  { id: "ordered", title: "Ordered" },
  { id: "materials-prep", title: "Prepping Materials" },
  { id: "cutting", title: "Cutting" },
  { id: "fire-polish", title: "Fire Polish" },
  { id: "packaging", title: "Packaging" },
  { id: "shipped", title: "Shipped" },
  { id: "delivered", title: "Delivered" },
];

interface InlayCardProps {
  inlay: KanbanInlay;
}

function InlayCard(props: InlayCardProps) {
  let ref!: HTMLDivElement;
  const [dragging, setDragging] = createSignal(false);

  let cleanup: CleanupFn;
  onMount(() => {
    if (!ref) return;
    const c1 = draggable({
      element: ref,
      getInitialData: () => ({
        uuid: props.inlay.uuid,
      }),
      onDragStart: () => setDragging(true),
      onDrop: () => setDragging(false),
    });
    cleanup = combine(c1);
  });

  onCleanup(() => {
    if (cleanup) cleanup();
  });

  return (
    <Card ref={ref} class={cn("w-full touch-none", dragging() && "opacity-30")}>
      <CardHeader class="flex-row justify-between items-start gap-2">
        <div class="flex items-start gap-2 min-w-0">
          <Show
            when={props.inlay.preview_url}
            fallback={
              <div class="w-[50px] h-[50px] flex-shrink-0 bg-gray-100 rounded flex items-center justify-center text-gray-400 text-xs">
                N/A
              </div>
            }
          >
            <img
              alt={props.inlay.name}
              src={props.inlay.preview_url}
              class="w-[50px] h-[50px] flex-shrink-0 object-contain rounded"
            />
          </Show>
          <div class="flex flex-col space-y-1 min-w-0">
            <CardTitle class="text-sm truncate">{props.inlay.name}</CardTitle>
            <CardDescription class="text-xs truncate">
              {props.inlay.dealership_name} — {props.inlay.project_name}
            </CardDescription>
          </div>
        </div>
        <Can permission="create_inlay_update">
          <AddInlayUpdateDialog
            inlayUuid={props.inlay.uuid}
            triggerLabel="+ Update"
            triggerClass="flex-shrink-0 text-xs font-medium text-primary hover:underline cursor-pointer"
          />
        </Can>
      </CardHeader>
    </Card>
  );
}

interface StepColumnProps {
  column: StepColumn;
}

const StepColumn: ParentComponent<StepColumnProps> = (props) => {
  let ref!: HTMLDivElement;
  const [draggedOver, setDraggedOver] = createSignal(false);

  let cleanup: CleanupFn;
  onMount(() => {
    if (!ref) return;
    const c1 = dropTargetForElements({
      element: ref,
      getData: () => ({ id: props.column.id }),
      onDragEnter: () => setDraggedOver(true),
      onDragLeave: () => setDraggedOver(false),
      onDrop: () => setDraggedOver(false),
    });
    const c2 = autoScrollForElements({ element: ref });
    cleanup = combine(c1, c2);
  });

  onCleanup(() => {
    if (cleanup) cleanup();
  });

  return (
    <div
      ref={ref}
      class={cn(
        "flex-shrink-0 flex flex-col gap-2 p-4 max-h-[calc(100vh-130px)] w-[300px] max-w-[90%] overflow-y-auto border rounded-lg",
        draggedOver() && "bg-blue-50 border-blue-300",
      )}
    >
      <span class="text-base font-semibold">{props.column.title}</span>
      {props.children}
    </div>
  );
};

function RouteComponent() {
  let ref!: HTMLDivElement;
  const queryClient = useQueryClient();
  const kanbanQuery = useQuery(() => getKanbanInlaysOpts());
  const patchStep = useMutation(() => ({
    mutationFn: patchInlayStep,
    onMutate: async ({
      uuid,
      step,
    }: {
      uuid: string;
      step: ManufacturingStep;
    }) => {
      await queryClient.cancelQueries({ queryKey: ["kanban-inlays"] });
      const previousInlays = queryClient.getQueryData<KanbanInlay[]>([
        "kanban-inlays",
      ]);
      queryClient.setQueryData<KanbanInlay[]>(
        ["kanban-inlays"],
        (old) =>
          old?.map((inlay) =>
            inlay.uuid === uuid
              ? { ...inlay, manufacturing_step: step }
              : inlay,
          ) ?? [],
      );
      return { previousInlays };
    },
    onError: (
      error: Error,
      _variables: { uuid: string; step: ManufacturingStep },
      context: { previousInlays: KanbanInlay[] | undefined } | undefined,
    ) => {
      queryClient.setQueryData(["kanban-inlays"], context?.previousInlays);
      if (isApiError(error)) {
        showToast({
          title: "Failed to move inlay",
          description: error?.data?.error ?? "Unknown error",
          variant: "error",
        });
      }
    },
    onSuccess: () => {
      showToast({
        title: "Inlay Status Updated",
        description: "The inlay status has been updated successfully.",
        variant: "success",
      });
    },
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: ["kanban-inlays"] }),
  }));

  const columnData = createMemo(() => {
    const inlays = kanbanQuery.isSuccess ? kanbanQuery.data : [];
    const map = new Map<ManufacturingStep, KanbanInlay[]>();
    for (const col of STEP_COLUMNS) {
      map.set(col.id, []);
    }
    for (const inlay of inlays) {
      if (inlay.manufacturing_step) {
        const step = inlay.manufacturing_step as ManufacturingStep;
        const existing = map.get(step) ?? [];
        map.set(step, [...existing, inlay]);
      }
    }
    return map;
  });

  let cleanup: CleanupFn;
  onMount(() => {
    const c1 = monitorForElements({
      onDrop({ source, location }) {
        const dest = location.current.dropTargets[0];
        if (!dest) return;

        const { success, data } = z
          .object({
            destId: z.string(),
            cardUuid: z.string(),
          })
          .safeParse({
            destId: dest.data.id,
            cardUuid: source.data.uuid,
          });

        if (!success) return;

        const { destId, cardUuid } = data;
        const destStep = destId as ManufacturingStep;

        patchStep.mutate({ uuid: cardUuid, step: destStep });
      },
    });

    const c2 = autoScrollForElements({ element: ref });

    const c3 = unsafeOverflowAutoScrollForElements({
      element: ref,
      getOverflow: () => ({
        forTopEdge: { top: 6000, right: 6000, left: 6000 },
        forRightEdge: { top: 6000, right: 6000, bottom: 6000 },
        forBottomEdge: { right: 6000, bottom: 6000, left: 6000 },
        forLeftEdge: { top: 6000, left: 6000, bottom: 6000 },
      }),
    });

    cleanup = combine(c1, c2, c3);
  });

  onCleanup(() => {
    if (cleanup) cleanup();
  });

  return (
    <>
      <Switch>
        <Match when={kanbanQuery.isLoading}>
          <div class="flex gap-4 overflow-x-auto">
            <For each={STEP_COLUMNS}>
              {() => (
                <div class="flex-shrink-0 w-[300px] h-[calc(100vh-130px)] bg-gray-100 rounded-lg animate-pulse" />
              )}
            </For>
          </div>
        </Match>

        <Match when={kanbanQuery.isError}>
          <div class="border-2 border-dashed border-red-300 rounded-xl p-8 text-center">
            <p class="text-red-600 font-medium">Failed to load kanban board</p>
            <Button
              variant="outline"
              class="mt-4"
              onClick={() => kanbanQuery.refetch()}
            >
              Retry
            </Button>
          </div>
        </Match>
      </Switch>

      <div ref={ref} class="flex gap-4 overflow-x-auto">
        <For each={STEP_COLUMNS}>
          {(col) => (
            <StepColumn column={col}>
              <For each={columnData().get(col.id) ?? []}>
                {(inlay) => <InlayCard inlay={inlay} />}
              </For>
            </StepColumn>
          )}
        </For>
      </div>
    </>
  );
}
