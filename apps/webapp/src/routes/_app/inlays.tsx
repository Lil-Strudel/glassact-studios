import {
  Alert,
  cn,
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  showToast,
} from "@glassact/ui";
import { createFileRoute, Link } from "@tanstack/solid-router";
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
import { IoAlertCircleOutline } from "solid-icons/io";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import type { ManufacturingStep } from "@glassact/data";
import {
  getKanbanInlaysOpts,
  patchInlayStep,
  type KanbanInlay,
} from "../../queries/manufacturing";
import { isApiError } from "../../utils/is-api-error";
import { Can } from "../../components/Can";
import { MirrorBadge } from "../../components/inlay/mirror-badge";
import { mirrorCssTransform } from "../../utils/inlay-mirror";
import { AddInlayUpdateDialog } from "../../components/manufacturing/add-inlay-update-dialog";
import { STEP_LABELS, STEP_ORDER } from "../../components/manufacturing/steps";
import { PageHeader } from "@glassact/ui";

export const Route = createFileRoute("/_app/inlays")({
  component: RouteComponent,
});

interface StepColumn {
  id: ManufacturingStep;
  title: string;
}

const STEP_COLUMNS: StepColumn[] = STEP_ORDER.map((id) => ({
  id,
  title: STEP_LABELS[id],
}));

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
    <Card
      ref={ref}
      class={cn(
        "w-full cursor-grab touch-none rounded-lg shadow-none transition-colors hover:border-input",
        dragging() && "opacity-30",
      )}
    >
      <CardHeader class="flex-row items-start justify-between gap-2 space-y-0 p-3">
        <Link
          to="/projects/$id/inlay/$inlayId"
          params={{
            id: props.inlay.project_uuid,
            inlayId: props.inlay.uuid,
          }}
          class="group flex min-w-0 items-start gap-3"
          draggable={false}
        >
          <Show
            when={props.inlay.preview_url}
            fallback={
              <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-md border border-dashed border-input text-xs text-faint">
                N/A
              </div>
            }
          >
            <img
              alt={props.inlay.name}
              src={props.inlay.preview_url}
              class="h-12 w-12 flex-shrink-0 rounded-md border border-input bg-plaque object-contain p-0.5"
              style={{ transform: mirrorCssTransform(props.inlay) }}
              draggable={false}
            />
          </Show>
          <div class="flex min-w-0 flex-col gap-0.5">
            <CardTitle class="truncate text-base leading-snug group-hover:text-crimson-light">
              {props.inlay.name}
            </CardTitle>
            <CardDescription class="truncate text-xs text-faint">
              {props.inlay.dealership_name} — {props.inlay.project_name}
            </CardDescription>
            <MirrorBadge inlay={props.inlay} />
          </div>
        </Link>
        <Can permission="create_inlay_update">
          <AddInlayUpdateDialog
            inlayUuid={props.inlay.uuid}
            triggerLabel="+ Update"
            triggerClass="flex-shrink-0 cursor-pointer text-xs font-medium text-crimson-light underline-offset-4 hover:underline"
          />
        </Can>
      </CardHeader>
    </Card>
  );
}

interface StepColumnProps {
  column: StepColumn;
  position: number;
  inlayCount: number;
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
        "flex w-[300px] max-w-[90%] flex-shrink-0 flex-col overflow-y-auto rounded-lg border bg-card transition-colors max-h-[calc(100dvh-16rem)]",
        draggedOver() ? "border-gilt/60 bg-accent/50" : "border-border",
      )}
    >
      <div class="sticky top-0 z-10 flex items-center gap-2.5 border-b border-border bg-card px-4 py-3">
        <span class="lozenge size-2 text-gilt" aria-hidden="true" />
        <div class="min-w-0">
          <p class="eyebrow text-[0.65rem] leading-none text-faint">
            Step {props.position}
          </p>
          <h2 class="truncate text-lg leading-tight">{props.column.title}</h2>
        </div>
        <span class="ml-auto rounded-full border border-border px-2 py-0.5 text-xs tabular-nums text-muted-foreground">
          {props.inlayCount}
        </span>
      </div>
      <div class="flex flex-col gap-2 p-3">
        <Show
          when={props.inlayCount > 0}
          fallback={
            <p class="py-6 text-center text-xs text-faint">
              No inlays at this step
            </p>
          }
        >
          {props.children}
        </Show>
      </div>
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
    <div class="flex flex-col gap-6">
      <PageHeader eyebrow="Production" title="Inlays" />

      <Switch>
        <Match when={kanbanQuery.isLoading}>
          <div class="flex gap-4 overflow-x-auto">
            <For each={STEP_COLUMNS}>
              {() => (
                <div class="h-[calc(100dvh-16rem)] w-[300px] flex-shrink-0 animate-pulse rounded-lg bg-card" />
              )}
            </For>
          </div>
        </Match>

        <Match when={kanbanQuery.isError}>
          <Alert variant="destructive">
            <IoAlertCircleOutline size={18} />
            <div class="flex flex-wrap items-center justify-between gap-3">
              <p class="font-medium">Failed to load kanban board</p>
              <Button
                variant="outline"
                size="sm"
                onClick={() => kanbanQuery.refetch()}
              >
                Retry
              </Button>
            </div>
          </Alert>
        </Match>
      </Switch>

      <div ref={ref} class="flex gap-4 overflow-x-auto pb-2">
        <For each={STEP_COLUMNS}>
          {(col, index) => (
            <StepColumn
              column={col}
              position={index() + 1}
              inlayCount={columnData().get(col.id)?.length ?? 0}
            >
              <For each={columnData().get(col.id) ?? []}>
                {(inlay) => <InlayCard inlay={inlay} />}
              </For>
            </StepColumn>
          )}
        </For>
      </div>
    </div>
  );
}
