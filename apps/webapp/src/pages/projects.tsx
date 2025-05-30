import { GET, Project } from "@glassact/data";
import { Button, Breadcrumb } from "@glassact/ui";
import { IoAddCircleOutline, IoCheckmarkCircleOutline } from "solid-icons/io";
import { type Component, Index } from "solid-js";

const Projects: Component = () => {
  const projects: GET<Project>[] = [
    {
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
    },
  ];
  return (
    <div>
      <Breadcrumb crumbs={[{ title: "Projects", href: "/projects" }]} />
      <div>
        <Button as="a" href="/projects/create-project">
          Create New Project
          <IoAddCircleOutline size={20} class="ml-2" />
        </Button>
      </div>
      <div class="flex flex-col gap-16 mt-4">
        <div>
          <div>
            <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
              Needs Action
            </h1>
            <p class="mt-2 text-sm text-gray-500">
              These projects have proofs that are awaiting your approval or
              invoices waiting to be paid.
            </p>
          </div>

          <div class="mt-4">
            <div class="space-y-20">
              <Index each={projects}>
                {(item) => (
                  <div class="border rounded-xl p-4">
                    <div class="flex items-center justify-between">
                      <span class="text-2xl font-bold">{item().name}</span>
                      <Button as="a" href={`/projects/${item().uuid}`}>
                        View Proofs
                      </Button>
                    </div>

                    <table class="mt-4 w-full text-gray-500">
                      <thead class="text-left text-sm text-gray-500">
                        <tr>
                          <th scope="col" class="py-3">
                            Inlay
                          </th>
                          <th scope="col" class="py-3 text-right">
                            Proof Status
                          </th>
                        </tr>
                      </thead>
                      <tbody class="divide-y divide-gray-200 border-y border-gray-200 text-sm">
                        <tr>
                          <td class="py-4">
                            <div class="flex items-center">
                              <img
                                src="https://tailwindcss.com/plus-assets/img/ecommerce-images/order-history-page-02-product-01.jpg"
                                alt="Detail of mechanical pencil tip with machined black steel shaft and chrome lead tip."
                                class="mr-6 size-16 rounded object-cover"
                              />
                              <div class="font-medium text-gray-900">
                                1234-78-A21
                              </div>
                            </div>
                          </td>
                          <td class="text-right">Proof Awaiting Approval</td>
                        </tr>
                        <tr>
                          <td class="py-4">
                            <div class="flex items-center">
                              <img
                                src="https://tailwindcss.com/plus-assets/img/ecommerce-images/order-history-page-02-product-01.jpg"
                                alt="Detail of mechanical pencil tip with machined black steel shaft and chrome lead tip."
                                class="mr-6 size-16 rounded object-cover"
                              />
                              <div class="font-medium text-gray-900">
                                BIR-203-152
                              </div>
                            </div>
                          </td>
                          <td>
                            <div class="flex justify-end">
                              Approved
                              <IoCheckmarkCircleOutline
                                size={20}
                                class="ml-2"
                              />
                            </div>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                )}
              </Index>
            </div>
          </div>
        </div>

        <div>
          <div>
            <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
              Active Projects
            </h1>
            <p class="mt-2 text-sm text-gray-500">
              These are projects currently in the pipeline.
            </p>
          </div>
        </div>

        <div>
          <div>
            <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
              Completed Projects
            </h1>
            <p class="mt-2 text-sm text-gray-500">
              These are projects that have been created and shipped.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Projects;
