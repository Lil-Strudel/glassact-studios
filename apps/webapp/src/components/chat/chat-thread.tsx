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
  return new Date(dateStr).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
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
  const InlayTag: Component<{ chat: GET<ProjectChat> }> = (tagProps) => {
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
            class="inline-flex max-w-full items-center gap-1.5 text-xs font-medium text-crimson-light hover:underline"
          >
            <span
              class="size-1.5 shrink-0 rotate-45 border border-current"
              aria-hidden="true"
            />
            <span class="truncate">{found().name}</span>
          </Link>
        )}
      </Show>
    );
  };

  const isStaffMessage = (chat: GET<ProjectChat>) =>
    chat.internal_user_id !== null;

  const senderLabel = (chat: GET<ProjectChat>) => {
    if (isMyMessage(chat)) return "You";
    return isStaffMessage(chat) ? "GlassAct Studios" : "Dealership";
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
        <div class="flex flex-1 items-center justify-center text-sm text-faint">
          Loading messages...
        </div>
      }
    >
      <Show
        when={messages().length > 0}
        fallback={
          <div class="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
            <div class="ornament-rule w-24" aria-hidden="true">
              <span class="lozenge" />
            </div>
            <p class="font-display text-lg text-foreground">No messages yet</p>
            <p class="max-w-xs text-sm text-muted-foreground">
              Questions and proof updates for this project will appear here.
            </p>
          </div>
        }
      >
        <div ref={attachScroller} class="flex-1 space-y-4 overflow-y-auto p-4">
          <For each={messages()}>
            {(chat) => (
              <Show
                when={!isSystemMessage(chat)}
                fallback={
                  <div
                    class={cn(
                      "flex flex-col gap-1.5 rounded-lg border border-dashed px-3 py-2.5 text-sm",
                      chat.message_type === "proof_sent" &&
                        "border-gilt/45 bg-gilt/5",
                      chat.message_type === "proof_approved" &&
                        "border-success-foreground/35 bg-success/60",
                      chat.message_type === "proof_declined" &&
                        "border-error-foreground/35 bg-error/60",
                      chat.message_type === "system" &&
                        "border-border bg-muted text-muted-foreground",
                    )}
                  >
                    <div class="flex flex-wrap items-center justify-between gap-2">
                      <Show when={chat.message_type === "proof_sent"}>
                        <ProofStatusBadge status="pending" />
                      </Show>
                      <Show when={chat.message_type === "proof_approved"}>
                        <ProofStatusBadge status="approved" />
                      </Show>
                      <Show when={chat.message_type === "proof_declined"}>
                        <ProofStatusBadge status="declined" />
                      </Show>
                      <span class="ml-auto text-[11px] tabular-nums text-faint">
                        {formatTimestamp(chat.created_at)}
                      </span>
                    </div>
                    <p class="text-foreground">{chat.message}</p>
                    <div class="flex flex-wrap items-center justify-between gap-2">
                      <InlayTag chat={chat} />
                      <Show
                        when={
                          chat.message_type === "proof_sent" &&
                          chat.attachment_url
                        }
                      >
                        <a
                          href={chat.attachment_url!}
                          target="_blank"
                          rel="noopener noreferrer"
                          class="text-xs text-crimson-light hover:underline"
                        >
                          View proof
                        </a>
                      </Show>
                    </div>
                  </div>
                }
              >
                <div
                  class={cn(
                    "flex flex-col gap-1",
                    isMyMessage(chat) ? "items-end" : "items-start",
                  )}
                >
                  <div class="flex flex-wrap items-baseline gap-2 text-xs text-faint">
                    <span
                      class={cn(
                        "font-semibold",
                        isStaffMessage(chat) && !isMyMessage(chat)
                          ? "text-gilt"
                          : "text-foreground",
                      )}
                    >
                      {senderLabel(chat)}
                    </span>
                    <span class="tabular-nums">
                      {formatTimestamp(chat.created_at)}
                    </span>
                  </div>
                  <div
                    class={cn(
                      "flex max-w-[85%] flex-col gap-1 rounded-lg border px-3 py-2",
                      isMyMessage(chat)
                        ? "border-primary/40 bg-crimson-deep/50"
                        : "border-border bg-secondary",
                    )}
                  >
                    <p class="whitespace-pre-wrap text-sm text-foreground">
                      {chat.message}
                    </p>
                    <InlayTag chat={chat} />
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
