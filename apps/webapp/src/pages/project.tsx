import { GET, Project } from "@glassact/data";
import { Breadcrumb, Button, cn } from "@glassact/ui";
import { createSignal, Component, Index } from "solid-js";

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
                      "cursor-pointer whitespace-nowrap border-b-2 border-primary px-1 pb-4 text-sm font-medium text-primary",
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
    </div>
  );
};

export default ProjectPage;
