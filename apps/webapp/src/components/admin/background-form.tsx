import { For, Show, createMemo, createSignal, onCleanup } from "solid-js";
import { useMutation } from "@tanstack/solid-query";
import type {
  ContentBBox,
  GET,
  RelatedCatalogItem,
  SandblastBackground,
  SandblastBackgroundWrite,
} from "@glassact/data";
import {
  Alert,
  AlertDescription,
  Button,
  ComboboxFreeMulti,
  FileUpload,
} from "@glassact/ui";
import {
  IoArrowDownOutline,
  IoArrowUpOutline,
  IoCloseOutline,
} from "solid-icons/io";
import { postUploadOpts } from "../../queries/upload";
import { postPrepareBackgroundOpts } from "../../queries/background";
import { CatalogItemPicker } from "../catalog/catalog-item-picker";
import { useContentBBox } from "./manifest-editor";

const FIELD_CLASS =
  "rounded-md border border-input bg-background px-3 py-2 text-sm font-normal text-foreground placeholder:text-faint focus:border-gilt focus:outline-none focus:ring-2 focus:ring-gilt/25";

interface BackgroundFormProps {
  initial?: GET<SandblastBackground>;
  tagOptions: string[];
  isSaving: boolean;
  onSave: (body: SandblastBackgroundWrite) => void;
}

interface BackgroundArtwork {
  svgUrl: string;
  previewSrc: string;
  // Null for the stored mask, which is already trimmed and is sent back as-is.
  contentBBox: ContentBBox | null;
  warnings: string[];
}

export function BackgroundForm(props: BackgroundFormProps) {
  const uploadMutation = useMutation(postUploadOpts);
  const prepareMutation = useMutation(postPrepareBackgroundOpts);
  const measureBBox = useContentBBox();

  const [code, setCode] = createSignal(props.initial?.code ?? "");
  const [name, setName] = createSignal(props.initial?.name ?? "");
  const [description, setDescription] = createSignal(
    props.initial?.description ?? "",
  );
  const [tags, setTags] = createSignal<string[]>(props.initial?.tags ?? []);
  const [isActive, setIsActive] = createSignal(
    props.initial?.is_active ?? true,
  );
  const [related, setRelated] = createSignal<RelatedCatalogItem[]>(
    props.initial?.related_items ?? [],
  );
  const [measureFailed, setMeasureFailed] = createSignal(false);
  const [artwork, setArtwork] = createSignal<BackgroundArtwork | null>(
    props.initial
      ? {
          svgUrl: props.initial.svg_url,
          previewSrc: props.initial.svg_url,
          contentBBox: null,
          warnings: [],
        }
      : null,
  );

  let previewObjectUrl: string | null = null;
  function replacePreviewUrl(next: string | null) {
    if (previewObjectUrl) URL.revokeObjectURL(previewObjectUrl);
    previewObjectUrl = next;
  }
  onCleanup(() => replacePreviewUrl(null));

  async function handleUploaded(svgUrl: string) {
    setMeasureFailed(false);
    let prepared;
    try {
      prepared = await prepareMutation.mutateAsync({ svg_url: svgUrl });
    } catch {
      return;
    }

    const bbox = measureBBox(prepared.mask_svg);
    if (!bbox) {
      setMeasureFailed(true);
      return;
    }

    const previewSrc = URL.createObjectURL(
      new Blob([prepared.mask_svg], { type: "image/svg+xml" }),
    );
    replacePreviewUrl(previewSrc);
    setArtwork({
      svgUrl,
      previewSrc,
      contentBBox: bbox,
      warnings: prepared.warnings,
    });
  }

  const relatedIds = createMemo(
    () => new Set(related().map((item) => item.id)),
  );

  function moveRelated(index: number, delta: number) {
    const next = [...related()];
    const target = index + delta;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];
    setRelated(next);
  }

  const canSave = () =>
    code().trim() !== "" &&
    name().trim() !== "" &&
    artwork() != null &&
    !props.isSaving;

  function save() {
    const art = artwork();
    if (!art) return;
    props.onSave({
      code: code().trim(),
      name: name().trim(),
      description: description().trim() || null,
      tags: tags(),
      is_active: isActive(),
      svg_url: art.svgUrl,
      content_bbox: art.contentBBox,
      related_catalog_item_ids: related().map((item) => item.id),
    });
  }

  return (
    <div class="flex flex-col gap-8 lg:flex-row">
      <section class="flex w-full flex-col gap-3 lg:w-96">
        <h2 class="eyebrow">Artwork</h2>
        <Show when={artwork()}>
          {(art) => (
            <div class="flex aspect-[4/3] items-center justify-center rounded-lg border border-input bg-plaque p-4">
              <img
                src={art().previewSrc}
                alt="Sandblast mask"
                class="max-h-full max-w-full object-contain"
              />
            </div>
          )}
        </Show>
        <FileUpload
          uploadPath="sandblast-backgrounds"
          accept=".svg"
          fileTypeLabel="SVG"
          multiple={false}
          uploadFn={uploadMutation.mutateAsync}
          onUrlChange={(url) => {
            if (typeof url === "string" && url) void handleUploaded(url);
          }}
        />
        <p class="text-xs text-faint">
          Every shape is saved as solid black and the file is trimmed to the
          artwork, which is what dealerships download for their sandblaster.
        </p>
        <Show when={prepareMutation.isPending}>
          <p class="text-sm text-faint">Preparing mask...</p>
        </Show>
        <Show when={prepareMutation.isError}>
          <Alert variant="destructive">
            <AlertDescription>
              {prepareMutation.error instanceof Error
                ? prepareMutation.error.message
                : "Failed to read the SVG."}
            </AlertDescription>
          </Alert>
        </Show>
        <Show when={measureFailed()}>
          <Alert variant="destructive">
            <AlertDescription>
              Could not measure the artwork. The SVG may be empty or contain no
              drawable shapes.
            </AlertDescription>
          </Alert>
        </Show>
        <Show when={(artwork()?.warnings.length ?? 0) > 0}>
          <ul class="list-disc pl-5 text-xs text-warning-foreground">
            <For each={artwork()?.warnings}>
              {(warning) => <li>{warning}</li>}
            </For>
          </ul>
        </Show>
      </section>

      <div class="flex min-w-0 flex-1 flex-col gap-5 rounded-lg border border-border bg-card p-6">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <label class="flex flex-col gap-1.5 text-sm font-medium text-foreground">
            Code
            <input
              value={code()}
              placeholder="SB-HRT-0003"
              onInput={(e) => setCode(e.currentTarget.value)}
              class={FIELD_CLASS}
            />
          </label>
          <label class="flex flex-col gap-1.5 text-sm font-medium text-foreground">
            Name
            <input
              value={name()}
              onInput={(e) => setName(e.currentTarget.value)}
              class={FIELD_CLASS}
            />
          </label>
        </div>
        <label class="flex flex-col gap-1.5 text-sm font-medium text-foreground">
          Description
          <textarea
            rows={2}
            value={description()}
            onInput={(e) => setDescription(e.currentTarget.value)}
            class={FIELD_CLASS}
          />
        </label>
        <ComboboxFreeMulti
          label="Tags"
          options={props.tagOptions}
          value={tags()}
          onValueChange={setTags}
          placeholder="Type to add a tag..."
        />
        <label class="flex items-center gap-2 text-sm text-foreground">
          <input
            type="checkbox"
            checked={isActive()}
            onChange={(e) => setIsActive(e.currentTarget.checked)}
            class="h-4 w-4 rounded border-input accent-primary"
          />
          Offered to dealerships
        </label>

        <section class="flex flex-col gap-2 border-t border-border pt-5">
          <div>
            <h3 class="eyebrow">Related inlays</h3>
            <p class="mt-1 text-xs text-faint">
              Suggested to dealerships, in this order, when this background is
              on their stone.
            </p>
          </div>
          <Show
            when={related().length > 0}
            fallback={
              <p class="text-xs text-faint">
                None yet. Pick some from the catalog below.
              </p>
            }
          >
            <ul class="flex flex-col gap-1.5">
              <For each={related()}>
                {(item, index) => (
                  <li class="flex items-center gap-3 rounded-md border border-border bg-background/40 p-2">
                    <img
                      src={item.svg_url}
                      alt=""
                      class="h-10 w-10 flex-shrink-0 rounded border border-input bg-plaque object-contain p-0.5"
                    />
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-sm text-foreground">
                        {item.name}
                      </p>
                      <p class="text-xs tabular-nums text-faint">
                        {item.catalog_code}
                        <Show when={!item.is_active}>
                          <span class="ml-2 text-warning-foreground">
                            Disabled in the catalog, not shown
                          </span>
                        </Show>
                      </p>
                    </div>
                    <Button
                      size="sm"
                      variant="ghost"
                      aria-label="Move up"
                      disabled={index() === 0}
                      onClick={() => moveRelated(index(), -1)}
                    >
                      <IoArrowUpOutline size={16} />
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      aria-label="Move down"
                      disabled={index() === related().length - 1}
                      onClick={() => moveRelated(index(), 1)}
                    >
                      <IoArrowDownOutline size={16} />
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      aria-label="Remove"
                      onClick={() =>
                        setRelated(related().filter((r) => r.id !== item.id))
                      }
                    >
                      <IoCloseOutline size={16} />
                    </Button>
                  </li>
                )}
              </For>
            </ul>
          </Show>
          <CatalogItemPicker
            title="Add from the catalog"
            class="mt-2"
            onPick={(item) => {
              if (relatedIds().has(item.id)) return;
              setRelated([
                ...related(),
                {
                  id: item.id,
                  uuid: item.uuid,
                  catalog_code: item.catalog_code,
                  name: item.name,
                  svg_url: item.svg_url,
                  default_width: item.default_width,
                  default_height: item.default_height,
                  is_active: item.is_active,
                },
              ]);
            }}
          />
        </section>

        <div class="flex items-center justify-end gap-3 border-t border-border pt-5">
          <Show when={!artwork()}>
            <p class="text-xs text-faint">Upload the SVG to save.</p>
          </Show>
          <Button disabled={!canSave()} onClick={save}>
            {props.isSaving ? "Saving..." : "Save background"}
          </Button>
        </div>
      </div>
    </div>
  );
}
