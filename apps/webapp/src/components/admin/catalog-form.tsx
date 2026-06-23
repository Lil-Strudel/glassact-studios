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
  FileUpload,
} from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { useMutation, useQuery } from "@tanstack/solid-query";
import { z } from "zod";
import { createMemo, createSignal, Show } from "solid-js";
import { getCatalogAllTagsOpts } from "../../queries/catalog-browse";
import { postCatalogAnalyzeOpts } from "../../queries/catalog";
import { getGlassColorsOpts } from "../../queries/glass-colors";
import { getGroutsOpts } from "../../queries/grouts";
import { postUploadOpts, type UploadResponse } from "../../queries/upload";
import PriceGroupCombobox from "../price-group-combobox";
import {
  ManifestEditor,
  useContentBBox,
  isManifestComplete,
} from "./manifest-editor";

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
  default_width: z.number().positive(),
  default_height: z.number().positive(),
  min_width: z.number().positive(),
  min_height: z.number().positive(),
  default_price_group_id: z.number().int().positive(),
  is_active: z.boolean(),
});

export function CatalogForm(props: CatalogFormProps) {
  const uploadMutation = useMutation(postUploadOpts);
  const analyzeMutation = useMutation(postCatalogAnalyzeOpts);

  const tagsQuery = useQuery(() => getCatalogAllTagsOpts());
  const glassQuery = useQuery(() => getGlassColorsOpts());
  const groutsQuery = useQuery(() => getGroutsOpts());

  const measureBBox = useContentBBox();

  // The working structure SVG and manifest. Seeded from analyze (create) or from
  // the stored item (edit).
  const [structureSvg, setStructureSvg] = createSignal<string | null>(
    props.edit?.svgText ?? null,
  );
  const [manifest, setManifest] = createSignal<Manifest | null>(
    props.edit?.manifest ?? null,
  );
  const [warnings, setWarnings] = createSignal<string[]>([]);
  const [category, setCategory] = createSignal(props.edit?.item.category ?? "");
  const [tags, setTags] = createSignal<string[]>(props.edit?.tags ?? []);

  const categoryOptions = createMemo(() => [...CATALOG_CATEGORIES]);
  const tagOptions = createMemo(() => tagsQuery.data ?? []);

  const form = createForm(() => ({
    defaultValues: {
      catalog_code: props.edit?.item.catalog_code ?? "",
      name: props.edit?.item.name ?? "",
      description: props.edit?.item.description ?? "",
      default_width:
        props.edit?.item.default_width ?? ("" as unknown as number),
      default_height:
        props.edit?.item.default_height ?? ("" as unknown as number),
      min_width: props.edit?.item.min_width ?? ("" as unknown as number),
      min_height: props.edit?.item.min_height ?? ("" as unknown as number),
      default_price_group_id: props.edit?.item.default_price_group_id ?? 0,
      is_active: props.edit?.item.is_active ?? true,
    },
    onSubmit: async ({ value }) => {
      const svg = structureSvg();
      const m = manifest();
      if (!svg || !m) return;

      const bbox = measureBBox(svg);
      if (!bbox) {
        throw new Error("Could not measure SVG content bounds.");
      }

      // Re-upload the working structure SVG to obtain a fresh svg_url; the server
      // bakes it (viewBox from dims at 300 u/in, fit+center, bake colors) on save.
      const file = new File([svg], "structure.svg", {
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
        default_width: value.default_width,
        default_height: value.default_height,
        min_width: value.min_width,
        min_height: value.min_height,
        default_price_group_id: value.default_price_group_id,
        svg_url: uploaded.url,
        manifest: m,
        content_bbox: bbox,
        is_active: value.is_active,
        tags: tags(),
      };

      await props.onSubmit(req);
    },
  }));

  async function handleAnalyze(svgUrl: string) {
    try {
      const result = await analyzeMutation.mutateAsync({ svg_url: svgUrl });
      setStructureSvg(result.structure_svg);
      setManifest(result.manifest);
      setWarnings(result.warnings ?? []);
    } catch {
      // analyzeMutation.isError surfaces the message below.
    }
  }

  const editorReady = createMemo(
    () => structureSvg() != null && manifest() != null,
  );
  const palettesReady = createMemo(
    () => glassQuery.data != null && groutsQuery.data != null,
  );
  const manifestComplete = createMemo(() => {
    const m = manifest();
    return m != null && isManifestComplete(m);
  });

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

      <div class="border-t pt-4">
        <h3 class="mb-4 text-sm font-medium text-gray-900">Dimensions (in)</h3>

        <div class="grid grid-cols-2 gap-4">
          <form.Field
            name="default_width"
            validators={{ onChange: metadataSchema.shape.default_width }}
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Default Width"
                decimalPlaces={2}
                placeholder="e.g., 3.00"
              />
            )}
          />

          <form.Field
            name="default_height"
            validators={{ onChange: metadataSchema.shape.default_height }}
            children={(field) => (
              <Form.NumberField
                field={field}
                label="Default Height"
                decimalPlaces={2}
                placeholder="e.g., 3.00"
              />
            )}
          />
        </div>

        <div class="mt-4 grid grid-cols-2 gap-4">
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

      <div class="border-t pt-4">
        <h3 class="mb-1 text-sm font-medium text-gray-900">Artwork</h3>
        <p class="mb-4 text-xs text-gray-500">
          Upload the raw SVG, then assign every glass group and the grout color
          in the editor below.
        </p>

        <FileUpload
          label="SVG File"
          uploadPath="catalog-items"
          accept=".svg"
          fileTypeLabel="SVG"
          description="Upload the raw SVG; it will be analyzed into editable color groups."
          multiple={false}
          uploadFn={uploadMutation.mutateAsync}
          onUrlChange={(url) => {
            if (typeof url === "string" && url) handleAnalyze(url);
          }}
        />

        <Show when={analyzeMutation.isPending}>
          <p class="mt-2 text-sm text-gray-600">Analyzing SVG...</p>
        </Show>

        <Show when={analyzeMutation.isError}>
          <Alert variant="destructive" class="mt-2">
            <AlertDescription>
              {analyzeMutation.error instanceof Error
                ? analyzeMutation.error.message
                : "Failed to analyze SVG."}
            </AlertDescription>
          </Alert>
        </Show>
      </div>

      <Show when={editorReady() && palettesReady()}>
        <div class="border-t pt-4">
          <ManifestEditor
            structureSvg={structureSvg()!}
            manifest={manifest()!}
            warnings={warnings()}
            glassColors={glassQuery.data!}
            grouts={groutsQuery.data!}
            onManifestChange={setManifest}
          />
        </div>
      </Show>

      <Show when={editorReady() && !manifestComplete()}>
        <Alert variant="destructive">
          <AlertDescription>
            Every glass group and the grout region must have a color assigned
            before saving.
          </AlertDescription>
        </Alert>
      </Show>

      <Button
        type="submit"
        disabled={
          props.isLoading ||
          !editorReady() ||
          !manifestComplete() ||
          uploadMutation.isPending
        }
      >
        {props.isLoading ? "Saving..." : "Save"}
      </Button>
    </form>
  );
}
