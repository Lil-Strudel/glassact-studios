import { createFileRoute, Link } from "@tanstack/solid-router";
import { Breadcrumb, Button, Form } from "@glassact/ui";
import { Index, Show } from "solid-js";
import { formatMoney } from "../../utils/format-money";
import { useProjectFormContext } from "./projects_.create-project";

export const Route = createFileRoute("/_app/projects_/create-project/")({
  component: RouteComponent,
});

function RouteComponent() {
  const form = useProjectFormContext();

  return (
    <div>
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          { title: "Create Project", to: "/projects/create-project" },
        ]}
      />
      <div class="mx-auto max-w-2xl px-4 sm:px-6 lg:px-0">
        <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
          Create a New Project
        </h1>

        <div class="mt-10">
          <form.Field
            name="name"
            children={(field) => (
              <Form.TextField
                field={field}
                label="Who is this project for?"
                placeholder="John Smith"
              />
            )}
          />
        </div>

        <div>
          <form.Field name="inlays" mode="array">
            {(field) => (
              <>
                <Show when={field().state.value.length === 0}>
                  <section class="grid place-items-center mt-10 gap-4 border rounded-md p-4 drop-shadow-sm">
                    <p>
                      It looks like you have no inlays added to the project.
                    </p>
                    <Button as={Link} to="/projects/create-project/add-inlay">
                      Add an Inlay
                    </Button>
                    <p>to get started!</p>
                  </section>
                  <form.Field
                    name="inlays"
                    children={(field) => <Form.ErrorLabel field={field} />}
                  />
                </Show>
                <Show when={field().state.value.length > 0}>
                  <section class="mt-10">
                    <ul
                      role="list"
                      class="divide-y divide-gray-200 border-b border-t border-gray-200"
                    >
                      <Index each={field().state.value}>
                        {(inlay, index) => (
                          <li class="flex py-6">
                            <div class="shrink-0">
                              <img
                                src={inlay().preview_url}
                                alt={inlay().name}
                                class="size-24 rounded-md object-cover sm:size-32"
                              />
                            </div>

                            <div class="ml-4 flex flex-1 flex-col sm:ml-6">
                              <div>
                                <div class="flex justify-between">
                                  <h4 class="text-sm">
                                    <a
                                      href="#"
                                      class="font-medium text-gray-700 hover:text-gray-800"
                                    >
                                      {inlay().name}
                                    </a>
                                  </h4>
                                  <p class="ml-4 text-sm font-medium text-gray-900">
                                    ~{formatMoney(inlay().price_group * 24)}
                                  </p>
                                </div>
                                <p class="mt-1 text-sm text-gray-500">
                                  {inlay().type[0].toUpperCase()}
                                  {inlay().type.slice(1)}
                                </p>
                                <p class="mt-1 text-sm text-gray-500">
                                  PG-{inlay().price_group}
                                </p>
                              </div>

                              <div class="mt-4 flex flex-1 items-end justify-between">
                                <p class="flex items-center space-x-2 text-sm text-gray-700">
                                  <svg
                                    class="size-5 shrink-0 text-green-500"
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
                                  <span>Lorem Ipsum</span>
                                </p>
                                <div class="ml-4">
                                  <Button
                                    variant="text"
                                    class="p-0"
                                    onClick={() => field().removeValue(index)}
                                  >
                                    Remove
                                  </Button>
                                </div>
                              </div>
                            </div>
                          </li>
                        )}
                      </Index>
                    </ul>
                  </section>
                  <section class="grid place-items-center mt-10">
                    <Button
                      as={Link}
                      to="/projects/create-project/add-inlay"
                      variant="outline"
                    >
                      Add Additional Inlay
                    </Button>
                  </section>
                </Show>
              </>
            )}
          </form.Field>

          <section class="mt-10">
            <div>
              <dl class="space-y-4">
                <div class="flex items-center justify-between">
                  <dt class="text-base font-medium text-gray-900">
                    Estimated Total
                  </dt>
                  <form.Subscribe
                    selector={(state) => ({
                      inlays: state.values.inlays,
                    })}
                    children={(state) => (
                      <dd class="ml-4 text-base font-medium text-gray-900">
                        ~
                        {formatMoney(
                          state().inlays.reduce(
                            (acc, cur) => cur.price_group * 24 + acc,
                            0,
                          ),
                        )}
                      </dd>
                    )}
                  />
                </div>
              </dl>
              <p class="mt-1 text-sm text-gray-500">
                Shipping and taxes will be calculated at time of shipment.
              </p>
            </div>

            <div class="mt-10 flex flex-col gap-2 items-center">
              <Button class="w-full" type="submit">
                Create Project
              </Button>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
