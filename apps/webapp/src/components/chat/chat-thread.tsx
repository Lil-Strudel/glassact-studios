import { Badge, cn } from "@glassact/ui";
import type { GET, InlayChat, ChatMessageType } from "@glassact/data";
import { useQuery } from "@tanstack/solid-query";
import { createEffect, createMemo, For, Show, type Component } from "solid-js";
import { getInlayChatsOpts } from "../../queries/chat";

interface ChatThreadProps {
  inlayUuid: string;
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

  const query = useQuery(() => getInlayChatsOpts(props.inlayUuid));

  const messages = createMemo(() => query.data ?? []);

  const isSystemMessage = (chat: GET<InlayChat>) =>
    SYSTEM_TYPES.includes(chat.message_type);

  const isDealershipMessage = (chat: GET<InlayChat>) =>
    chat.dealership_user_id !== null;

  createEffect(() => {
    const _ = messages().length;
    if (scrollRef) {
      scrollRef.scrollTop = scrollRef.scrollHeight;
    }
  });

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
          <div class="flex-1 flex items-center justify-center text-gray-500">
            No messages yet. Start the conversation!
          </div>
        }
      >
        <div
          ref={scrollRef}
          class="flex-1 overflow-y-auto p-4 space-y-3"
        >
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
                      <Show when={chat.message_type === "proof_sent"}>
                        <div class="flex flex-col items-center gap-1">
                          <Badge variant="outline" class="bg-blue-50 text-blue-700 border-blue-200">
                            Proof Sent
                          </Badge>
                          <p>{chat.message}</p>
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
                        <div class="flex flex-col items-center gap-1">
                          <Badge variant="outline" class="bg-green-50 text-green-700 border-green-200">
                            Proof Approved
                          </Badge>
                          <p>{chat.message}</p>
                        </div>
                      </Show>
                      <Show when={chat.message_type === "proof_declined"}>
                        <div class="flex flex-col items-center gap-1">
                          <Badge variant="outline" class="bg-red-50 text-red-700 border-red-200">
                            Proof Declined
                          </Badge>
                          <p>{chat.message}</p>
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
                    isDealershipMessage(chat)
                      ? "justify-end"
                      : "justify-start",
                  )}
                >
                  <div
                    class={cn(
                      "max-w-xs lg:max-w-md px-4 py-2 rounded-lg",
                      isDealershipMessage(chat)
                        ? "bg-blue-500 text-white"
                        : "bg-gray-100 text-gray-900",
                    )}
                  >
                    <p class="text-sm">{chat.message}</p>
                    <p
                      class={cn(
                        "text-xs mt-1",
                        isDealershipMessage(chat)
                          ? "text-blue-200"
                          : "text-gray-500",
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
