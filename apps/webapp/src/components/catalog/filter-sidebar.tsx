import { For, createMemo } from "solid-js";
import { TextField, TextFieldRoot, Button } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import { getCatalogCategoriesOpts } from "../../queries/catalog-browse";
import { GranitePicker } from "../granite/granite-picker";
import { useGranitePreference } from "../../hooks/use-granite-preference";

interface FilterSidebarProps {
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  searchValue: string;
  selectedCategory: string;
}

export function FilterSidebar(props: FilterSidebarProps) {
  const categoriesQuery = useQuery(() => getCatalogCategoriesOpts());
  const granitePreference = useGranitePreference();
  const categories = createMemo(() =>
    categoriesQuery.isSuccess ? categoriesQuery.data : [],
  );

  const clearFilters = () => {
    props.onSearchChange("");
    props.onCategoryChange("");
  };

  return (
    <div class="w-full flex-shrink-0 self-start rounded-lg border border-border bg-card p-5 lg:w-64">
      <div class="flex flex-col gap-6">
        <div>
          <label class="eyebrow">Search</label>
          <TextFieldRoot
            value={props.searchValue}
            onChange={(value) => props.onSearchChange(value)}
          >
            <TextField placeholder="Search by name or code..." class="mt-2" />
          </TextFieldRoot>
        </div>

        <div>
          <label class="eyebrow">Category</label>
          <select
            value={props.selectedCategory}
            onChange={(e) => props.onCategoryChange(e.currentTarget.value)}
            class="mt-2 h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm text-foreground focus-visible:border-gilt focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gilt/25"
          >
            <option value="">All categories</option>
            <For each={categories()}>
              {(category) => <option value={category}>{category}</option>}
            </For>
          </select>
        </div>

        <Button variant="secondary" class="w-full" onClick={clearFilters}>
          Clear filters
        </Button>

        <div class="border-t border-border pt-6">
          <label class="eyebrow">
            Granite backdrop
          </label>
          <p class="mt-1 text-xs text-faint">
            Preview only — does not filter results.
          </p>
          <GranitePicker preference={granitePreference} class="mt-2" />
        </div>
      </div>
    </div>
  );
}
