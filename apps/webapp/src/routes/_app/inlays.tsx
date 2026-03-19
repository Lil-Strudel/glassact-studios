import {
  cn,
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  showToast,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
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
import { IoWarningOutline } from "solid-icons/io";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import type { ManufacturingStep } from "@glassact/data";
import {
  getKanbanInlaysOpts,
  patchInlayStep,
  type KanbanInlay,
} from "../../queries/manufacturing";
import { BlockersDialog } from "../../components/manufacturing/blockers-dialog";
import { isApiError } from "../../utils/is-api-error";

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
  onOpenBlockers: (uuid: string) => void;
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
        hasHardBlocker: props.inlay.has_hard_blocker,
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
        <Show when={props.inlay.has_hard_blocker}>
          <Button
            variant="ghost"
            size="icon"
            class="text-red-500 hover:text-red-600 hover:bg-red-50 flex-shrink-0"
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              props.onOpenBlockers(props.inlay.uuid);
            }}
          >
            <IoWarningOutline size={20} />
          </Button>
        </Show>
        <Show when={!props.inlay.has_hard_blocker}>
          <Button
            variant="ghost"
            size="icon"
            class="text-gray-400 hover:text-gray-600 flex-shrink-0"
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              props.onOpenBlockers(props.inlay.uuid);
            }}
          >
            <IoWarningOutline size={20} />
          </Button>
        </Show>
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

  const [blockersDialogUuid, setBlockersDialogUuid] = createSignal<
    string | null
  >(null);
  const [hardBlockerWarningUuid, setHardBlockerWarningUuid] = createSignal<
    string | null
  >(null);

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
            hasHardBlocker: z.boolean(),
          })
          .safeParse({
            destId: dest.data.id,
            cardUuid: source.data.uuid,
            hasHardBlocker: source.data.hasHardBlocker,
          });

        if (!success) return;

        const { destId, cardUuid, hasHardBlocker } = data;
        const destStep = destId as ManufacturingStep;

        if (hasHardBlocker) {
          setHardBlockerWarningUuid(cardUuid);
          return;
        }

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
                {(inlay) => (
                  <InlayCard
                    inlay={inlay}
                    onOpenBlockers={setBlockersDialogUuid}
                  />
                )}
              </For>
            </StepColumn>
          )}
        </For>
      </div>

      <Show when={blockersDialogUuid()}>
        {(uuid) => (
          <BlockersDialog
            inlayUuid={uuid()}
            open={true}
            onOpenChange={(open) => {
              if (!open) setBlockersDialogUuid(null);
            }}
          />
        )}
      </Show>

      <Dialog
        open={hardBlockerWarningUuid() !== null}
        onOpenChange={(open) => {
          if (!open) setHardBlockerWarningUuid(null);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Cannot Move Inlay</DialogTitle>
          </DialogHeader>
          <div class="space-y-3">
            <p class="text-sm text-gray-600">
              This inlay has unresolved hard blockers. You must resolve all hard
              blockers before moving it to the next step.
            </p>
            <Button
              class="w-full"
              onClick={() => {
                const uuid = hardBlockerWarningUuid();
                setHardBlockerWarningUuid(null);
                if (uuid) setBlockersDialogUuid(uuid);
              }}
            >
              View Blockers
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
