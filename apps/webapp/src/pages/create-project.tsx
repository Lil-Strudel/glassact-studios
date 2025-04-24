import { Breadcrumb, Button, Form, textfieldLabel } from "@glassact/ui";
import { type Component, For, Show } from "solid-js";
import { formatMoney } from "../utils/format-money";
import { initialAppState, useAppState } from "../providers/app-state";
import { useNavigate } from "@solidjs/router";
import { reconcile } from "solid-js/store";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";

const CreateProject: Component = () => {
  const navigate = useNavigate();
  const [state, setState] = useAppState();

  function setName(value: string) {
    setState("createProject", "name", value);
  }

  const totalPrice = () =>
    state.createProject.inlays.reduce(
      (acc, cur) => cur.price_group * 24 + acc,
      0,
    );

  function removeInlay(idx: number) {
    setState(
      "createProject",
      "inlays",
      state.createProject.inlays.filter((_, index) => index !== idx),
    );
  }

  const form = createForm(() => ({
    defaultValues: {
      name: state.createProject.name,
      inlays: state.createProject.inlays,
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        inlays: z.array(z.any()).min(1),
      }),
    },
    onSubmit: async ({ value }) => {
      resetState();
      navigate("/projects");
    },
  }));

  function resetState() {
    setState("createProject", reconcile(initialAppState.createProject));
    form.reset();
  }

  return (
    <div>
      <Breadcrumb
        crumbs={[
          { title: "Projects", href: "/projects" },
          { title: "Create Project", href: "/projects/create-project" },
        ]}
      />
      <div class="mx-auto max-w-2xl px-4 sm:px-6 lg:px-0">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();

            form.handleSubmit();
          }}
        >
          <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Create a New Project
          </h1>

          <div class="mt-10">
            <form.Field
              name="name"
              listeners={{
                onChange: ({ value }) => setName(value),
              }}
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
            <Show when={state.createProject.inlays.length === 0}>
              <section class="grid place-items-center mt-10 gap-4 border rounded-md p-4 drop-shadow-sm">
                <p>It looks like you have no inlays added to the project.</p>
                <Button as="a" href="/projects/create-project/add-inlay">
                  Add an Inlay
                </Button>
                <p>to get started!</p>
              </section>
              <form.Field
                name="inlays"
                children={(field) => <Form.ErrorLabel field={field} />}
              />
            </Show>
            <Show when={state.createProject.inlays.length > 0}>
              <section class="mt-10">
                <ul
                  role="list"
                  class="divide-y divide-gray-200 border-b border-t border-gray-200"
                >
                  <For each={state.createProject.inlays}>
                    {(inlay, index) => (
                      <li class="flex py-6">
                        <div class="shrink-0">
                          <img
                            src={inlay.preview_url}
                            alt={inlay.name}
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
                                  {inlay.name}
                                </a>
                              </h4>
                              <p class="ml-4 text-sm font-medium text-gray-900">
                                ~{formatMoney(inlay.price_group * 24)}
                              </p>
                            </div>
                            <p class="mt-1 text-sm text-gray-500">
                              {inlay.type[0].toUpperCase()}
                              {inlay.type.slice(1)}
                            </p>
                            <p class="mt-1 text-sm text-gray-500">
                              PG-{inlay.price_group}
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
                                onClick={() => removeInlay(index())}
                              >
                                Remove
                              </Button>
                            </div>
                          </div>
                        </div>
                      </li>
                    )}
                  </For>
                </ul>
              </section>
              <section class="grid place-items-center mt-10">
                <Button
                  as="a"
                  href="/projects/create-project/add-inlay"
                  variant="outline"
                >
                  Add Additional Inlay
                </Button>
              </section>
            </Show>

            <section class="mt-10">
              <div>
                <dl class="space-y-4">
                  <div class="flex items-center justify-between">
                    <dt class="text-base font-medium text-gray-900">
                      Estimated Total
                    </dt>
                    <dd class="ml-4 text-base font-medium text-gray-900">
                      ~{formatMoney(totalPrice())}
                    </dd>
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

                <span>or</span>

                <Button class="w-full" onClick={resetState} variant="text">
                  Start Fresh
                </Button>
              </div>
            </section>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateProject;
