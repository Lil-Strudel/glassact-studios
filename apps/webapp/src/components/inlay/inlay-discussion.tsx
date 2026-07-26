import { Show, createMemo, createSignal } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { IoChevronDown, IoChevronForward } from "solid-icons/io";
import { getInlayChatsOpts } from "../../queries/chat";
import ChatThread from "../chat/chat-thread";
import ChatInput from "../chat/chat-input";

interface InlayDiscussionProps {
  inlayUuid: string;
  projectUuid: string;
  // A customer-facing proof is waiting on this inlay, so the conversation is
  // likely why the user is here.
  expandByDefault: boolean;
}

// Per-inlay chat, collapsed until it matters. Stock and customized-catalog
// inlays almost never accumulate messages, so the thread earns its space rather
// than being handed it.
export function InlayDiscussion(props: InlayDiscussionProps) {
  const chatsQuery = useQuery(() => getInlayChatsOpts(props.inlayUuid));
  const messageCount = createMemo(() => (chatsQuery.data ?? []).length);

  const [wasToggled, setWasToggled] = createSignal(false);
  const [manualState, setManualState] = createSignal(false);

  const isOpen = createMemo(() =>
    wasToggled()
      ? manualState()
      : props.expandByDefault || messageCount() > 0,
  );

  function toggle() {
    setManualState(!isOpen());
    setWasToggled(true);
  }

  return (
    <section class="overflow-hidden rounded-lg border">
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
          Discussion ({messageCount()})
        </h2>
      </button>

      <Show when={isOpen()}>
        <div
          class="flex flex-col border-t"
          style={{ "min-height": "320px", "max-height": "520px" }}
        >
          <ChatThread
            inlayUuid={props.inlayUuid}
            projectUuid={props.projectUuid}
          />
          <ChatInput inlayUuid={props.inlayUuid} />
        </div>
      </Show>
    </section>
  );
}
