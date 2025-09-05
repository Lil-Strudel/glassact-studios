import { createFileRoute, Link, useNavigate } from "@tanstack/solid-router";
import {
  Tabs,
  TabsContent,
  TabsIndicator,
  TabsList,
  TabsTrigger,
  Button,
  Breadcrumb,
  Form,
  textfieldLabel,
} from "@glassact/ui";
import { IoClose } from "solid-icons/io";
import { createForm } from "@tanstack/solid-form";
import { Inlay, POST } from "@glassact/data";
import { z } from "zod";
import { zodStringNumber } from "../../utils/zod-string-number";
import { useProjectFormContext } from "./projects_.create-project";

export const Route = createFileRoute(
  "/_appLayout/projects_/create-project/add-inlay",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const form = useProjectFormContext();
  const navigate = useNavigate();

  function addInlay(inlay: POST<Inlay>) {
    form.setFieldValue("inlays", (oldInlays) => {
      oldInlays.push(inlay);
      return oldInlays;
    });
  }

  const catalogForm = createForm(() => ({
    defaultValues: {
      catalog_number: "",
      description: "",
    },
    validators: {
      onChange: z.object({
        catalog_number: z.string().min(1),
        description: z.string().min(1),
      }),
    },
    onSubmit: async ({ value }) => {
      addInlay({
        project_id: -1,
        preview_url: "https://placehold.co/400",
        name: value.catalog_number,
        price_group: 2,
        type: "catalog",
        catalog_info: {
          inlay_id: -1,
          catalog_item_id: 123,
        },
      });
      navigate({ to: "/projects/create-project" });
    },
  }));

  const customForm = createForm(() => ({
    defaultValues: {
      project_name: "",
      description: "",
      images: [{ url: "" }],
      width: "",
      height: "",
    },
    validators: {
      onChange: z.object({
        project_name: z.string().min(1),
        description: z.string().min(1),
        images: z.array(z.object({ url: z.string() })),
        width: z
          .string()
          .min(1)
          .refine(...zodStringNumber),
        height: z
          .string()
          .min(1)
          .refine(...zodStringNumber),
      }),
    },
    onSubmit: async ({ value }) => {
      addInlay({
        project_id: -1,
        preview_url: "https://placehold.co/400",
        name: value.project_name,
        price_group: 2,
        type: "custom",
        custom_info: {
          inlay_id: -1,
          description: value.description,
          width: Number(value.width),
          height: Number(value.height),
          images: value.images,
        },
      });
      navigate({ to: "/projects/create-project" });
    },
  }));

  return (
    <div>
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          { title: "Create Project", to: "/projects/create-project" },
          { title: "Add Inlay", to: "/projects/create-project/add-inlay" },
        ]}
      />
      <Tabs defaultValue="catalog">
        <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
          Add Item
        </h1>
        <div class="max-w-[400px] mx-auto mt-4">
          <TabsList>
            <TabsTrigger value="catalog">Catalog</TabsTrigger>
            <TabsTrigger value="custom">Custom</TabsTrigger>
            <TabsIndicator />
          </TabsList>
        </div>
        <div class="mx-auto max-w-2xl p-4 flex flex-col sm:px-6 lg:px-0">
          <TabsContent value="catalog">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                catalogForm.handleSubmit();
              }}
            >
              <div class="flex gap-8 flex-col">
                <catalogForm.Field
                  name="catalog_number"
                  children={(field) => (
                    <Form.TextField
                      field={field}
                      label="Catalog Number"
                      placeholder="X-XXX-0000"
                    />
                  )}
                />
                <catalogForm.Field
                  name="description"
                  children={(field) => (
                    <Form.TextArea
                      field={field}
                      label="Describe any modifications to the design (colors, size, ect...)"
                      placeholder="Type your message here."
                    />
                  )}
                />
                <div class="mx-auto flex gap-4">
                  <Button
                    variant="outline"
                    as={Link}
                    to="/projects/create-project"
                  >
                    Cancel
                  </Button>
                  <Button type="submit">Add to Order</Button>
                </div>
              </div>
            </form>
          </TabsContent>
          <TabsContent value="custom">
            <form
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                customForm.handleSubmit();
              }}
            >
              <div class="flex gap-8 flex-col">
                <customForm.Field
                  name="project_name"
                  children={(field) => (
                    <Form.TextField
                      field={field}
                      label="Project Name"
                      placeholder="Codename: platypus"
                    />
                  )}
                />
                <customForm.Field
                  name="description"
                  children={(field) => (
                    <Form.TextArea
                      field={field}
                      label="Describe what the design will be"
                      placeholder="Type your message here."
                    />
                  )}
                />
                <customForm.Field
                  name="images[0].url"
                  children={(field) => (
                    <Form.TextArea
                      field={field}
                      label="Upload any reference images or designs you have"
                      placeholder="Upload...."
                    />
                  )}
                />
                <div>
                  <span class={textfieldLabel()}>
                    What are the desired dimentions of the finished peice
                  </span>
                  <div class="flex items-center gap-4">
                    <customForm.Field
                      name="width"
                      children={(field) => (
                        <Form.TextField
                          field={field}
                          placeholder="Width (in)"
                        />
                      )}
                    />
                    <IoClose />
                    <customForm.Field
                      name="height"
                      children={(field) => (
                        <Form.TextField
                          field={field}
                          placeholder="Height (in)"
                        />
                      )}
                    />
                  </div>
                </div>
                <div class="mx-auto flex gap-4">
                  <Button
                    variant="outline"
                    as={Link}
                    to="/projects/create-project"
                  >
                    Cancel
                  </Button>
                  <Button type="submit">Add to Order</Button>
                </div>
              </div>
            </form>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}
