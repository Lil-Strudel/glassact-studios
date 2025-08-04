import { createFileRoute } from "@tanstack/solid-router";
import { GET, Project } from "@glassact/data";
import { Button, Breadcrumb } from "@glassact/ui";
import { IoAddCircleOutline, IoCheckmarkCircleOutline } from "solid-icons/io";
import { Index } from "solid-js";

export const Route = createFileRoute("/_appLayout/projects")({
  component: RouteComponent,
});

function RouteComponent() {
  const newProjects: GET<Project>[] = [
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

  const invoiceProjects: GET<Project>[] = [
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
  const activeProjects: GET<Project>[] = [
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
  const completedProjects: GET<Project>[] = [
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
            <div class="space-y-4">
              <Index each={newProjects}>
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
              <Index each={invoiceProjects}>
                {(item) => (
                  <div class="border rounded-xl p-4">
                    <div class="flex items-center justify-between">
                      <span class="text-2xl font-bold">{item().name}</span>
                      <Button as="a" href={`/projects/${item().uuid}`}>
                        View Invoice
                      </Button>
                    </div>

                    <table class="mt-4 w-full text-gray-500">
                      <thead class="text-left text-sm text-gray-500">
                        <tr>
                          <th scope="col" class="py-3">
                            Inlay
                          </th>
                          <th scope="col" class="py-3 text-right">
                            Status
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
                          <td class="text-right">Shipped</td>
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
                          <td class="text-right">Shipped</td>
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
          <div class="mt-4">
            <div class="space-y-4">
              <Index each={activeProjects}>
                {(item) => (
                  <div class="border rounded-xl p-4">
                    <div class="flex items-center justify-between">
                      <span class="text-2xl font-bold">{item().name}</span>
                      <Button as="a" href={`/projects/${item().uuid}`}>
                        View Project
                      </Button>
                    </div>

                    <table class="mt-4 w-full text-gray-500">
                      <thead class="text-left text-sm text-gray-500">
                        <tr>
                          <th scope="col" class="py-3">
                            Inlay
                          </th>
                          <th scope="col" class="py-3">
                            Progress
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
                          <td>
                            <ol role="list" class="flex items-center">
                              <li class="relative pr-8 sm:pr-20">
                                <div class="absolute inset-0 flex items-center">
                                  <div class="h-0.5 w-full bg-primary"></div>
                                </div>
                                <a
                                  href="#"
                                  class="relative flex size-8 items-center justify-center rounded-full bg-primary"
                                >
                                  <svg
                                    class="size-5 text-white"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                    data-slot="icon"
                                  >
                                    <path
                                      fill-rule="evenodd"
                                      d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z"
                                      clip-rule="evenodd"
                                    />
                                  </svg>
                                </a>
                              </li>
                              <li class="relative pr-8 sm:pr-20">
                                <div class="absolute inset-0 flex items-center">
                                  <div class="h-0.5 w-full bg-primary"></div>
                                </div>
                                <a
                                  href="#"
                                  class="relative flex size-8 items-center justify-center rounded-full bg-primary"
                                >
                                  <svg
                                    class="size-5 text-white"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                    data-slot="icon"
                                  >
                                    <path
                                      fill-rule="evenodd"
                                      d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z"
                                      clip-rule="evenodd"
                                    />
                                  </svg>
                                </a>
                              </li>
                              <li class="relative pr-8 sm:pr-20">
                                <div class="absolute inset-0 flex items-center">
                                  <div class="h-0.5 w-full bg-gray-200"></div>
                                </div>
                                <a
                                  href="#"
                                  class="relative flex size-8 items-center justify-center rounded-full border-2 border-primary bg-white"
                                >
                                  <span class="size-2.5 rounded-full bg-primary"></span>
                                </a>
                              </li>
                              <li class="relative pr-8 sm:pr-20">
                                <div class="absolute inset-0 flex items-center">
                                  <div class="h-0.5 w-full bg-gray-200"></div>
                                </div>
                                <a
                                  href="#"
                                  class="group relative flex size-8 items-center justify-center rounded-full border-2 border-gray-300 bg-white hover:border-gray-400"
                                >
                                  <span class="size-2.5 rounded-full bg-transparent group-hover:bg-gray-300"></span>
                                </a>
                              </li>
                              <li class="relative">
                                <div class="absolute inset-0 flex items-center">
                                  <div class="h-0.5 w-full bg-gray-200"></div>
                                </div>
                                <a
                                  href="#"
                                  class="group relative flex size-8 items-center justify-center rounded-full border-2 border-gray-300 bg-white hover:border-gray-400"
                                >
                                  <span class="size-2.5 rounded-full bg-transparent group-hover:bg-gray-300"></span>
                                </a>
                              </li>
                            </ol>
                          </td>
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
                          <td class="">Shipped</td>
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
              Completed Projects
            </h1>
            <p class="mt-2 text-sm text-gray-500">
              These are projects that have been created and shipped.
            </p>
          </div>
          <div class="mt-4">
            <div class="space-y-4">
              <Index each={completedProjects}>
                {(item) => (
                  <div class="border rounded-xl p-4">
                    <div class="flex items-center justify-between">
                      <span class="text-2xl font-bold">{item().name}</span>
                      <Button as="a" href={`/projects/${item().uuid}`}>
                        View Receipt
                      </Button>
                    </div>

                    <table class="mt-4 w-full text-gray-500">
                      <thead class="text-left text-sm text-gray-500">
                        <tr>
                          <th scope="col" class="py-3">
                            Inlay
                          </th>
                          <th scope="col" class="py-3 text-right">
                            Status
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
                          <td class="text-right">Delivered</td>
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
                          <td class="text-right">Delivered</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                )}
              </Index>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
