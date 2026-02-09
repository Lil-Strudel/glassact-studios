import { CatalogItem, GET, POST, PATCH } from "@glassact/data";
import { Form, Button, Badge } from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { useQuery } from "@tanstack/solid-query";
import { z } from "zod";
import { For, Show, createSignal, createEffect } from "solid-js";
import { getPriceGroupsOpts } from "../../queries/price-group";
import { getCatalogAllTagsOpts } from "../../queries/catalog-browse";

interface CatalogFormProps {
  defaultValues?: GET<CatalogItem>;
  onSubmit: (
    data: POST<CatalogItem> | PATCH<CatalogItem>,
    tags: string[],
  ) => Promise<void>;
  isLoading?: boolean;
  isEditMode?: boolean;
}

const catalogSchema = z.object({
  catalog_code: z.string().min(1).max(255),
  name: z.string().min(1).max(255),
  description: z.string().max(2000).nullable().optional(),
  category: z.string().min(1).max(255),
  default_width: z.number().positive(),
  default_height: z.number().positive(),
  min_width: z.number().positive(),
  min_height: z.number().positive(),
  default_price_group_id: z.number().positive().int(),
  svg_url: z.string().min(1),
  is_active: z.boolean().default(true),
});

type CatalogSchema = z.infer<typeof catalogSchema>;

export function CatalogForm(props: CatalogFormProps) {
  const [tags, setTags] = createSignal<string[]>([]);
  const [tagInput, setTagInput] = createSignal("");
  const [showTagSuggestions, setShowTagSuggestions] = createSignal(false);

  const priceGroupsQuery = useQuery(getPriceGroupsOpts({ limit: 100 }));
  const tagsQuery = useQuery(getCatalogAllTagsOpts());

  const filteredSuggestions = () => {
    if (!tagInput() || !tagsQuery.data) return [];
    const input = tagInput().toLowerCase();
    const existingTags = new Set(tags());
    return tagsQuery.data
      .filter(
        (tag) => tag.toLowerCase().includes(input) && !existingTags.has(tag),
      )
      .slice(0, 10);
  };

  const form = createForm(() => ({
    defaultValues: props.defaultValues ?? {
      catalog_code: "",
      name: "",
      description: null,
      category: "",
      default_width: 0,
      default_height: 0,
      min_width: 0,
      min_height: 0,
      default_price_group_id: 0,
      svg_url: "",
      is_active: true,
    },
    onSubmit: async ({ value }) => {
      await props.onSubmit(value as any, tags());
    },
  }));

  const validateMinDimensions = (field: "min_width" | "min_height") => {
    const minValue =
      field === "min_width"
        ? form.getFieldValue("min_width")
        : form.getFieldValue("min_height");
    const maxValue =
      field === "min_width"
        ? form.getFieldValue("default_width")
        : form.getFieldValue("default_height");

    if (minValue > maxValue) {
      return `Must be ≤ default ${field === "min_width" ? "width" : "height"}`;
    }
    return undefined;
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
      class="flex flex-col gap-6"
    >
      <form.Field
        name="catalog_code"
        children={(field) => (
          <div class="flex flex-col gap-2">
            <label class="text-sm font-medium text-gray-900">
              Catalog Code
            </label>
            <input
              type="text"
              value={field().state.value}
              onInput={(e) => field().handleChange(e.currentTarget.value)}
              placeholder="e.g., CAT-001"
              disabled={props.isEditMode}
              class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            />
            <Show when={field().state.meta.errors.length > 0}>
              <span class="text-sm text-red-600">
                {field().state.meta.errors[0]}
              </span>
            </Show>
          </div>
        )}
      />

      <form.Field
        name="name"
        children={(field) => (
          <Form.TextField field={field} label="Name" placeholder="Item name" />
        )}
      />

      <form.Field
        name="description"
        children={(field) => (
          <Form.TextArea
            field={field}
            label="Description"
            placeholder="Optional item description"
          />
        )}
      />

      <form.Field
        name="category"
        children={(field) => (
          <Form.TextField
            field={field}
            label="Category"
            placeholder="e.g., Windows, Doors"
          />
        )}
      />

       <div class="border-t pt-4">
         <h3 class="text-sm font-medium text-gray-900 mb-4">Dimensions</h3>

         <div class="grid grid-cols-2 gap-4">
           <form.Field
             name="default_width"
             children={(field) => (
               <div class="flex flex-col gap-2">
                 <label class="text-sm font-medium text-gray-900">
                   Default Width
                 </label>
                 <input
                   type="number"
                   value={field().state.value}
                   onInput={(e) =>
                     field().handleChange(Number(e.currentTarget.value))
                   }
                   step="0.1"
                   class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                 />
                 <Show when={field().state.meta.errors.length > 0}>
                   <span class="text-sm text-red-600">
                     {field().state.meta.errors[0]}
                   </span>
                 </Show>
               </div>
             )}
           />

           <form.Field
             name="default_height"
             children={(field) => (
               <div class="flex flex-col gap-2">
                 <label class="text-sm font-medium text-gray-900">
                   Default Height
                 </label>
                 <input
                   type="number"
                   value={field().state.value}
                   onInput={(e) =>
                     field().handleChange(Number(e.currentTarget.value))
                   }
                   step="0.1"
                   class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                 />
                 <Show when={field().state.meta.errors.length > 0}>
                   <span class="text-sm text-red-600">
                     {field().state.meta.errors[0]}
                   </span>
                 </Show>
               </div>
             )}
           />
         </div>

         <div class="grid grid-cols-2 gap-4 mt-4">
           <form.Field
             name="min_width"
             children={(field) => (
               <div class="flex flex-col gap-2">
                 <label class="text-sm font-medium text-gray-900">
                   Minimum Width
                 </label>
                 <input
                   type="number"
                   value={field().state.value}
                   onInput={(e) =>
                     field().handleChange(Number(e.currentTarget.value))
                   }
                   step="0.1"
                   class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                 />
                 <Show when={validateMinDimensions("min_width")}>
                   <span class="text-sm text-red-600">
                     {validateMinDimensions("min_width")}
                   </span>
                 </Show>
               </div>
             )}
           />

           <form.Field
             name="min_height"
             children={(field) => (
               <div class="flex flex-col gap-2">
                 <label class="text-sm font-medium text-gray-900">
                   Minimum Height
                 </label>
                 <input
                   type="number"
                   value={field().state.value}
                   onInput={(e) =>
                     field().handleChange(Number(e.currentTarget.value))
                   }
                   step="0.1"
                   class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                 />
                 <Show when={validateMinDimensions("min_height")}>
                   <span class="text-sm text-red-600">
                     {validateMinDimensions("min_height")}
                   </span>
                 </Show>
               </div>
             )}
           />
         </div>
       </div>

       <form.Field
         name="default_price_group_id"
         children={(field) => (
           <Form.Select
             field={field}
             label="Default Price Group"
             placeholder="Select a price group..."
             options={
               priceGroupsQuery.data?.items?.map((pg) => ({
                 label: `${pg.name} ($${(pg.base_price_cents / 100).toFixed(2)})`,
                 value: pg.id,
               })) ?? []
             }
           />
         )}
       />

      <form.Field
        name="svg_url"
        children={(field) => (
          <Form.FileUpload
            field={field}
            label="SVG File"
            uploadPath="catalog-items"
            accept=".svg"
            fileTypeLabel="SVG"
            description="Upload the SVG file for this catalog item"
            multiple={false}
          />
        )}
      />

      <form.Field
        name="is_active"
        children={(field) => (
          <label class="flex items-center gap-2">
            <input
              type="checkbox"
              checked={field().state.value}
              onChange={(e) => field().handleChange(e.currentTarget.checked)}
              class="rounded border-gray-300"
            />
            <span class="text-sm font-medium text-gray-900">Active</span>
          </label>
        )}
      />

      <div class="border-t pt-4">
        <h3 class="text-sm font-medium text-gray-900 mb-3">Tags</h3>

        <div class="relative mb-4">
          <input
            type="text"
            value={tagInput()}
            onInput={(e) => {
              setTagInput(e.currentTarget.value);
              setShowTagSuggestions(true);
            }}
            onFocus={() => setShowTagSuggestions(true)}
            onBlur={() => {
              setTimeout(() => setShowTagSuggestions(false), 150);
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                const input = tagInput().trim();
                if (input && !tags().includes(input)) {
                  setTags([...tags(), input]);
                  setTagInput("");
                }
              }
            }}
            placeholder="Type to add a tag or press Enter..."
            class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          />

          <Show when={showTagSuggestions() && filteredSuggestions().length > 0}>
            <div class="absolute top-full left-0 right-0 mt-1 bg-white border border-input rounded-md shadow-lg z-10">
              <For each={filteredSuggestions()}>
                {(suggestion) => (
                  <button
                    type="button"
                    onClick={(e) => {
                      e.preventDefault();
                      if (!tags().includes(suggestion)) {
                        setTags([...tags(), suggestion]);
                      }
                      setTagInput("");
                      setShowTagSuggestions(false);
                    }}
                    class="w-full text-left px-3 py-2 hover:bg-gray-100 text-sm"
                  >
                    {suggestion}
                  </button>
                )}
              </For>
            </div>
          </Show>
        </div>

        <div class="flex flex-wrap gap-2">
          <For each={tags()}>
            {(tag) => (
              <Badge variant="secondary" class="flex items-center gap-2">
                {tag}
                <button
                  type="button"
                  onClick={() => setTags(tags().filter((t) => t !== tag))}
                  class="ml-1 hover:text-red-600"
                >
                  ✕
                </button>
              </Badge>
            )}
          </For>
        </div>
      </div>

      <Button type="submit" disabled={props.isLoading}>
        {props.isLoading ? "Saving..." : "Save"}
      </Button>
    </form>
  );
}
