import { CatalogItem, GET, POST } from "@glassact/data";
import { Form, Button, Badge } from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { useQuery } from "@tanstack/solid-query";
import { z } from "zod";
import { For, Show, createSignal } from "solid-js";
import { getCatalogAllTagsOpts } from "../../queries/catalog-browse";
import api from "../../queries/api";
import PriceGroupCombobox from "../price-group-combobox";

interface CatalogFormProps {
  defaultValues?: POST<CatalogItem>;
  onSubmit: (data: POST<CatalogItem>, tags: string[]) => Promise<void>;
  isLoading?: boolean;
  isEditMode?: boolean;
}

const catalogSchema = z.object({
  catalog_code: z.string().min(1).max(255),
  name: z.string().min(1).max(255),
  description: z.string().max(2000).nullable(),
  category: z.string().min(1).max(255),
  default_width: z.number().positive(),
  default_height: z.number().positive(),
  min_width: z.number().positive(),
  min_height: z.number().positive(),
  default_price_group_id: z.number().int(),
  svg_url: z.string().min(1),
  is_active: z.boolean(),
});

export function CatalogForm(props: CatalogFormProps) {
  const [tags, setTags] = createSignal<string[]>([]);
  const [tagInput, setTagInput] = createSignal("");
  const [showTagSuggestions, setShowTagSuggestions] = createSignal(false);

  const tagsQuery = useQuery(getCatalogAllTagsOpts());

  const handleFileUpload = async (file: File, uploadPath: string) => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("uploadPath", uploadPath);

    const response = await api.post("/upload", formData);
    return response.data;
  };

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

  const defaultValues: POST<CatalogItem> = props.defaultValues ?? {
    catalog_code: "",
    name: "",
    description: "",
    category: "",
    default_width: "" as unknown as number,
    default_height: "" as unknown as number,
    min_width: "" as unknown as number,
    min_height: "" as unknown as number,
    default_price_group_id: 0,
    svg_url: "",
    is_active: true,
  };

  const form = createForm(() => ({
    defaultValues,
    validators: {
      onBlur: catalogSchema,
      onSubmit: catalogSchema,
    },
    onSubmit: async ({ value }) => {
      await props.onSubmit(value, tags());
    },
  }));

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
          <Form.TextField
            field={field}
            label="Catalog Code"
            placeholder="ABC-234-V2"
          />
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
              <Form.NumberField
                field={field}
                label="Default Width"
                decimalPlaces={2}
                placeholder="e.g., 100.50"
              />
            )}
          />

          <form.Field
            name="default_height"
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Default Height"
                decimalPlaces={2}
                placeholder="e.g., 100.50"
              />
            )}
          />
        </div>

        <div class="grid grid-cols-2 gap-4 mt-4">
          <form.Field
            name="min_width"
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Minimum Width"
                decimalPlaces={2}
                placeholder="e.g., 50.25"
              />
            )}
          />

          <form.Field
            name="min_height"
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Minimum Height"
                decimalPlaces={2}
                placeholder="e.g., 50.25"
              />
            )}
          />
        </div>
      </div>

      <form.Field
        name="default_price_group_id"
        children={(field) => <PriceGroupCombobox field={field} />}
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
            uploadFn={handleFileUpload}
          />
        )}
      />

      <form.Field
        name="is_active"
        children={(field) => <Form.Checkbox field={field} label="Active" />}
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
