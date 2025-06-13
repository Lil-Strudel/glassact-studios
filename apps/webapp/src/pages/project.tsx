import { GET, Project } from "@glassact/data";
import { Breadcrumb, Button, cn } from "@glassact/ui";
import { createSignal, Component, Index } from "solid-js";
import { IoCheckmarkCircleOutline } from "solid-icons/io";

const ProjectPage: Component = () => {
  const [selectedInlay, setSelectedInlay] = createSignal(0);

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
    { name: "Shipped", href: "#", status: "upcoming" },
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
        <div class="border rounded-xl p-4 w-full">hello</div>
      </div>
    </div>
  );
};

export default ProjectPage;
