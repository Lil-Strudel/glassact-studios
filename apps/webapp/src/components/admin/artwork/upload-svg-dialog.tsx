import { createEffect, createMemo, createSignal, For, Show } from "solid-js";
import { useMutation } from "@tanstack/solid-query";
import type {
  ContentBBox,
  GlassColor,
  Grout,
  Manifest,
  GET,
} from "@glassact/data";
import {
  Alert,
  AlertDescription,
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  FileUpload,
} from "@glassact/ui";
import { postCatalogAnalyzeOpts } from "../../../queries/catalog";
import { postUploadOpts } from "../../../queries/upload";
import { SvgPreview } from "../../customizer/shared";
import { SizeFields } from "./size-fields";
import type { CatalogArtwork } from "./types";

interface AnalyzedSvg {
  structureSvg: string;
  manifest: Manifest;
  warnings: string[];
  contentBBox: ContentBBox;
}

interface UploadSvgDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  measureBBox: (svgText: string) => ContentBBox | null;
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
  onSave: (artwork: CatalogArtwork) => void;
}

// Upload -> analyze -> size, staged locally and only handed to the form on save,
// so cancelling a re-upload leaves the current artwork untouched.
export function UploadSvgDialog(props: UploadSvgDialogProps) {
  const uploadMutation = useMutation(postUploadOpts);
  const analyzeMutation = useMutation(postCatalogAnalyzeOpts);

  const [analyzed, setAnalyzed] = createSignal<AnalyzedSvg | null>(null);
  const [measureError, setMeasureError] = createSignal(false);
  const [width, setWidth] = createSignal<number | null>(null);
  const [height, setHeight] = createSignal<number | null>(null);

  createEffect(() => {
    if (!props.open) return;
    setAnalyzed(null);
    setMeasureError(false);
    setWidth(null);
    setHeight(null);
    analyzeMutation.reset();
  });

  async function handleAnalyze(svgUrl: string) {
    setMeasureError(false);
    let result;
    try {
      result = await analyzeMutation.mutateAsync({ svg_url: svgUrl });
    } catch {
      // analyzeMutation.isError surfaces the message below.
      return;
    }

    const bbox = props.measureBBox(result.structure_svg);
    if (!bbox) {
      setMeasureError(true);
      return;
    }

    setAnalyzed({
      structureSvg: result.structure_svg,
      manifest: result.manifest,
      warnings: result.warnings ?? [],
      contentBBox: bbox,
    });
  }

  const aspect = createMemo(() => {
    const svg = analyzed();
    if (!svg) return 1;
    return svg.contentBBox.height / svg.contentBBox.width;
  });

  const pieceCount = createMemo(() => {
    const svg = analyzed();
    if (!svg) return 0;
    return (
      Object.values(svg.manifest.glass_regions).reduce(
        (total, region) => total + region.piece_ids.length,
        0,
      ) + svg.manifest.grout_region.piece_ids.length
    );
  });

  const groupCount = createMemo(
    () => Object.keys(analyzed()?.manifest.glass_regions ?? {}).length,
  );

  const canSave = createMemo(
    () => analyzed() != null && (width() ?? 0) > 0 && (height() ?? 0) > 0,
  );

  function handleSave() {
    const svg = analyzed();
    const w = width();
    const h = height();
    if (!svg || !w || !h) return;

    props.onSave({ ...svg, defaultWidth: w, defaultHeight: h });
    props.onOpenChange(false);
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent class="max-h-[90vh] max-w-lg overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Upload SVG</DialogTitle>
          <DialogDescription>
            The SVG is analyzed into editable color groups. Colors are assigned
            afterwards in Modify SVG.
          </DialogDescription>
        </DialogHeader>

        <Show
          when={analyzed()}
          fallback={
            <div class="flex flex-col gap-2">
              <FileUpload
                uploadPath="catalog-items"
                accept=".svg"
                fileTypeLabel="SVG"
                multiple={false}
                uploadFn={uploadMutation.mutateAsync}
                onUrlChange={(url) => {
                  if (typeof url === "string" && url) handleAnalyze(url);
                }}
              />

              <Show when={analyzeMutation.isPending}>
                <p class="text-sm text-gray-600">Analyzing SVG...</p>
              </Show>

              <Show when={analyzeMutation.isError}>
                <Alert variant="destructive">
                  <AlertDescription>
                    {analyzeMutation.error instanceof Error
                      ? analyzeMutation.error.message
                      : "Failed to analyze SVG."}
                  </AlertDescription>
                </Alert>
              </Show>

              <Show when={measureError()}>
                <Alert variant="destructive">
                  <AlertDescription>
                    Could not measure the artwork's bounds. The SVG may be empty
                    or contain no drawable shapes.
                  </AlertDescription>
                </Alert>
              </Show>
            </div>
          }
        >
          {(svg) => (
            <div class="flex flex-col gap-4">
              <Alert>
                <AlertDescription>
                  Upload successful — {pieceCount()} shapes in {groupCount()}{" "}
                  color {groupCount() === 1 ? "group" : "groups"}.
                </AlertDescription>
              </Alert>

              <div class="h-48 rounded-lg border border-gray-200 bg-gray-100 p-3">
                <SvgPreview
                  svgText={svg().structureSvg}
                  manifest={svg().manifest}
                  glassColors={props.glassColors}
                  grouts={props.grouts}
                />
              </div>

              <SizeFields
                width={width()}
                height={height()}
                aspect={aspect()}
                onChange={(w, h) => {
                  setWidth(w);
                  setHeight(h);
                }}
              />

              <Show when={svg().warnings.length > 0}>
                <Alert>
                  <AlertDescription>
                    <p class="mb-1 font-medium">Analysis notes</p>
                    <ul class="list-inside list-disc text-sm">
                      <For each={svg().warnings}>{(w) => <li>{w}</li>}</For>
                    </ul>
                  </AlertDescription>
                </Alert>
              </Show>
            </div>
          )}
        </Show>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => props.onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type="button" disabled={!canSave()} onClick={handleSave}>
            Save & close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
