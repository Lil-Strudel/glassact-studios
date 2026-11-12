import { For, Show, createMemo, createSignal } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import type { CatalogItem, GET } from "@glassact/data";
import { cn } from "@glassact/ui";
import { browseCatalogOpts } from "../../queries/catalog-browse";
import { useDebouncedSearchInput } from "../../hooks/use-debounced-search-input";

interface CatalogItemPickerProps {
  title: string;
  description?: string;
  onPick: (item: GET<CatalogItem>) => void;
  isDisabled?: boolean;
  class?: string;
}

export function CatalogItemPicker(props: CatalogItemPickerProps) {
  const [search, setSearch] = createSignal("");
  const [searchText, setSearchText] = useDebouncedSearchInput(search, setSearch);
  const query = useQuery(() => browseCatalogOpts({ search: search(), limit: 24 }));
  const items = createMemo(() => query.data?.items ?? []);

  return (
    <section class={cn("flex flex-col gap-2", props.class)}>
      <h3 class="text-sm font-medium text-gray-900">{props.title}</h3>
      <Show when={props.description}>
        <p class="text-xs text-gray-500">{props.description}</p>
      </Show>
      <input
        placeholder="Search by name or code..."
        value={searchText()}
        onInput={(e) => setSearchText(e.currentTarget.value)}
        class="rounded-md border border-gray-300 px-3 py-2 text-sm"
      />
      <div class="grid max-h-80 grid-cols-3 gap-2 overflow-y-auto">
        <For each={items()}>
          {(item) => (
            <button
              type="button"
              title={`${item.name} (${item.catalog_code})`}
              disabled={props.isDisabled}
              class="flex aspect-square flex-col items-center justify-center gap-1 rounded-md border border-gray-200 p-1.5 text-[10px] text-gray-600 hover:border-blue-500 disabled:cursor-wait disabled:opacity-50"
              onClick={() => props.onPick(item)}
            >
              <img src={item.svg_url} alt={item.name} loading="lazy" class="min-h-0 flex-1 object-contain" />
              <span class="w-full truncate">{item.catalog_code}</span>
            </button>
          )}
        </For>
      </div>
    </section>
  );
}
