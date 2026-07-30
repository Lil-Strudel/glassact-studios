import { createMemo, createSignal, Show } from "solid-js";
import type { ContentBBox, GlassColor, Grout, GET } from "@glassact/data";
import { Alert, AlertDescription, Badge, Button } from "@glassact/ui";
import { isManifestComplete } from "../manifest-editor";
import { SvgPreview } from "../../customizer/shared";
import { ModifySvgDialog } from "./modify-svg-dialog";
import { UploadSvgDialog } from "./upload-svg-dialog";
import type { CatalogArtwork } from "./types";

interface ArtworkPanelProps {
  value: CatalogArtwork | null;
  measureBBox: (svgText: string) => ContentBBox | null;
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
  onChange: (artwork: CatalogArtwork) => void;
}

// The catalog form's artwork section: an upload button when empty, otherwise a
// preview with the two editing entry points. All the heavy editing UI lives in
// the dialogs so the form page stays a metadata form.
export function ArtworkPanel(props: ArtworkPanelProps) {
  const [uploadOpen, setUploadOpen] = createSignal(false);
  const [modifyOpen, setModifyOpen] = createSignal(false);

  const unassignedCount = createMemo(() => {
    const manifest = props.value?.manifest;
    if (!manifest) return 0;
    return (
      Object.values(manifest.glass_regions).filter(
        (region) => region.glass_color_id == null,
      ).length + (manifest.grout_region.grout_id == null ? 1 : 0)
    );
  });

  const groupCount = createMemo(
    () => Object.keys(props.value?.manifest.glass_regions ?? {}).length,
  );

  const isComplete = createMemo(
    () => props.value != null && isManifestComplete(props.value.manifest),
  );

  return (
    <div class="border-t pt-4">
      <h3 class="mb-1 text-sm font-medium text-gray-900">Artwork</h3>
      <p class="mb-4 text-xs text-gray-500">
        Upload the raw SVG and set its default size, then assign every glass
        group and the grout color.
      </p>

      <Show
        when={props.value}
        fallback={
          <Button type="button" onClick={() => setUploadOpen(true)}>
            Upload SVG
          </Button>
        }
      >
        {(artwork) => (
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
            <div class="h-32 w-32 shrink-0 rounded-lg border border-gray-200 bg-gray-100 p-2">
              <Show when={artwork().structureSvg} keyed>
                {(svgText) => (
                  <SvgPreview
                    svgText={svgText}
                    manifest={artwork().manifest}
                    glassColors={props.glassColors}
                    grouts={props.grouts}
                  />
                )}
              </Show>
            </div>

            <div class="flex flex-col items-start gap-2">
              <p class="text-sm text-gray-600">
                Size{" "}
                <span class="font-medium text-gray-900">
                  {artwork().defaultWidth.toFixed(2)}" ×{" "}
                  {artwork().defaultHeight.toFixed(2)}"
                </span>
              </p>

              <div class="flex flex-wrap items-center gap-2">
                <Badge variant="secondary" class="rounded-full">
                  {groupCount()} color {groupCount() === 1 ? "group" : "groups"}
                </Badge>
                <Show
                  when={unassignedCount() > 0}
                  fallback={
                    <Badge variant="secondary" class="rounded-full">
                      All regions assigned
                    </Badge>
                  }
                >
                  <Badge variant="warning" class="rounded-full">
                    {unassignedCount()} region
                    {unassignedCount() === 1 ? "" : "s"} unassigned
                  </Badge>
                </Show>
              </div>

              <div class="flex flex-wrap gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setUploadOpen(true)}
                >
                  Upload new SVG
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setModifyOpen(true)}
                >
                  Modify SVG
                </Button>
              </div>
            </div>
          </div>
        )}
      </Show>

      <Show when={props.value && !isComplete()}>
        <Alert variant="destructive" class="mt-4">
          <AlertDescription>
            Every glass group and the grout region must have a color assigned
            before saving. Open Modify SVG to assign them.
          </AlertDescription>
        </Alert>
      </Show>

      <UploadSvgDialog
        open={uploadOpen()}
        onOpenChange={setUploadOpen}
        measureBBox={props.measureBBox}
        glassColors={props.glassColors}
        grouts={props.grouts}
        onSave={props.onChange}
      />

      <Show when={props.value}>
        {(artwork) => (
          <ModifySvgDialog
            open={modifyOpen()}
            onOpenChange={setModifyOpen}
            artwork={artwork()}
            glassColors={props.glassColors}
            grouts={props.grouts}
            onSave={props.onChange}
          />
        )}
      </Show>
    </div>
  );
}
