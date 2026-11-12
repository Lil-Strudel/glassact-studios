import { createFileRoute, stripSearchParams } from "@tanstack/solid-router";
import { useQuery } from "@tanstack/solid-query";
import { For, Match, Switch, createMemo } from "solid-js";
import { browseBackgroundsOpts } from "../../queries/background";
import { useDebouncedSearchInput } from "../../hooks/use-debounced-search-input";
import { CatalogSectionTabs } from "../../components/catalog/catalog-section-tabs";
import { BackgroundDetailModal } from "../../components/catalog/background-detail-modal";
import { matchesBackgroundSearch } from "../../components/catalog/background-search";
import { listSearchDefaults, listSearchSchema, type ListSearchInput } from "../../utils/list-search";

export const Route = createFileRoute("/_app/catalog/backgrounds")({
  validateSearch: listSearchSchema,
  search: { middlewares: [stripSearchParams(listSearchDefaults)] },
  component: RouteComponent,
});

function RouteComponent() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();

  function updateSearch(patch: ListSearchInput) {
    navigate({ search: (prev) => ({ ...prev, ...patch }), replace: true, resetScroll: false });
  }

  const [searchText, setSearchText] = useDebouncedSearchInput(
    () => search().q,
    (q) => updateSearch({ q, page: 1 }),
  );

  const query = useQuery(() => browseBackgroundsOpts());
  const visible = createMemo(() =>
    (query.data ?? []).filter((background) => matchesBackgroundSearch(background, search().q)),
  );

  return (
    <div class="flex flex-col gap-6">
      <CatalogSectionTabs />

      <input
        placeholder="Search by name, code or tag..."
        value={searchText()}
        onInput={(e) => setSearchText(e.currentTarget.value)}
        class="w-full max-w-sm rounded-md border border-gray-300 px-3 py-2 text-sm"
      />

      <Switch>
        <Match when={query.isLoading}>
          <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
            <For each={Array.from({ length: 8 })}>
              {() => <div class="h-48 animate-pulse rounded-lg bg-gray-200" />}
            </For>
          </div>
        </Match>
        <Match when={query.isError}>
          <p class="text-sm text-red-600">Couldn't load backgrounds.</p>
        </Match>
        <Match when={visible().length === 0}>
          <p class="rounded-lg border border-dashed border-gray-300 p-8 text-center text-sm text-gray-500">
            No backgrounds match that search.
          </p>
        </Match>
        <Match when={query.isSuccess}>
          <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
            <For each={visible()}>
              {(background) => (
                <BackgroundDetailModal background={background}>
                  <button
                    type="button"
                    class="flex flex-col overflow-hidden rounded-lg border border-gray-200 bg-white text-left transition-shadow hover:shadow-lg"
                  >
                    <div class="flex h-40 w-full items-center justify-center bg-gray-50 p-3">
                      <img
                        src={background.svg_url}
                        alt={background.name}
                        loading="lazy"
                        class="max-h-full max-w-full object-contain brightness-0"
                      />
                    </div>
                    <div class="flex flex-col gap-1 p-3">
                      <h3 class="line-clamp-1 text-sm font-semibold text-gray-900">{background.name}</h3>
                      <code class="text-xs text-gray-500">{background.code}</code>
                    </div>
                  </button>
                </BackgroundDetailModal>
              )}
            </For>
          </div>
        </Match>
      </Switch>
    </div>
  );
}
