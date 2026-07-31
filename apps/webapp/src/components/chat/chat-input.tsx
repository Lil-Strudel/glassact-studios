import { createSignal } from "solid-js";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { TextArea, TextFieldRoot, Button } from "@glassact/ui";
import { postProjectChatOpts } from "../../queries/chat";
import type { Component } from "solid-js";

interface ChatInputProps {
  projectUuid: string;
  // When present, messages sent from here are tagged with this inlay.
  inlayUuid?: string;
  placeholder?: string;
  onMessageSent?: () => void;
}

const ChatInput: Component<ChatInputProps> = (props) => {
  const [message, setMessage] = createSignal("");
  const queryClient = useQueryClient();
  const mutation = useMutation(() => postProjectChatOpts());

  const handleSubmit = () => {
    const text = message().trim();
    if (!text) return;

    mutation.mutate(
      {
        projectUuid: props.projectUuid,
        body: {
          message: text,
          message_type: "text",
          ...(props.inlayUuid ? { inlay_uuid: props.inlayUuid } : {}),
        },
      },
      {
        onSuccess() {
          setMessage("");
          queryClient.invalidateQueries({
            queryKey: ["project", props.projectUuid, "chats"],
          });
          props.onMessageSent?.();
        },
      },
    );
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <div class="flex items-end gap-2 p-4 border-t">
      <TextFieldRoot class="flex-1">
        <TextArea
          value={message()}
          onInput={(e: InputEvent & { currentTarget: HTMLTextAreaElement }) =>
            setMessage(e.currentTarget.value)
          }
          onKeyDown={handleKeyDown}
          placeholder={props.placeholder ?? "Type a message..."}
        />
      </TextFieldRoot>
      <Button
        variant="default"
        size="sm"
        disabled={!message().trim() || mutation.isPending}
        onClick={handleSubmit}
      >
        Send
      </Button>
    </div>
  );
};

export default ChatInput;
