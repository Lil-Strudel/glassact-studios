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
} from "@glassact/ui";
import { IoClose } from "solid-icons/io";

const AddItem: Component = () => {
  return (
    <div>
      <Tabs defaultValue="custom">
        <div class="max-w-[400px] mx-auto">
          <TabsList>
            <TabsTrigger value="catalog">Catalog</TabsTrigger>
            <TabsTrigger value="custom">Custom</TabsTrigger>
            <TabsIndicator />
          </TabsList>
        </div>
        <div class="mx-auto max-w-2xl p-4 flex flex-col sm:px-6 lg:px-0">
          <TabsContent value="catalog">
            <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
              Catalog
            </h1>
            <div class="flex gap-8 flex-col">
              <TextFieldRoot class="w-full">
                <TextFieldLabel>Catalog Number</TextFieldLabel>
                <TextField placeholder="X-XXX-0000" />
              </TextFieldRoot>

              <TextFieldRoot class="w-full">
                <TextFieldLabel>
                  Describe any modifications to the design (colors, size,
                  ect...)
                </TextFieldLabel>
                <TextArea placeholder="Type your message here." />
              </TextFieldRoot>
              <Button class="mx-auto">Add to Order</Button>
            </div>
          </TabsContent>
          <TabsContent value="custom">
            <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
              Custom
            </h1>
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
              <TextFieldRoot class="w-full">
                <TextFieldLabel>
                  What are the desired dimentions of the finished peice
                </TextFieldLabel>
                <div class="flex items-center gap-4">
                  <TextField placeholder="Width (in)" />

                  <IoClose />
                  <TextField placeholder="Height (in)" />
                </div>
              </TextFieldRoot>
              <Button class="mx-auto">Add to Order</Button>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
};

export default AddItem;
