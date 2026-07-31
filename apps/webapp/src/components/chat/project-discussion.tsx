import { Show, createMemo, createSignal } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { cn } from "@glassact/ui";
import { IoChevronDown, IoChevronForward } from "solid-icons/io";
import { getProjectChatsOpts } from "../../queries/chat";
import { getInlaysByProjectOpts } from "../../queries/inlay";
import ChatThread from "./chat-thread";
import ChatInput from "./chat-input";

interface ProjectDiscussionProps {
  projectUuid: string;
  // On an inlay page: filters the thread and tags anything sent from here.
  focusInlayUuid?: string;
  // A customer-facing proof is waiting, so the conversation is likely why the
  // user is here.
  expandByDefault?: boolean;
  // The project page shows the thread outright; the inlay page keeps it
  // collapsed until it earns the space.
  collapsible?: boolean;
  class?: string;
}

// The project's single conversation. Every message belongs to it; messages
// tagged with an inlay carry a link to that inlay so nobody has to guess which
// one "this one" means.
export function ProjectDiscussion(props: ProjectDiscussionProps) {
  const chatsQuery = useQuery(() => getProjectChatsOpts(props.projectUuid));
  const inlaysQuery = useQuery(() => getInlaysByProjectOpts(props.projectUuid));

  const [showOnlyFocused, setShowOnlyFocused] = createSignal(true);

  const focusInlayId = createMemo(() => {
    if (!props.focusInlayUuid) return null;
    const match = (inlaysQuery.data ?? []).find(
      (inlay) => inlay.uuid === props.focusInlayUuid,
    );
    return match ? match.id : null;
  });

  const totalCount = createMemo(() => (chatsQuery.data ?? []).length);

  const focusedCount = createMemo(() => {
    const id = focusInlayId();
    if (id === null) return 0;
    return (chatsQuery.data ?? []).filter((chat) => chat.inlay_id === id).length;
  });

  const isCollapsible = createMemo(() => props.collapsible ?? false);

  const [wasToggled, setWasToggled] = createSignal(false);
  const [manualState, setManualState] = createSignal(false);

  const isOpen = createMemo(() => {
    if (!isCollapsible()) return true;
    return wasToggled()
      ? manualState()
      : (props.expandByDefault ?? false) || totalCount() > 0;
  });

  function toggle() {
    setManualState(!isOpen());
    setWasToggled(true);
  }

  return (
    <section
      class={cn("flex flex-col overflow-hidden rounded-lg border", props.class)}
    >
      <Show
        when={isCollapsible()}
        fallback={
          <div class="flex items-center justify-between gap-2 px-4 py-3">
            <h2 class="text-sm font-semibold text-gray-700">
              Discussion ({totalCount()})
            </h2>
          </div>
        }
      >
        <button
          type="button"
          onClick={toggle}
          class="flex w-full items-center gap-2 px-4 py-3 text-left transition-colors hover:bg-gray-50"
        >
          <span class="text-gray-400">
            <Show when={isOpen()} fallback={<IoChevronForward size={14} />}>
              <IoChevronDown size={14} />
            </Show>
          </span>
          <h2 class="text-sm font-semibold text-gray-700">
            Discussion ({totalCount()})
          </h2>
        </button>
      </Show>

      <Show when={isOpen()}>
        <Show when={props.focusInlayUuid}>
          <div class="flex items-center gap-1 border-t px-4 py-2 text-xs">
            <button
              type="button"
              onClick={() => setShowOnlyFocused(true)}
              class={cn(
                "rounded px-2 py-1 transition-colors",
                showOnlyFocused()
                  ? "bg-gray-900 text-white"
                  : "text-gray-500 hover:bg-gray-100",
              )}
            >
              This inlay ({focusedCount()})
            </button>
            <button
              type="button"
              onClick={() => setShowOnlyFocused(false)}
              class={cn(
                "rounded px-2 py-1 transition-colors",
                showOnlyFocused()
                  ? "text-gray-500 hover:bg-gray-100"
                  : "bg-gray-900 text-white",
              )}
            >
              Whole project ({totalCount()})
            </button>
          </div>
        </Show>

        <div
          class="flex min-h-0 flex-1 flex-col border-t"
          style={{ "min-height": "320px" }}
        >
          <ChatThread
            projectUuid={props.projectUuid}
            focusInlayUuid={props.focusInlayUuid}
            showOnlyFocused={Boolean(props.focusInlayUuid) && showOnlyFocused()}
          />
          <ChatInput
            projectUuid={props.projectUuid}
            inlayUuid={props.focusInlayUuid}
            placeholder={
              props.focusInlayUuid
                ? "Message about this inlay..."
                : "Message about this project..."
            }
          />
        </div>
      </Show>
    </section>
  );
}
