import { createFileRoute } from "@tanstack/solid-router";
import { createMemo, createSignal } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { browseCatalogOpts } from "../../queries/catalog-browse";
import { FilterSidebar } from "../../components/catalog/filter-sidebar";
import { CatalogGrid } from "../../components/catalog/catalog-grid";

export const Route = createFileRoute("/_app/catalog/")({
  component: RouteComponent,
});

function RouteComponent() {
  const [search, setSearch] = createSignal("");
  const [category, setCategory] = createSignal("");
  const [tags, setTags] = createSignal<string[]>([]);
  const [page, setPage] = createSignal(1);

  const limit = 50;

  const query = useQuery(() =>
    browseCatalogOpts({
      search: search(),
      category: category(),
      tags: tags(),
      limit,
      offset: (page() - 1) * limit,
    }),
  );

  const totalPages = createMemo(() =>
    Math.ceil((query.data?.total ?? 0) / limit),
  );

  const handleSearchChange = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handleCategoryChange = (value: string) => {
    setCategory(value);
    setPage(1);
  };

  const handleTagsChange = (newTags: string[]) => {
    setTags(newTags);
    setPage(1);
  };

  return (
    <div class="flex flex-col lg:flex-row gap-6">
      <FilterSidebar
        searchValue={search()}
        selectedCategory={category()}
        selectedTags={tags()}
        onSearchChange={handleSearchChange}
        onCategoryChange={handleCategoryChange}
        onTagsChange={handleTagsChange}
      />

      <CatalogGrid
        items={query.data?.items ?? []}
        isLoading={query.isLoading}
        page={page()}
        totalPages={totalPages()}
        onPageChange={setPage}
      />
    </div>
  );
}
