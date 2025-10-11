import { GET, Inlay } from "@glassact/data";
import { Button, cn, TextField, TextFieldRoot } from "@glassact/ui";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { createSignal, Index, Show, type Component } from "solid-js";
import {
  getInlayChatsByInlayUUIDOpts,
  postInlayChatOpts,
} from "../queries/inlay-chat";

interface InlayChatboxProps {
  inlay: () => GET<Inlay>;
}
const InlayChatbox: Component<InlayChatboxProps> = (props) => {
  const query = useQuery(() =>
    getInlayChatsByInlayUUIDOpts(props.inlay().uuid),
  );
  const postInlayChat = useMutation(postInlayChatOpts);
  const queryClient = useQueryClient();

  const [newMessage, setNewMessage] = createSignal("");
  const [isDeclineFeedback, setIsDeclineFeedback] = createSignal(false);
  const [declineFeedback, setDeclineFeedback] = createSignal("");

  const sendMessage = () => {
    const message = newMessage().trim();
    if (message) {
      postInlayChat.mutate(
        {
          inlay_id: props.inlay().id,
          sender_type: "customer",
          message,
        },
        {
          onSettled() {
            queryClient.invalidateQueries({ queryKey: ["inlay-chat"] });
          },
        },
      );

      setNewMessage("");
    }
  };

  const handleKeyPress = (e: KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const handleDeclineClick = () => {
    setIsDeclineFeedback(true);
  };

  const handleDeclineSubmit = () => {
    const feedback = declineFeedback().trim();
    if (feedback) {
      // const newMsg = {
      //   id: messages().length + 1,
      //   sender: "client",
      //   type: "message",
      //   text: `Proof declined: ${feedback}`,
      //   time: new Date(),
      // };
      // setMessages([...messages(), newMsg]);
      setDeclineFeedback("");
      setIsDeclineFeedback(false);
    }
  };

  const handleDeclineCancel = () => {
    setDeclineFeedback("");
    setIsDeclineFeedback(false);
  };
  return (
    <div class="border rounded-xl py-4 w-full flex flex-col min-h-0">
      {JSON.stringify(props.inlay)}
      <div class="flex flex-col h-full min-h-[400px] max-h-[calc(100vh-400px)]">
        <div class="border-b pb-3 mb-4 px-4">
          <h3 class="text-lg font-semibold text-gray-900">Proof Approval</h3>
          <p class="text-sm text-gray-500">
            Work with us to get your inlay design perfect before placing your
            order!
          </p>
        </div>
        <Show when={query.isSuccess} fallback={<div>loading...</div>}>
          <div class="flex flex-col-reverse overflow-y-auto scroll-smooth">
            <div class="flex-1 space-y-4 mb-4 px-4">
              <Index each={query.data!}>
                {(chat) => (
                  <div
                    class={cn(
                      "flex",
                      chat().sender_type === "customer"
                        ? "justify-end"
                        : "justify-start",
                    )}
                  >
                    <div
                      class={cn(
                        "max-w-xs lg:max-w-md px-4 py-2 rounded-lg",
                        chat().sender_type === "customer"
                          ? "bg-primary text-white"
                          : "bg-gray-100 text-gray-900",
                      )}
                    >
                      <Show
                        when={
                          false
                          // Boolean(chat().img)
                        }
                      >
                        {/* <img src={chat().img} class="py-2" /> */}
                        Proof Image
                      </Show>
                      <p class="text-sm">{chat().message}</p>
                      <Show
                        when={
                          false
                          // Boolean(chat().type === "proof") &&
                          // index === messages().length - 1
                        }
                      >
                        <Show
                          when={!isDeclineFeedback()}
                          fallback={
                            <div class="space-y-2 my-2">
                              <TextFieldRoot>
                                <TextField
                                  value={declineFeedback()}
                                  onInput={(e) =>
                                    setDeclineFeedback(e.currentTarget.value)
                                  }
                                  placeholder="What would you like changed?"
                                  class="w-full"
                                />
                              </TextFieldRoot>
                              <div class="flex gap-2">
                                <Button
                                  variant="outline"
                                  onClick={handleDeclineCancel}
                                  class="flex-1"
                                >
                                  Cancel
                                </Button>
                                <Button
                                  onClick={handleDeclineSubmit}
                                  disabled={!declineFeedback().trim()}
                                  class="flex-1"
                                >
                                  Decline w/ Feedback
                                </Button>
                              </div>
                            </div>
                          }
                        >
                          <div class="flex gap-4 my-2">
                            <Button
                              variant="outline"
                              class="w-full"
                              onClick={handleDeclineClick}
                            >
                              Decline
                            </Button>
                            <Button class="w-full">Approve</Button>
                          </div>
                        </Show>
                      </Show>
                      <Show
                        when={
                          false
                          // Boolean(chat().type === "proof") &&
                          // index !== messages().length - 1
                        }
                      >
                        <div class="flex justify-center mt-2">
                          <p class="text-sm text-gray-500">Declined</p>
                        </div>
                      </Show>
                      <p
                        class={cn(
                          "text-xs mt-1",
                          chat().sender_type === "customer"
                            ? "text-primary-100"
                            : "text-gray-500",
                        )}
                      >
                        {new Date(chat().created_at).toLocaleTimeString([], {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </p>
                    </div>
                  </div>
                )}
              </Index>
              <Show
                when={
                  query.data![query.data!.length - 1]?.sender_type ===
                  "customer"
                }
              >
                <div class="flex justify-center">
                  <p class="text-sm text-gray-500 mt-2">
                    Awaiting response from GlassAct Studios....
                  </p>
                </div>
              </Show>
            </div>
          </div>
        </Show>
        <div class="border-t pt-4 px-4">
          <div class="flex gap-2">
            <TextFieldRoot class="w-full">
              <TextField
                value={newMessage()}
                onInput={(e) => setNewMessage(e.currentTarget.value)}
                onKeyPress={handleKeyPress}
                placeholder="Type your message..."
              />
            </TextFieldRoot>
            <Button onClick={sendMessage} disabled={!newMessage().trim()}>
              Send
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default InlayChatbox;
