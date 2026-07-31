import { cn } from "@glassact/ui";
import type { GET, ProjectChat, ChatMessageType } from "@glassact/data";
import { useQuery } from "@tanstack/solid-query";
import { Link } from "@tanstack/solid-router";
import {
  createEffect,
  createMemo,
  For,
  onCleanup,
  Show,
  type Component,
} from "solid-js";
import { getProjectChatsOpts } from "../../queries/chat";
import { getInlaysByProjectOpts } from "../../queries/inlay";
import { useUserContext } from "../../providers/user";
import { ProofStatusBadge } from "../proof/proof-status-badge";

interface ChatThreadProps {
  projectUuid: string;
}

const SYSTEM_TYPES: ChatMessageType[] = [
  "proof_sent",
  "proof_approved",
  "proof_declined",
  "system",
];

function formatTimestamp(dateStr: string): string {
  return new Date(dateStr).toLocaleString();
}

const ChatThread: Component<ChatThreadProps> = (props) => {
  let scrollRef: HTMLDivElement | undefined;
  const { isInternal } = useUserContext();

  const query = useQuery(() => getProjectChatsOpts(props.projectUuid));
  // Messages carry inlay_id; routes need the uuid, so the whole inlay is needed
  // to resolve a tag into a name and a link.
  const inlaysQuery = useQuery(() => getInlaysByProjectOpts(props.projectUuid));

  const inlaysById = createMemo(() => {
    const map = new Map<number, { uuid: string; name: string }>();
    for (const inlay of inlaysQuery.data ?? []) {
      map.set(inlay.id, { uuid: inlay.uuid, name: inlay.name });
    }
    return map;
  });

  const messages = createMemo(() => query.data ?? []);

  const isSystemMessage = (chat: GET<ProjectChat>) =>
    SYSTEM_TYPES.includes(chat.message_type);

  const isMyMessage = (chat: GET<ProjectChat>) => {
    if (isInternal()) {
      return chat.internal_user_id !== null;
    }
    return chat.dealership_user_id !== null;
  };

  // Always rendered when a message is tagged: an untagged "this one is too dark"
  // in a project-wide thread is unactionable without it.
  const InlayTag: Component<{ chat: GET<ProjectChat>; onDark?: boolean }> = (
    tagProps,
  ) => {
    const inlay = createMemo(() => {
      const id = tagProps.chat.inlay_id;
      return id === null ? undefined : inlaysById().get(id);
    });

    return (
      <Show when={inlay()}>
        {(found) => (
          <Link
            to="/projects/$id/inlay/$inlayId"
            params={{ id: props.projectUuid, inlayId: found().uuid }}
            class={cn(
              "mb-1 inline-block max-w-full truncate text-[11px] font-medium not-italic hover:underline",
              tagProps.onDark ? "text-blue-100" : "text-blue-600",
            )}
          >
            {found().name}
          </Link>
        )}
      </Show>
    );
  };

  function scrollToNewest() {
    if (scrollRef) {
      scrollRef.scrollTop = scrollRef.scrollHeight;
    }
  }

  createEffect(() => {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const _ = messages().length; // Needed for reactivity with scroll
    scrollToNewest();
  });

  // Navigating to a page whose chat is already cached renders the whole list
  // before the pane has been laid out, so the effect above scrolls a box that
  // is still zero-height — and the message count never changes to trigger a
  // retry. Pinning on resize catches that first layout.
  function attachScroller(el: HTMLDivElement) {
    scrollRef = el;
    const observer = new ResizeObserver(scrollToNewest);
    observer.observe(el);
    onCleanup(() => observer.disconnect());
  }

  return (
    <Show
      when={!query.isLoading}
      fallback={
        <div class="flex-1 flex items-center justify-center text-gray-500">
          Loading messages...
        </div>
      }
    >
      <Show
        when={messages().length > 0}
        fallback={
          <div class="flex-1 flex items-center justify-center text-gray-500 text-sm text-center px-4">
            No messages yet. Start the conversation!
          </div>
        }
      >
        <div ref={attachScroller} class="flex-1 overflow-y-auto p-4 space-y-3">
          <For each={messages()}>
            {(chat) => (
              <Show
                when={!isSystemMessage(chat)}
                fallback={
                  <div class="flex justify-center">
                    <div
                      class={cn(
                        "px-4 py-2 rounded-lg text-xs italic text-center max-w-md",
                        chat.message_type === "proof_sent" &&
                          "bg-gray-50 text-gray-600 border border-gray-200",
                        chat.message_type === "proof_approved" &&
                          "bg-green-50 text-green-700 border border-green-200",
                        chat.message_type === "proof_declined" &&
                          "bg-red-50 text-red-700 border border-red-200",
                        chat.message_type === "system" &&
                          "bg-gray-50 text-gray-500",
                      )}
                    >
                      <InlayTag chat={chat} />
                      <Show when={chat.message_type === "proof_sent"}>
                        <div class="flex flex-col items-center gap-2">
                          <ProofStatusBadge status="pending" />
                          <p class="not-italic">{chat.message}</p>
                          <Show when={chat.attachment_url}>
                            <a
                              href={chat.attachment_url!}
                              target="_blank"
                              rel="noopener noreferrer"
                              class="text-blue-600 underline text-xs"
                            >
                              View Proof
                            </a>
                          </Show>
                        </div>
                      </Show>
                      <Show when={chat.message_type === "proof_approved"}>
                        <div class="flex flex-col items-center gap-2">
                          <ProofStatusBadge status="approved" />
                          <p class="not-italic">{chat.message}</p>
                        </div>
                      </Show>
                      <Show when={chat.message_type === "proof_declined"}>
                        <div class="flex flex-col items-center gap-2">
                          <ProofStatusBadge status="declined" />
                          <p class="not-italic">{chat.message}</p>
                        </div>
                      </Show>
                      <Show when={chat.message_type === "system"}>
                        <p>{chat.message}</p>
                      </Show>
                      <p class="text-[10px] text-gray-400 mt-1">
                        {formatTimestamp(chat.created_at)}
                      </p>
                    </div>
                  </div>
                }
              >
                <div
                  class={cn(
                    "flex",
                    isMyMessage(chat) ? "justify-end" : "justify-start",
                  )}
                >
                  <div
                    class={cn(
                      "max-w-xs lg:max-w-md px-4 py-2 rounded-lg",
                      isMyMessage(chat)
                        ? "bg-blue-500 text-white"
                        : "bg-gray-100 text-gray-900",
                    )}
                  >
                    <InlayTag chat={chat} onDark={isMyMessage(chat)} />
                    <p class="text-sm">{chat.message}</p>
                    <p
                      class={cn(
                        "text-xs mt-1",
                        isMyMessage(chat) ? "text-blue-200" : "text-gray-500",
                      )}
                    >
                      {formatTimestamp(chat.created_at)}
                    </p>
                  </div>
                </div>
              </Show>
            )}
          </For>
        </div>
      </Show>
    </Show>
  );
};

export default ChatThread;
