import { createFileRoute, Link } from "@tanstack/solid-router";
import { Breadcrumb, Button, cn, TextField, TextFieldRoot } from "@glassact/ui";
import { createSignal, Index, Show } from "solid-js";
import { IoCheckmarkCircleOutline } from "solid-icons/io";
import { useQuery } from "@tanstack/solid-query";
import { getProjectOpts } from "../../queries/project";
import InlayChatbox from "../../components/inlay-chatbox";

export const Route = createFileRoute("/_app/projects_/$id")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const [selectedInlayIndex, setSelectedInlayIndex] = createSignal(0);

  const query = useQuery(
    getProjectOpts(params().id, { expand: { inlays: true } }),
  );

  const steps = [
    { name: "Proof Creation", to: "#", status: "complete" },
    { name: "Proof Approval", to: "#", status: "current" },
    { name: "Order Placement", to: "#", status: "upcoming" },
    { name: "Material Prep", to: "#", status: "upcoming" },
    { name: "Cutting", to: "#", status: "upcoming" },
    { name: "Fire Polishing", to: "#", status: "upcoming" },
    { name: "Packaging", to: "#", status: "upcoming" },
    { name: "Shipping", to: "#", status: "upcoming" },
    { name: "Delivered", to: "#", status: "upcoming" },
  ];

  const selectedInlay = () => query.data!.inlays[selectedInlayIndex()];

  return (
    <Show when={query.isSuccess} fallback={<div>Loading</div>}>
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          { title: query.data!.name, to: `/projects/${query.data!.uuid}` },
        ]}
      />
      <div class="relative border-b border-gray-200 pb-5 sm:pb-0">
        <div class="md:flex md:items-center md:justify-between">
          <h1 class="text-2xl font-bold text-gray-900">{query.data!.name}</h1>
          <div class="mt-3 flex gap-4 md:absolute md:right-0 md:top-3 md:mt-0">
            <Button variant="outline">Cancel Project</Button>
            <Button disabled>Place Order</Button>
          </div>
        </div>
        <div class="mt-4">
          <div>
            <nav class="-mb-px flex space-x-8">
              <Index each={query.data!.inlays}>
                {(item, index) => (
                  <div
                    onClick={() => setSelectedInlayIndex(index)}
                    class={cn(
                      "cursor-pointer whitespace-nowrap border-b-2 border-primary px-1 pb-2 text-sm font-medium text-primary",
                      index === selectedInlayIndex()
                        ? "border-primary text-primary"
                        : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700",
                    )}
                  >
                    {item().name}
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
                      <Link to={step().to} class="group">
                        <span class="flex items-start">
                          <span class="relative flex size-5 shrink-0 items-center justify-center">
                            <IoCheckmarkCircleOutline class="size-full text-primary group-hover:text-primary" />
                          </span>
                          <span class="ml-3 text-sm font-medium text-gray-500 group-hover:text-gray-900">
                            {step().name}
                          </span>
                        </span>
                      </Link>
                    ) : step().status === "current" ? (
                      <Link
                        to={step().to}
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
                      </Link>
                    ) : (
                      <Link to={step().to} class="grooup">
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
                      </Link>
                    )}
                  </li>
                )}
              </Index>
            </ol>
          </nav>
        </div>
        <Show when={selectedInlay()}>
          <InlayChatbox inlay={selectedInlay} />
        </Show>
      </div>
    </Show>
  );
}
