import type { Component } from "solid-js";
import {
  TextFieldRoot,
  TextFieldLabel,
  TextField,
  Tabs,
  TabsContent,
  TabsIndicator,
  TabsList,
  TabsTrigger,
  TextArea,
  Button,
  Breadcrumb,
  textfieldLabel,
} from "@glassact/ui";
import { IoClose } from "solid-icons/io";
import { createForm } from "@tanstack/solid-form";

const AddInlay: Component = () => {
  const form = createForm(() => ({
    defaultValues: {
      catalog_number: "",
      description: "",
    },
    onSubmit: async ({ value }) => {
      // Do something with form data
      console.log(value);
    },
  }));
  return (
    <div>
      <Breadcrumb
        crumbs={[
          { title: "Projects", href: "/projects" },
          { title: "Create Project", href: "/projects/create-project" },
          { title: "Add Inlay", href: "/projects/create-project/add-inlay" },
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
                form.handleSubmit();
              }}
            >
              <div class="flex gap-8 flex-col">
                <form.Field
                  name="catalog_number"
                  children={(field) => (
                    <TextFieldRoot class="w-full">
                      <TextFieldLabel>Catalog Number</TextFieldLabel>
                      <TextField
                        placeholder="X-XXX-0000"
                        name={field().name}
                        value={field().state.value}
                        onBlur={field().handleBlur}
                        onInput={(e) => field().handleChange(e.target.value)}
                      />
                    </TextFieldRoot>
                  )}
                />
                <form.Field
                  name="description"
                  children={(field) => (
                    <TextFieldRoot class="w-full">
                      <TextFieldLabel>
                        Describe any modifications to the design (colors, size,
                        ect...)
                      </TextFieldLabel>
                      <TextArea
                        placeholder="Type your message here."
                        name={field().name}
                        value={field().state.value}
                        onBlur={field().handleBlur}
                        onInput={(e) => field().handleChange(e.target.value)}
                      />
                    </TextFieldRoot>
                  )}
                />
                <Button class="mx-auto" type="submit">
                  Add to Order
                </Button>
              </div>
            </form>
          </TabsContent>
          <TabsContent value="custom">
            <div class="flex gap-8 flex-col">
              <TextFieldRoot class="w-full">
                <TextFieldLabel>Project Name</TextFieldLabel>
                <TextField placeholder="Codename: platypus" />
              </TextFieldRoot>
              <TextFieldRoot class="w-full">
                <TextFieldLabel>
                  Describe what the design will be
                </TextFieldLabel>
                <TextArea placeholder="Type your message here." />
              </TextFieldRoot>
              <TextFieldRoot class="w-full">
                <TextFieldLabel>
                  Upload any reference images or designs you have
                </TextFieldLabel>
                <TextArea placeholder="Upload...." />
              </TextFieldRoot>
              <div>
                <span class={textfieldLabel()}>
                  What are the desired dimentions of the finished peice
                </span>
                <div class="flex items-center gap-4">
                  <TextFieldRoot class="w-full">
                    <TextField placeholder="Width (in)" />
                  </TextFieldRoot>
                  <IoClose />
                  <TextFieldRoot class="w-full">
                    <TextField placeholder="Height (in)" />
                  </TextFieldRoot>
                </div>
              </div>
              <Button class="mx-auto">Add to Order</Button>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
};

export default AddInlay;
