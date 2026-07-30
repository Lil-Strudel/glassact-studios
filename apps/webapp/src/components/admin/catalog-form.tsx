import {
  CatalogItem,
  CATALOG_CATEGORIES,
  type CatalogWriteRequest,
  type Manifest,
  GET,
} from "@glassact/data";
import {
  Form,
  Button,
  Alert,
  AlertDescription,
  ComboboxFree,
  ComboboxFreeMulti,
} from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { useMutation, useQuery } from "@tanstack/solid-query";
import { z } from "zod";
import { createMemo, createSignal, onMount, Show } from "solid-js";
import { getCatalogAllTagsOpts } from "../../queries/catalog-browse";
import { getGlassColorsOpts } from "../../queries/glass-colors";
import { getGroutsOpts } from "../../queries/grouts";
import { postUploadOpts, type UploadResponse } from "../../queries/upload";
import PriceGroupCombobox from "../price-group-combobox";
import { ArtworkPanel, type CatalogArtwork } from "./artwork";
import { useContentBBox, isManifestComplete } from "./manifest-editor";

// Pre-loaded editor state for the edit flow: the stored item already has a baked
// SVG (fetched as text) and a finalized manifest.
export interface CatalogFormEditState {
  item: GET<CatalogItem>;
  svgText: string;
  manifest: Manifest;
  tags: string[];
}

interface CatalogFormProps {
  edit?: CatalogFormEditState;
  onSubmit: (req: CatalogWriteRequest) => Promise<void>;
  isLoading?: boolean;
}

const metadataSchema = z.object({
  catalog_code: z.string().min(1).max(255),
  name: z.string().min(1).max(255),
  description: z.string().max(2000),
  category: z.string().min(1).max(255),
  min_width: z.number().positive(),
  min_height: z.number().positive(),
  default_price_group_id: z.number().int().positive(),
  is_active: z.boolean(),
});

export function CatalogForm(props: CatalogFormProps) {
  const uploadMutation = useMutation(postUploadOpts);

  const tagsQuery = useQuery(() => getCatalogAllTagsOpts());
  const glassQuery = useQuery(() => getGlassColorsOpts());
  const groutsQuery = useQuery(() => getGroutsOpts());

  const measureBBox = useContentBBox();

  // The SVG, its manifest and the default size, all edited through the artwork
  // dialogs. Seeded from the stored item on edit.
  const [artwork, setArtwork] = createSignal<CatalogArtwork | null>(null);
  const [category, setCategory] = createSignal(props.edit?.item.category ?? "");
  const [tags, setTags] = createSignal<string[]>(props.edit?.tags ?? []);

  const categoryOptions = createMemo(() => [...CATALOG_CATEGORIES]);
  const tagOptions = createMemo(() => tagsQuery.data ?? []);

  // Measuring needs the hook's offscreen container, which only exists after
  // mount — hence seeding the stored artwork here rather than at signal init.
  onMount(() => {
    const edit = props.edit;
    if (!edit) return;

    setArtwork({
      structureSvg: edit.svgText,
      manifest: edit.manifest,
      warnings: [],
      contentBBox: measureBBox(edit.svgText),
      defaultWidth: edit.item.default_width,
      defaultHeight: edit.item.default_height,
    });
  });

  const form = createForm(() => ({
    defaultValues: {
      catalog_code: props.edit?.item.catalog_code ?? "",
      name: props.edit?.item.name ?? "",
      description: props.edit?.item.description ?? "",
      min_width: props.edit?.item.min_width ?? ("" as unknown as number),
      min_height: props.edit?.item.min_height ?? ("" as unknown as number),
      default_price_group_id: props.edit?.item.default_price_group_id ?? 0,
      is_active: props.edit?.item.is_active ?? true,
    },
    onSubmit: async ({ value }) => {
      const art = artwork();
      if (!art) return;

      const bbox = art.contentBBox ?? measureBBox(art.structureSvg);
      if (!bbox) {
        throw new Error("Could not measure SVG content bounds.");
      }

      // Re-upload the working structure SVG to obtain a fresh svg_url; the server
      // bakes it (viewBox from dims at 300 u/in, fit+center, bake colors) on save.
      const file = new File([art.structureSvg], "structure.svg", {
        type: "image/svg+xml",
      });
      const uploaded: UploadResponse = await uploadMutation.mutateAsync({
        file,
        uploadPath: "catalog-items",
      });

      const req: CatalogWriteRequest = {
        catalog_code: value.catalog_code,
        name: value.name,
        description: value.description ? value.description : null,
        category: category(),
        default_width: art.defaultWidth,
        default_height: art.defaultHeight,
        min_width: value.min_width,
        min_height: value.min_height,
        default_price_group_id: value.default_price_group_id,
        svg_url: uploaded.url,
        manifest: art.manifest,
        content_bbox: bbox,
        is_active: value.is_active,
        tags: tags(),
      };

      await props.onSubmit(req);
    },
  }));

  // The minimum size follows the default size: any change to the default resets
  // it. Editing a minimum afterwards is preserved until the default moves again.
  function handleArtworkChange(next: CatalogArtwork) {
    const prev = artwork();
    setArtwork(next);

    const sizeChanged =
      prev == null ||
      prev.defaultWidth !== next.defaultWidth ||
      prev.defaultHeight !== next.defaultHeight;

    if (sizeChanged) {
      form.setFieldValue("min_width", next.defaultWidth);
      form.setFieldValue("min_height", next.defaultHeight);
    }
  }

  const palettesReady = createMemo(
    () => glassQuery.data != null && groutsQuery.data != null,
  );
  const manifestComplete = createMemo(() => {
    const art = artwork();
    return art != null && isManifestComplete(art.manifest);
  });

  // The server rejects a minimum larger than the default (400); catch it here so
  // it shows up next to the fields instead of after a round trip.
  function exceedsDefault(minWidth: unknown, minHeight: unknown): boolean {
    const art = artwork();
    if (!art) return false;
    return (
      (typeof minWidth === "number" && minWidth > art.defaultWidth) ||
      (typeof minHeight === "number" && minHeight > art.defaultHeight)
    );
  }

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
      class="flex flex-col gap-6"
    >
      <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
        <form.Field
          name="catalog_code"
          validators={{ onChange: metadataSchema.shape.catalog_code }}
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
          validators={{ onChange: metadataSchema.shape.name }}
          children={(field) => (
            <Form.TextField
              field={field}
              label="Name"
              placeholder="Item name"
            />
          )}
        />
      </div>

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

      <ComboboxFree
        label="Category"
        options={categoryOptions()}
        value={category()}
        onValueChange={setCategory}
        placeholder="e.g., A-ANIMALS"
        description="Pick a suggested category or type a custom one."
      />

      <Show when={palettesReady()}>
        <ArtworkPanel
          value={artwork()}
          measureBBox={measureBBox}
          glassColors={glassQuery.data!}
          grouts={groutsQuery.data!}
          onChange={handleArtworkChange}
        />
      </Show>

      <div class="border-t pt-4">
        <h3 class="mb-1 text-sm font-medium text-gray-900">
          Minimum size (in)
        </h3>
        <p class="mb-4 text-xs text-gray-500">
          Defaults to the artwork's default size. Lower it only when the design
          can be ordered smaller.
        </p>

        <div class="grid grid-cols-2 gap-4">
          <form.Field
            name="min_width"
            validators={{ onChange: metadataSchema.shape.min_width }}
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Minimum Width"
                decimalPlaces={2}
                placeholder="e.g., 1.50"
              />
            )}
          />

          <form.Field
            name="min_height"
            validators={{ onChange: metadataSchema.shape.min_height }}
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Minimum Height"
                decimalPlaces={2}
                placeholder="e.g., 1.50"
              />
            )}
          />
        </div>
      </div>

      <form.Field
        name="default_price_group_id"
        validators={{ onChange: metadataSchema.shape.default_price_group_id }}
        children={(field) => <PriceGroupCombobox field={field} />}
      />

      <ComboboxFreeMulti
        label="Tags"
        options={tagOptions()}
        value={tags()}
        onValueChange={setTags}
        placeholder="Type to add a tag..."
      />

      <form.Field
        name="is_active"
        children={(field) => <Form.Checkbox field={field} label="Active" />}
      />

      <form.Subscribe selector={(state) => state.values}>
        {(values) => (
          <>
            <Show
              when={exceedsDefault(values().min_width, values().min_height)}
            >
              <Alert variant="destructive">
                <AlertDescription>
                  The minimum size cannot be larger than the default size (
                  {artwork()!.defaultWidth.toFixed(2)}" ×{" "}
                  {artwork()!.defaultHeight.toFixed(2)}").
                </AlertDescription>
              </Alert>
            </Show>

            <Button
              type="submit"
              disabled={
                props.isLoading ||
                artwork() == null ||
                !manifestComplete() ||
                exceedsDefault(values().min_width, values().min_height) ||
                uploadMutation.isPending
              }
            >
              {props.isLoading ? "Saving..." : "Save"}
            </Button>
          </>
        )}
      </form.Subscribe>
    </form>
  );
}
