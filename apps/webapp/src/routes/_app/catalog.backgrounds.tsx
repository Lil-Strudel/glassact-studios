import { createFileRoute, stripSearchParams } from "@tanstack/solid-router";
import { useQuery } from "@tanstack/solid-query";
import { For, Match, Switch, createMemo } from "solid-js";
import { browseBackgroundsOpts } from "../../queries/background";
import { useDebouncedSearchInput } from "../../hooks/use-debounced-search-input";
import { CatalogSectionTabs } from "../../components/catalog/catalog-section-tabs";
import { PageHeader } from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
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
      <PageHeader
        eyebrow="Catalog"
        title="Sandblast backgrounds"
        description="Masks for the stone face. Place one in a project's stone designer, or download the file for your sandblaster."
      />
      <CatalogSectionTabs />

      <input
        placeholder="Search by name, code or tag..."
        value={searchText()}
        onInput={(e) => setSearchText(e.currentTarget.value)}
        class="h-9 w-full max-w-sm rounded-md border border-input bg-background px-3 py-1 text-sm text-foreground placeholder:text-faint focus-visible:border-gilt focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gilt/25"
      />

      <Switch>
        <Match when={query.isLoading}>
          <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
            <For each={Array.from({ length: 8 })}>
              {() => <div class="h-56 animate-pulse rounded-lg border border-border bg-card" />}
            </For>
          </div>
        </Match>
        <Match when={query.isError}>
          <p class="flex items-center gap-2 text-sm text-error-foreground">
            <IoAlertCircleOutline size={16} aria-hidden="true" />
            Couldn't load backgrounds.
          </p>
        </Match>
        <Match when={visible().length === 0}>
          <p class="rounded-lg border border-dashed border-input p-8 text-center text-sm text-faint">
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
                    class="group flex flex-col overflow-hidden rounded-lg border border-border bg-card text-left transition-colors hover:border-input"
                  >
                    <div class="came-frame m-3 mb-0 flex h-40 items-center justify-center rounded-arch bg-plaque px-4 pb-3 pt-6">
                      <img
                        src={background.svg_url}
                        alt={background.name}
                        loading="lazy"
                        class="max-h-full max-w-full object-contain brightness-0"
                      />
                    </div>
                    <div class="flex flex-col gap-1 p-3">
                      <h3 class="line-clamp-1 text-base text-foreground group-hover:text-crimson-light">{background.name}</h3>
                      <code class="font-mono text-xs tracking-wide text-faint">{background.code}</code>
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
