import { createFileRoute, stripSearchParams } from "@tanstack/solid-router";
import { createMemo } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { browseCatalogOpts } from "../../queries/catalog-browse";
import { FilterSidebar } from "../../components/catalog/filter-sidebar";
import { CatalogGrid } from "../../components/catalog/catalog-grid";
import { useClampPage } from "../../hooks/use-clamp-page";
import { useDebouncedSearchInput } from "../../hooks/use-debounced-search-input";
import {
  catalogBrowseSearchDefaults,
  catalogBrowseSearchSchema,
  type CatalogBrowseSearchInput,
} from "../../utils/list-search";

const PAGE_SIZE = 50;

export const Route = createFileRoute("/_app/catalog/")({
  validateSearch: catalogBrowseSearchSchema,
  search: { middlewares: [stripSearchParams(catalogBrowseSearchDefaults)] },
  component: RouteComponent,
});

function RouteComponent() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();

  function updateSearch(patch: CatalogBrowseSearchInput) {
    navigate({
      search: (prev) => ({ ...prev, ...patch }),
      replace: true,
      resetScroll: false,
    });
  }

  const [searchText, setSearchText] = useDebouncedSearchInput(
    () => search().q,
    (q) => updateSearch({ q, page: 1 }),
  );

  const query = useQuery(() =>
    browseCatalogOpts({
      search: search().q,
      category: search().category,
      limit: PAGE_SIZE,
      offset: (search().page - 1) * PAGE_SIZE,
    }),
  );

  const totalPages = createMemo(() =>
    Math.max(1, Math.ceil((query.data?.total ?? 0) / PAGE_SIZE)),
  );

  useClampPage({
    page: () => search().page,
    pageCount: totalPages,
    hasSettled: () => query.isSuccess && !query.isPlaceholderData,
    setPage: (page) => updateSearch({ page }),
  });

  return (
    <div class="flex flex-col lg:flex-row gap-6">
      <FilterSidebar
        searchValue={searchText()}
        selectedCategory={search().category}
        onSearchChange={setSearchText}
        onCategoryChange={(category) => updateSearch({ category, page: 1 })}
      />

      <CatalogGrid
        items={query.data?.items ?? []}
        isLoading={query.isLoading}
        page={search().page}
        totalPages={totalPages()}
        onPageChange={(page) => updateSearch({ page })}
      />
    </div>
  );
}
