import { GET, Project } from "@glassact/data";
import { Breadcrumb, Button, cn, TextField, TextFieldRoot } from "@glassact/ui";
import {
  createSignal,
  Component,
  Index,
  Show,
  createEffect,
  onMount,
} from "solid-js";
import { IoCheckmarkCircleOutline } from "solid-icons/io";

const ProjectPage: Component = () => {
  const [selectedInlay, setSelectedInlay] = createSignal(0);
  const [messages, setMessages] = createSignal([
    {
      id: 1,
      sender: "client",
      type: "order",
      text: "Catalog Item: 1234-78-A21 Please switch out the blue with light blue.",
      time: new Date("2025-06-06T17:30:45"),
    },
    {
      id: 2,
      sender: "glassact",
      type: "proof",
      text: "Here is the proof!",
      img: "https://placehold.co/600x400",
      time: new Date("2025-06-06T17:30:45"),
    },
    {
      id: 3,
      sender: "client",
      type: "message",
      text: "The dimensions are wrong. Can you make it smaller?",
      time: new Date("2025-06-06T17:30:45"),
    },
    {
      id: 4,
      sender: "glassact",
      type: "message",
      text: "What are the exact dimensions you want it to be?",
      time: new Date("2025-06-06T17:30:45"),
    },
    {
      id: 5,
      sender: "client",
      type: "message",
      text: "2.5 x 3.5in",
      time: new Date("2025-06-06T17:30:45"),
    },
    {
      id: 6,
      sender: "glassact",
      type: "proof",
      text: "How does this look?",
      img: "https://placehold.co/600x400",
      time: new Date("2025-06-06T17:30:45"),
    },
  ]);

  const [newMessage, setNewMessage] = createSignal("");

  let chatBoxRef: HTMLDivElement | undefined;
  let messagesContainerRef: HTMLDivElement | undefined;

  const scrollToBottom = () => {
    if (messagesContainerRef) {
      messagesContainerRef.scrollTop = messagesContainerRef.scrollHeight;
    }
  };

  // Auto-scroll to bottom when messages change
  createEffect(() => {
    messages(); // Track messages signal
    setTimeout(scrollToBottom, 0); // Use setTimeout to ensure DOM is updated
  });

  // Scroll to bottom on mount
  onMount(() => {
    scrollToBottom();
  });

  const sendMessage = () => {
    const message = newMessage().trim();
    if (message) {
      const newMsg = {
        id: messages().length + 1,
        sender: "client",
        type: "message",
        text: message,
        time: new Date(),
      };
      setMessages([...messages(), newMsg]);
      setNewMessage("");
    }
  };

  const handleKeyPress = (e: KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const project: GET<Project> = {
    id: 123,
    uuid: "1234",
    name: "John Doe",
    status: "1234",
    approved: false,
    dealership_id: 123,
    shipment_id: 123,
    created_at: "qw34",
    updated_at: "1234",
    version: 1,
  };

  const inlays = ["1234-78-A21", "BIR-203-152"];

  const steps = [
    { name: "Proof Creation", href: "#", status: "complete" },
    { name: "Proof Approval", href: "#", status: "current" },
    { name: "Order Placement", href: "#", status: "upcoming" },
    { name: "Material Prep", href: "#", status: "upcoming" },
    { name: "Cutting", href: "#", status: "upcoming" },
    { name: "Fire Polishing", href: "#", status: "upcoming" },
    { name: "Packaging", href: "#", status: "upcoming" },
    { name: "Shipping", href: "#", status: "upcoming" },
    { name: "Delivered", href: "#", status: "upcoming" },
  ];

  return (
    <div>
      <Breadcrumb
        crumbs={[
          { title: "Projects", href: "/projects" },
          { title: project.name, href: `/projects/${project.uuid}` },
        ]}
      />
      <div class="relative border-b border-gray-200 pb-5 sm:pb-0">
        <div class="md:flex md:items-center md:justify-between">
          <h1 class="text-2xl font-bold text-gray-900">{project.name}</h1>
          <div class="mt-3 flex gap-4 md:absolute md:right-0 md:top-3 md:mt-0">
            <Button variant="outline">Cancel Project</Button>
            <Button disabled>Place Order</Button>
          </div>
        </div>
        <div class="mt-4">
          <div>
            <nav class="-mb-px flex space-x-8">
              <Index each={inlays}>
                {(item, index) => (
                  <div
                    onClick={() => setSelectedInlay(index)}
                    class={cn(
                      "cursor-pointer whitespace-nowrap border-b-2 border-primary px-1 pb-2 text-sm font-medium text-primary",
                      index === selectedInlay()
                        ? "border-primary text-primary"
                        : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700",
                    )}
                  >
                    {item()}
                  </div>
                )}
              </Index>
            </nav>
          </div>
        </div>
      </div>
      <div class="mt-8 flex gap-4">
        <div>
          <nav class="flex">
            <ol role="list" class="space-y-6">
              <Index each={steps}>
                {(step) => (
                  <li>
                    {step().status === "complete" ? (
                      <a href={step().href} class="group">
                        <span class="flex items-start">
                          <span class="relative flex size-5 shrink-0 items-center justify-center">
                            <IoCheckmarkCircleOutline class="size-full text-primary group-hover:text-primary" />
                          </span>
                          <span class="ml-3 text-sm font-medium text-gray-500 group-hover:text-gray-900">
                            {step().name}
                          </span>
                        </span>
                      </a>
                    ) : step().status === "current" ? (
                      <a
                        href={step().href}
                        aria-current="step"
                        class="flex items-start"
                      >
                        <span
                          aria-hidden="true"
                          class="relative flex size-5 shrink-0 items-center justify-center"
                        >
                          <span class="absolute size-4 rounded-full bg-red-100" />
                          <span class="relative block size-2 rounded-full bg-primary" />
                        </span>
                        <span class="ml-3 text-sm font-medium text-primary">
                          {step().name}
                        </span>
                      </a>
                    ) : (
                      <a href={step().href} class="group">
                        <div class="flex items-start">
                          <div
                            aria-hidden="true"
                            class="relative flex size-5 shrink-0 items-center justify-center"
                          >
                            <div class="size-2 rounded-full bg-gray-300 group-hover:bg-gray-400" />
                          </div>
                          <p class="ml-3 text-sm font-medium text-gray-500 group-hover:text-gray-900">
                            {step().name}
                          </p>
                        </div>
                      </a>
                    )}
                  </li>
                )}
              </Index>
            </ol>
          </nav>
        </div>
        <div class="border rounded-xl py-4 w-full flex flex-col min-h-0">
          <div
            ref={chatBoxRef}
            class="flex flex-col h-full min-h-[400px] max-h-[calc(100vh-400px)]"
          >
            <div class="border-b pb-3 mb-4 px-4">
              <h3 class="text-lg font-semibold text-gray-900">
                Proof Approval
              </h3>
              <p class="text-sm text-gray-500">
                Work with us to get your inlay design perfect before placing
                your order!
              </p>
            </div>
            <div
              ref={messagesContainerRef}
              class="flex-1 overflow-y-auto space-y-4 mb-4 px-4 scroll-smooth"
            >
              <Index each={messages()}>
                {(message, index) => (
                  <div
                    class={cn(
                      "flex",
                      message().sender === "client"
                        ? "justify-end"
                        : "justify-start",
                    )}
                  >
                    <div
                      class={cn(
                        "max-w-xs lg:max-w-md px-4 py-2 rounded-lg",
                        message().sender === "client"
                          ? "bg-primary text-white"
                          : "bg-gray-100 text-gray-900",
                      )}
                    >
                      <Show when={Boolean(message().img)}>
                        <img src={message().img} class="py-2" />
                      </Show>
                      <p class="text-sm">{message().text}</p>
                      <Show
                        when={
                          Boolean(message().type === "proof") &&
                          index === messages().length - 1
                        }
                      >
                        <div class="flex gap-4 my-2">
                          <Button variant="outline" class="w-full">
                            Decline
                          </Button>
                          <Button class="w-full">Approve</Button>
                        </div>
                      </Show>
                      <Show
                        when={
                          Boolean(message().type === "proof") &&
                          index !== messages().length - 1
                        }
                      >
                        <div class="flex justify-center mt-2">
                          <p class="text-sm text-gray-500">Declined</p>
                        </div>
                      </Show>
                      <p
                        class={cn(
                          "text-xs mt-1",
                          message().sender === "client"
                            ? "text-primary-100"
                            : "text-gray-500",
                        )}
                      >
                        {message().time.toLocaleTimeString([], {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      </p>
                    </div>
                  </div>
                )}
              </Index>
              <Show
                when={messages()[messages().length - 1].sender === "client"}
              >
                <div class="flex justify-center">
                  <p class="text-sm text-gray-500 mt-2">
                    Awaiting response from GlassAct Studios....
                  </p>
                </div>
              </Show>
            </div>
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
      </div>
    </div>
  );
};

export default ProjectPage;
