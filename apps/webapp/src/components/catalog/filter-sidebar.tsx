import { For } from "solid-js";
import { TextField, TextFieldRoot, Button } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import { getCatalogCategoriesOpts } from "../../queries/catalog-browse";

interface FilterSidebarProps {
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  searchValue: string;
  selectedCategory: string;
}

export function FilterSidebar(props: FilterSidebarProps) {
  const categoriesQuery = useQuery(() => getCatalogCategoriesOpts());

  const clearFilters = () => {
    props.onSearchChange("");
    props.onCategoryChange("");
  };

  return (
    <div class="w-full lg:w-64 flex-shrink-0 bg-white lg:border-r border-gray-200 p-4 lg:p-6">
      <div class="flex flex-col gap-6">
        <div>
          <label class="text-sm font-medium text-gray-900">Search</label>
          <TextFieldRoot
            value={props.searchValue}
            onChange={(value) => props.onSearchChange(value)}
          >
            <TextField placeholder="Search by name or code..." class="mt-2" />
          </TextFieldRoot>
        </div>

        <div>
          <label class="text-sm font-medium text-gray-900">Category</label>
          <select
            value={props.selectedCategory}
            onChange={(e) => props.onCategoryChange(e.currentTarget.value)}
            class="w-full mt-2 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          >
            <option value="">All categories</option>
            <For each={categoriesQuery.isSuccess ? categoriesQuery.data : []}>
              {(category) => <option value={category}>{category}</option>}
            </For>
          </select>
        </div>

        <Button variant="outline" class="w-full" onClick={clearFilters}>
          Clear filters
        </Button>
      </div>
    </div>
  );
}
