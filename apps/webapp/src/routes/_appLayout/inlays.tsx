import {
  cn,
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  Button,
} from "@glassact/ui";
import { createFileRoute } from "@tanstack/solid-router";
import { createSignal, For, JSXElement, onCleanup, onMount } from "solid-js";
import {
  draggable,
  dropTargetForElements,
  monitorForElements,
} from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { autoScrollForElements } from "@atlaskit/pragmatic-drag-and-drop-auto-scroll/element";
import { combine } from "@atlaskit/pragmatic-drag-and-drop/combine";
import { CleanupFn } from "@atlaskit/pragmatic-drag-and-drop/dist/types/internal-types";
import { z } from "zod";
import { produce } from "immer";
import { IoWarningOutline } from "solid-icons/io";

export const Route = createFileRoute("/_appLayout/inlays")({
  component: RouteComponent,
});

function DraggableDiv(props: { id: number }) {
  let ref!: HTMLDivElement;
  const [dragging, setDragging] = createSignal(false);

  let cleanup: CleanupFn;
  onMount(() => {
    if (!ref) return;
    const c1 = draggable({
      element: ref,
      getInitialData: () => ({ id: props.id }),
      onDragStart: () => setDragging(true),
      onDrop: () => setDragging(false),
    });

    cleanup = combine(c1);
  });

  onCleanup(() => {
    if (cleanup) {
      cleanup();
    }
  });

  return (
    <Card ref={ref} class={cn("w-full touch-none", dragging() && "opacity-30")}>
      <CardHeader class="flex-row justify-between">
        <div class="flex items-center gap-2">
          <img
            alt="inlay name"
            src="https://placehold.co/75x75"
            class="w-[50px]"
          />
          <div class="flex flex-col space-y-1.5">
            <CardTitle>Inlay Name</CardTitle>
            <CardDescription>Dealership - Project</CardDescription>
          </div>
        </div>
        <Button variant="ghost" size="icon">
          <IoWarningOutline size={24} />
        </Button>
      </CardHeader>
    </Card>
  );
}

function DroppableDiv(props: { id: number; children: JSXElement }) {
  let ref!: HTMLDivElement;
  const [draggedOver, setDraggedOver] = createSignal(false);

  let cleanup: CleanupFn;
  onMount(() => {
    if (!ref) return;
    const c1 = dropTargetForElements({
      element: ref,
      getData: () => ({ id: props.id }),
      onDragEnter: () => setDraggedOver(true),
      onDragLeave: () => setDraggedOver(false),
      onDrop: () => setDraggedOver(false),
    });

    const c2 = autoScrollForElements({
      element: ref,
    });

    cleanup = combine(c1, c2);
  });

  onCleanup(() => {
    if (cleanup) {
      cleanup();
    }
  });

  return (
    <div
      ref={ref}
      class={cn(
        "flex-shrink-0 flex flex-col gap-2 p-4 max-h-[calc(100vh-130px)] w-[300px] max-w-[90%] overflow-y-auto border rounded-lg",
        draggedOver() && "bg-blue-100",
      )}
    >
      <span class="text-xl">Title</span>
      {props.children}
    </div>
  );
}

interface Column {
  id: number;
  cards: number[];
}

type Data = Column[];

function RouteComponent() {
  let ref!: HTMLDivElement;
  const [data, setData] = createSignal<Data>([
    { id: 0, cards: [0, 1, 2] },
    { id: 1, cards: [3, 4, 5] },
    { id: 2, cards: [6, 7, 8] },
    { id: 3, cards: [9, 10, 11] },
    { id: 4, cards: [12, 13, 14] },
    { id: 5, cards: [15, 16, 17] },
    {
      id: 6,
      cards: [18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32],
    },
    { id: 7, cards: [33, 34] },
    { id: 8, cards: [35] },
  ]);

  let cleanup: CleanupFn;
  onMount(() => {
    const c1 = monitorForElements({
      onDrop({ source, location }) {
        const dest = location.current.dropTargets[0];
        if (!dest) return;

        const { success, data } = z
          .object({ destId: z.number(), cardId: z.number() })
          .safeParse({ destId: dest.data.id, cardId: source.data.id });
        if (!success) return;

        const { destId, cardId } = data;

        setData(
          produce((oldData: Data) => {
            const srcIdx = oldData.findIndex((col) =>
              col.cards.includes(cardId),
            );
            const destIdx = oldData.findIndex((col) => col.id === destId);

            if (srcIdx === -1 || destIdx === -1) return;

            oldData[srcIdx].cards = oldData[srcIdx].cards.filter(
              (card) => card !== cardId,
            );
            oldData[destIdx].cards.push(cardId);
          }),
        );
      },
    });

    const c2 = autoScrollForElements({
      element: ref,
    });

    cleanup = combine(c1, c2);
  });

  onCleanup(() => {
    if (cleanup) {
      cleanup();
    }
  });

  return (
    <div ref={ref} class="flex gap-4 overflow-x-auto">
      <For each={data()}>
        {(item) => (
          <DroppableDiv id={item.id}>
            <For each={item.cards}>
              {(cardId) => <DraggableDiv id={cardId} />}
            </For>
          </DroppableDiv>
        )}
      </For>
    </div>
  );
}
