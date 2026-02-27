import { createFileRoute } from "@tanstack/solid-router";
import { createSignal, Show } from "solid-js";
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
  const [offset, setOffset] = createSignal(0);

  const limit = 50;

  const query = useQuery(() =>
    browseCatalogOpts({
      search: search(),
      category: category(),
      tags: tags(),
      limit,
      offset: offset(),
    }),
  );

  const handleLoadMore = () => {
    setOffset(offset() + limit);
  };

  const handleSearchChange = (value: string) => {
    setSearch(value);
    setOffset(0);
  };

  const handleCategoryChange = (value: string) => {
    setCategory(value);
    setOffset(0);
  };

  const handleTagsChange = (newTags: string[]) => {
    setTags(newTags);
    setOffset(0);
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
        total={query.data?.total ?? 0}
        currentOffset={offset()}
        limit={limit}
        onLoadMore={handleLoadMore}
      />
    </div>
  );
}
