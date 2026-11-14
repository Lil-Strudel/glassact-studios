import { CatalogItem, GET } from "@glassact/data";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Button,
} from "@glassact/ui";
import { ParentProps, Show } from "solid-js";
import { GraniteImageLightbox } from "../granite/granite-image-lightbox";

interface ItemDetailModalProps {
  item: GET<CatalogItem>;
}

export function ItemDetailModal(props: ParentProps<ItemDetailModalProps>) {
  return (
    <Dialog>
      <DialogTrigger as="div" class="contents">
        {props.children ?? (
          <Button variant="secondary">View details</Button>
        )}
      </DialogTrigger>
      <DialogContent class="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{props.item.name}</DialogTitle>
        </DialogHeader>

        <div class="flex flex-col gap-6">
          <GraniteImageLightbox
            images={[{ src: props.item.svg_url, alt: props.item.name }]}
            inches={{ width: props.item.default_width, height: props.item.default_height }}
            triggerClass="came-frame block w-full overflow-hidden rounded-md"
            previewClass="p-4 flex items-center justify-center min-h-[400px]"
          >
            <img
              src={props.item.svg_url}
              alt={props.item.name}
              class="max-w-full max-h-[400px] object-contain"
            />
          </GraniteImageLightbox>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="eyebrow">Code</label>
              <p class="mt-1 text-sm tabular-nums text-muted-foreground">
                <code class="rounded bg-accent px-2 py-1 font-mono tracking-wide">
                  {props.item.catalog_code}
                </code>
              </p>
            </div>

            <div>
              <label class="eyebrow">Category</label>
              <p class="mt-1 text-sm tabular-nums text-muted-foreground">{props.item.category}</p>
            </div>

            <Show when={props.item.description}>
              {(desc) => (
                <div class="col-span-2">
                  <label class="eyebrow">
                    Description
                  </label>
                  <p class="mt-1 text-sm tabular-nums text-muted-foreground">{desc()}</p>
                </div>
              )}
            </Show>

            <div>
              <label class="eyebrow">Default dimensions
              </label>
              <p class="mt-1 text-sm tabular-nums text-muted-foreground">
                {props.item.default_width} x {props.item.default_height}
              </p>
            </div>

            <div>
              <label class="eyebrow">Minimum dimensions
              </label>
              <p class="mt-1 text-sm tabular-nums text-muted-foreground">
                {props.item.min_width} x {props.item.min_height}
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
