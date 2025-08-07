import { cn } from "@glassact/ui";
import { createFileRoute } from "@tanstack/solid-router";
import { createSignal, For, JSXElement, onCleanup, onMount } from "solid-js";

export const Route = createFileRoute("/_appLayout/inlays")({
  component: RouteComponent,
});

const colors = [
  "bg-red-400",
  "bg-green-400",
  "bg-gray-400",
  "bg-purple-400",
  "bg-pink-400",
  "bg-yellow-400",
  "bg-indigo-400",
  "bg-teal-400",
  "bg-cyan-400",
];

function DraggableDiv(props: { id: number }) {
  let ref!: HTMLDivElement;
  const [dragging, setDragging] = createSignal(false);

  const color = colors[props.id];
  return (
    <div
      ref={ref}
      class={cn(
        "w-[175px] h-[40px] touch-none",
        color,
        dragging() && "opacity-30",
      )}
    />
  );
}

function DroppableDiv(props: { id: number; children: JSXElement }) {
  let ref!: HTMLDivElement;
  const [draggedOver, setDraggedOver] = createSignal(false);

  return (
    <div
      ref={ref}
      class={cn(
        "bg-blue-100 flex flex-col gap-2 p-4",
        draggedOver() && "bg-blue-300",
      )}
    >
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
  const [data, setData] = createSignal<Data>([
    { id: 0, cards: [0, 1, 2] },
    { id: 1, cards: [3, 4, 5] },
    { id: 2, cards: [6, 7, 8] },
  ]);

  return (
    <div class="flex gap-4">
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
