import { createMemo } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { cn } from "@glassact/ui";
import { getProjectChatsOpts } from "../../queries/chat";
import ChatThread from "./chat-thread";
import ChatInput from "./chat-input";

interface ProjectDiscussionProps {
  projectUuid: string;
  // On an inlay page: tags anything sent from here with that inlay. It never
  // narrows what is shown — the thread is always the whole conversation.
  tagInlayUuid?: string;
  class?: string;
}

// The project's single conversation. Every message belongs to it; messages
// tagged with an inlay carry a link to that inlay so nobody has to guess which
// one "this one" means.
export function ProjectDiscussion(props: ProjectDiscussionProps) {
  const chatsQuery = useQuery(() => getProjectChatsOpts(props.projectUuid));

  const totalCount = createMemo(() => (chatsQuery.data ?? []).length);

  return (
    <section
      class={cn(
        "flex min-h-0 flex-col overflow-hidden rounded-lg border",
        props.class,
      )}
    >
      <div class="flex items-center justify-between gap-2 px-4 py-3">
        <h2 class="text-sm font-semibold text-gray-700">
          Discussion ({totalCount()})
        </h2>
      </div>

      <div class="flex min-h-0 flex-1 flex-col border-t">
        <ChatThread projectUuid={props.projectUuid} />
        <ChatInput
          projectUuid={props.projectUuid}
          inlayUuid={props.tagInlayUuid}
          placeholder={
            props.tagInlayUuid
              ? "Message about this inlay..."
              : "Message about this project..."
          }
        />
      </div>
    </section>
  );
}
