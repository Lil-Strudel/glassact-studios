import { CatalogItem, GET } from "@glassact/data";
import {
  Button,
  Pagination,
  PaginationEllipsis,
  PaginationItem,
  PaginationItems,
  PaginationNext,
  PaginationPrevious,
} from "@glassact/ui";
import { For, Show } from "solid-js";
import { ItemDetailModal } from "./item-detail-modal";
import { GraniteImageLightbox } from "../granite/granite-image-lightbox";
import PriceGroupTag from "../price-group-tag";
import { useScrollToTopOnChange } from "../../hooks/use-scroll-to-top-on-change";

interface CatalogGridProps {
  items: GET<CatalogItem>[];
  isLoading: boolean;
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export function CatalogGrid(props: CatalogGridProps) {
  let gridTop: HTMLDivElement | undefined;
  useScrollToTopOnChange(
    () => props.page,
    () => gridTop,
  );

  return (
    <div ref={gridTop} class="flex-1 flex flex-col gap-6 scroll-mt-6">
      <Show
        when={!props.isLoading && props.items.length > 0}
        fallback={
          <Show when={props.isLoading}>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <For each={Array.from({ length: 6 })}>
                {() => (
                  <div class="h-80 animate-pulse rounded-lg border border-border bg-card" />
                )}
              </For>
            </div>
          </Show>
        }
      >
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <For each={props.items}>
            {(item) => (
              <div class="group flex flex-col overflow-hidden rounded-lg border border-border bg-card transition-colors hover:border-input">
                <GraniteImageLightbox
                  images={[{ src: item.svg_url, alt: item.name }]}
                  inches={{ width: item.default_width, height: item.default_height }}
                  triggerClass="came-frame m-4 mb-0 block h-60 overflow-hidden rounded-arch bg-plaque"
                  previewClass="h-full w-full px-6 pb-4 pt-10"
                >
                  <img
                    src={item.svg_url}
                    alt={item.name}
                    class="w-full h-full object-contain"
                  />
                </GraniteImageLightbox>

                <div class="flex flex-1 flex-col gap-3 p-4">
                  <code class="self-start rounded bg-accent px-2 py-0.5 font-mono text-xs tracking-wide text-muted-foreground">
                    {item.catalog_code}
                  </code>

                  <div class="flex items-end justify-between gap-2">
                    <div class="min-w-0">
                      <h3 class="line-clamp-2 text-base text-foreground">
                        {item.name}
                      </h3>
                      <p class="mt-1 text-xs text-faint">{item.category}</p>
                    </div>
                    <div class="flex flex-shrink-0">
                      <PriceGroupTag
                        priceGroupId={item.default_price_group_id}
                      />
                    </div>
                  </div>

                  <ItemDetailModal item={item}>
                    <Button variant="secondary" size="sm" class="mt-auto w-full">
                      View details
                    </Button>
                  </ItemDetailModal>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>

      <Show when={!props.isLoading && props.items.length === 0}>
        <div class="flex flex-1 flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-input px-6 py-12 text-center">
          <div class="ornament-rule w-40" aria-hidden="true">
            <span class="lozenge" />
          </div>
          <h3 class="text-xl text-foreground">No inlays match</h3>
          <p class="text-sm text-muted-foreground">
            Try adjusting your filters or search.
          </p>
        </div>
      </Show>

      <Show when={!props.isLoading && props.totalPages > 1}>
        <div class="flex justify-center">
          <Pagination
            count={props.totalPages}
            page={props.page}
            onPageChange={props.onPageChange}
            itemComponent={(p) => (
              <PaginationItem page={p.page}>{p.page}</PaginationItem>
            )}
            ellipsisComponent={() => <PaginationEllipsis />}
          >
            <PaginationPrevious />
            <PaginationItems />
            <PaginationNext />
          </Pagination>
        </div>
      </Show>
    </div>
  );
}
