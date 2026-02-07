import { CatalogItem, GET } from "@glassact/data";
import { Button, Badge } from "@glassact/ui";
import { For, Show } from "solid-js";
import { ItemDetailModal } from "./item-detail-modal";

interface CatalogGridProps {
  items: GET<CatalogItem>[];
  isLoading: boolean;
  total: number;
  currentOffset: number;
  limit: number;
  onLoadMore: () => void;
}

export function CatalogGrid(props: CatalogGridProps) {
  return (
    <div class="flex-1 flex flex-col gap-6">
      {/* Grid */}
      <Show
        when={!props.isLoading && props.items.length > 0}
        fallback={
          <Show when={props.isLoading}>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <For each={Array.from({ length: 6 })}>
                {() => (
                  <div class="bg-gray-200 rounded-lg h-64 animate-pulse" />
                )}
              </For>
            </div>
          </Show>
        }
      >
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <For each={props.items}>
            {(item) => (
              <div class="bg-white border border-gray-200 rounded-lg overflow-hidden hover:shadow-lg transition-shadow">
                {/* Preview */}
                <div class="bg-gray-50 p-4 flex items-center justify-center h-64 overflow-hidden">
                  <img
                    src={item.svg_url}
                    alt={item.name}
                    class="max-w-full max-h-full object-contain"
                  />
                </div>

                {/* Content */}
                <div class="p-4 flex flex-col gap-3">
                  <div>
                    <code class="text-xs font-mono bg-gray-100 px-2 py-1 rounded">
                      {item.catalog_code}
                    </code>
                  </div>

                  <div>
                    <h3 class="font-semibold text-gray-900 text-sm line-clamp-2">
                      {item.name}
                    </h3>
                    <p class="text-xs text-gray-500 mt-1">{item.category}</p>
                  </div>

                  {/* Tags Placeholder */}
                  <div class="text-xs text-gray-400 py-1">
                    Tags available in detail view
                  </div>

                  {/* Details Button */}
                  <ItemDetailModal item={item}>
                    <Button variant="outline" class="w-full text-xs">
                      View Details
                    </Button>
                  </ItemDetailModal>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>

      {/* Empty State */}
      <Show when={!props.isLoading && props.items.length === 0}>
        <div class="flex-1 flex items-center justify-center py-12">
          <div class="text-center">
            <h3 class="text-lg font-semibold text-gray-900">
              No items found
            </h3>
            <p class="text-gray-600 mt-2">
              Try adjusting your filters or search criteria
            </p>
          </div>
        </div>
      </Show>

      {/* Load More */}
      <Show
        when={
          !props.isLoading &&
          props.items.length > 0 &&
          props.currentOffset + props.limit < props.total
        }
      >
        <div class="flex justify-center">
          <Button onClick={props.onLoadMore} variant="outline">
            Load More
          </Button>
        </div>
      </Show>
    </div>
  );
}
