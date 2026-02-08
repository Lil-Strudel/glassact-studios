import { createSignal, For, Show } from "solid-js";
import { TextField, TextFieldRoot, Button, Badge } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import {
  getCatalogCategoriesOpts,
  getCatalogAllTagsOpts,
} from "../../../../queries/catalog-browse";

interface FilterSidebarProps {
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onTagsChange: (tags: string[]) => void;
  searchValue: string;
  selectedCategory: string;
  selectedTags: string[];
}

export function FilterSidebar(props: FilterSidebarProps) {
  const [showTagSuggestions, setShowTagSuggestions] = createSignal(false);
  const [tagInput, setTagInput] = createSignal("");

  const categoriesQuery = useQuery(getCatalogCategoriesOpts());
  const tagsQuery = useQuery(getCatalogAllTagsOpts());

  const filteredTagSuggestions = () => {
    if (!tagInput() || !tagsQuery.data) return [];
    const input = tagInput().toLowerCase();
    const selectedSet = new Set(props.selectedTags);
    return tagsQuery.data
      .filter(
        (tag) => tag.toLowerCase().includes(input) && !selectedSet.has(tag),
      )
      .slice(0, 10);
  };

  const addTag = (tag: string) => {
    if (!props.selectedTags.includes(tag)) {
      props.onTagsChange([...props.selectedTags, tag]);
    }
    setTagInput("");
    setShowTagSuggestions(false);
  };

  const removeTag = (tag: string) => {
    props.onTagsChange(props.selectedTags.filter((t) => t !== tag));
  };

  const clearFilters = () => {
    props.onSearchChange("");
    props.onCategoryChange("");
    props.onTagsChange([]);
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
            <For each={categoriesQuery.data ?? []}>
              {(category) => <option value={category}>{category}</option>}
            </For>
          </select>
        </div>

        <div>
          <label class="text-sm font-medium text-gray-900">Tags</label>

          <div class="relative mt-2">
            <input
              type="text"
              value={tagInput()}
              onInput={(e) => {
                setTagInput(e.currentTarget.value);
                setShowTagSuggestions(true);
              }}
              onFocus={() => setShowTagSuggestions(true)}
              onBlur={() => {
                setTimeout(() => setShowTagSuggestions(false), 150);
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  const input = tagInput().trim();
                  if (input) {
                    addTag(input);
                  }
                }
              }}
              placeholder="Add tags..."
              class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
            />

            <Show
              when={showTagSuggestions() && filteredTagSuggestions().length > 0}
            >
              <div class="absolute top-full left-0 right-0 mt-1 bg-white border border-input rounded-md shadow-lg z-10 max-h-48 overflow-y-auto">
                <For each={filteredTagSuggestions()}>
                  {(suggestion) => (
                    <button
                      type="button"
                      onClick={(e) => {
                        e.preventDefault();
                        addTag(suggestion);
                      }}
                      class="w-full text-left px-3 py-2 hover:bg-gray-100 text-sm"
                    >
                      {suggestion}
                    </button>
                  )}
                </For>
              </div>
            </Show>
          </div>

          <div class="flex flex-wrap gap-2 mt-3">
            <For each={props.selectedTags}>
              {(tag) => (
                <Badge variant="secondary" class="flex items-center gap-2">
                  {tag}
                  <button
                    type="button"
                    onClick={() => removeTag(tag)}
                    class="ml-1 hover:text-red-600"
                  >
                    ✕
                  </button>
                </Badge>
              )}
            </For>
          </div>
        </div>

        <Button variant="outline" class="w-full" onClick={clearFilters}>
          Clear filters
        </Button>
      </div>
    </div>
  );
}
