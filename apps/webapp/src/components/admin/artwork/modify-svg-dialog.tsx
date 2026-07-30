import { createEffect, createMemo, createSignal, untrack } from "solid-js";
import type { GlassColor, Grout, Manifest, GET } from "@glassact/data";
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@glassact/ui";
import { ManifestEditor } from "../manifest-editor";
import { SizeFields } from "./size-fields";
import { artworkAspect, type CatalogArtwork } from "./types";

interface ModifySvgDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  artwork: CatalogArtwork;
  glassColors: GET<GlassColor>[];
  grouts: GET<Grout>[];
  onSave: (artwork: CatalogArtwork) => void;
}

// Size + color editing for artwork that is already uploaded. Edits are staged in
// local signals and committed to the form only on save.
export function ModifySvgDialog(props: ModifySvgDialogProps) {
  const [manifest, setManifest] = createSignal<Manifest>(
    untrack(() => props.artwork.manifest),
  );
  const [width, setWidth] = createSignal<number | null>(
    untrack(() => props.artwork.defaultWidth),
  );
  const [height, setHeight] = createSignal<number | null>(
    untrack(() => props.artwork.defaultHeight),
  );

  createEffect(() => {
    if (!props.open) return;
    setManifest(props.artwork.manifest);
    setWidth(props.artwork.defaultWidth);
    setHeight(props.artwork.defaultHeight);
  });

  const aspect = createMemo(() => artworkAspect(props.artwork));

  const canSave = createMemo(() => (width() ?? 0) > 0 && (height() ?? 0) > 0);

  function handleSave() {
    const w = width();
    const h = height();
    if (!w || !h) return;

    props.onSave({
      ...props.artwork,
      manifest: manifest(),
      defaultWidth: w,
      defaultHeight: h,
    });
    props.onOpenChange(false);
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent class="max-h-[90vh] w-[95vw] max-w-6xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Modify SVG</DialogTitle>
          <DialogDescription>
            Set the default size and assign a glass color to every group plus the
            grout region.
          </DialogDescription>
        </DialogHeader>

        <SizeFields
          class="border-b pb-4"
          width={width()}
          height={height()}
          aspect={aspect()}
          onChange={(w, h) => {
            setWidth(w);
            setHeight(h);
          }}
        />

        <ManifestEditor
          structureSvg={props.artwork.structureSvg}
          manifest={manifest()}
          warnings={props.artwork.warnings}
          glassColors={props.glassColors}
          grouts={props.grouts}
          onManifestChange={setManifest}
        />

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => props.onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type="button" disabled={!canSave()} onClick={handleSave}>
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
