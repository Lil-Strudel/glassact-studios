import { CatalogItem, GET } from "@glassact/data";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Button,
} from "@glassact/ui";
import { Show } from "solid-js";
import { IoClose } from "solid-icons/io";

interface ItemDetailModalProps {
  item: GET<CatalogItem>;
  children?: any;
}

export function ItemDetailModal(props: ItemDetailModalProps) {
  return (
    <Dialog>
      <Show when={props.children} fallback={<DialogTrigger as={Button} variant="outline">View Details</DialogTrigger>}>
        {props.children}
      </Show>
      <DialogContent class="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{props.item.name}</DialogTitle>
        </DialogHeader>

        <div class="flex flex-col gap-6">
          {/* SVG Preview */}
          <div class="bg-gray-50 rounded-md p-4 flex items-center justify-center min-h-[400px]">
            <img
              src={props.item.svg_url}
              alt={props.item.name}
              class="max-w-full max-h-[400px] object-contain"
            />
          </div>

          {/* Details */}
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-sm font-medium text-gray-900">Code</label>
              <p class="text-sm text-gray-600 mt-1">
                <code class="bg-gray-100 px-2 py-1 rounded">
                  {props.item.catalog_code}
                </code>
              </p>
            </div>

            <div>
              <label class="text-sm font-medium text-gray-900">Category</label>
              <p class="text-sm text-gray-600 mt-1">{props.item.category}</p>
            </div>

            <Show when={props.item.description}>
              {(desc) => (
                <div class="col-span-2">
                  <label class="text-sm font-medium text-gray-900">
                    Description
                  </label>
                  <p class="text-sm text-gray-600 mt-1">{desc()}</p>
                </div>
              )}
            </Show>

            <div>
              <label class="text-sm font-medium text-gray-900">
                Default Dimensions
              </label>
              <p class="text-sm text-gray-600 mt-1">
                {props.item.default_width} x {props.item.default_height}
              </p>
            </div>

            <div>
              <label class="text-sm font-medium text-gray-900">
                Minimum Dimensions
              </label>
              <p class="text-sm text-gray-600 mt-1">
                {props.item.min_width} x {props.item.min_height}
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
