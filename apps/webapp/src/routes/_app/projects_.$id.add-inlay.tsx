import { createFileRoute, Link, useNavigate } from "@tanstack/solid-router";
import {
  Tabs,
  TabsContent,
  TabsIndicator,
  TabsList,
  TabsTrigger,
  Button,
  Breadcrumb,
  Form,
  textfieldLabel,
  showToast,
} from "@glassact/ui";
import { IoClose } from "solid-icons/io";
import { createForm } from "@tanstack/solid-form";
import { createMemo, createSignal, For, Show, untrack } from "solid-js";
import { z } from "zod";
import { zodStringNumber } from "../../utils/zod-string-number";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { browseCatalogOpts } from "../../queries/catalog-browse";
import { useDebounce } from "../../hooks/use-debounce";
import { FilterSidebar } from "../../components/catalog/filter-sidebar";
import { postCatalogInlayOpts, postCustomInlayOpts } from "../../queries/inlay";
import { getProjectOpts } from "../../queries/project";
import { isApiError } from "../../utils/is-api-error";
import type { CatalogItem, GET } from "@glassact/data";

export const Route = createFileRoute("/_app/projects_/$id/add-inlay")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const projectQuery = useQuery(() => getProjectOpts(params().id));

  const projectName = createMemo(() => {
    if (projectQuery.isSuccess) {
      return untrack(() => projectQuery.data.name);
    }
    return "asdfkj";
  });

  function handleSuccess() {
    queryClient.invalidateQueries({
      queryKey: ["project", params().id, "inlays"],
    });
    navigate({ to: `/projects/${params().id}` });
  }

  return (
    <div>
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          { title: projectName(), to: `/projects/${params().id}` },
          {
            title: "Add Inlay",
            to: `/projects/${params().id}/add-inlay`,
          },
        ]}
      />
      <Tabs defaultValue="catalog">
        <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
          Add Inlay
        </h1>
        <div class="max-w-[400px] mx-auto mt-4">
          <TabsList>
            <TabsTrigger value="catalog">Catalog</TabsTrigger>
            <TabsTrigger value="custom">Custom</TabsTrigger>
            <TabsIndicator />
          </TabsList>
        </div>
        <div class="mt-6">
          <TabsContent value="catalog">
            <CatalogSelector
              projectUuid={params().id}
              onSuccess={handleSuccess}
            />
          </TabsContent>
          <TabsContent value="custom">
            <CustomInlayForm
              projectUuid={params().id}
              onSuccess={handleSuccess}
            />
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}

interface CatalogSelectorProps {
  projectUuid: string;
  onSuccess: () => void;
}

function CatalogSelector(props: CatalogSelectorProps) {
  const [search, setSearch] = createSignal("");
  const [category, setCategory] = createSignal("");
  const [tags, setTags] = createSignal<string[]>([]);
  const [offset, setOffset] = createSignal(0);
  const [selectedItem, setSelectedItem] = createSignal<GET<CatalogItem> | null>(
    null,
  );
  const [customizationNotes, setCustomizationNotes] = createSignal("");

  const limit = 50;

  const debouncedSearch = useDebounce(search, 300);

  const query = useQuery(() =>
    browseCatalogOpts({
      search: debouncedSearch(),
      category: category(),
      tags: tags(),
      limit,
      offset: offset(),
    }),
  );

  const postCatalogInlay = useMutation(postCatalogInlayOpts);

  function handleConfirm() {
    const item = selectedItem();
    if (!item) return;

    postCatalogInlay.mutate(
      {
        projectUuid: props.projectUuid,
        body: {
          name: item.name,
          catalog_item_id: item.id,
          customization_notes: customizationNotes() || undefined,
        },
      },
      {
        onSuccess() {
          showToast({
            title: "Inlay added",
            description: `${item.name} has been added to the project.`,
            variant: "success",
          });
          props.onSuccess();
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to add inlay",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  return (
    <Show
      when={!selectedItem()}
      fallback={
        <div class="mx-auto max-w-2xl px-4 sm:px-6 lg:px-0">
          <div class="bg-white border rounded-lg p-6">
            <div class="flex items-start gap-4">
              <div class="bg-gray-50 rounded-md p-3 flex-shrink-0">
                <img
                  src={selectedItem()!.svg_url}
                  alt={selectedItem()!.name}
                  class="w-32 h-32 object-contain"
                />
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-lg font-semibold text-gray-900">
                  {selectedItem()!.name}
                </h3>
                <code class="text-xs font-mono bg-gray-100 px-2 py-1 rounded">
                  {selectedItem()!.catalog_code}
                </code>
                <p class="text-sm text-gray-500 mt-1">
                  {selectedItem()!.category}
                </p>
              </div>
            </div>

            <div class="mt-4">
              <label class="text-sm font-medium text-gray-900">
                Customization Notes
              </label>
              <textarea
                value={customizationNotes()}
                onInput={(e) => setCustomizationNotes(e.currentTarget.value)}
                placeholder="Describe any modifications to the design (colors, size, etc.)"
                class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 min-h-[80px]"
              />
            </div>

            <div class="flex justify-center gap-4 mt-6">
              <Button
                variant="outline"
                onClick={() => {
                  setSelectedItem(null);
                  setCustomizationNotes("");
                }}
                disabled={postCatalogInlay.isPending}
              >
                Back to Catalog
              </Button>
              <Button
                onClick={handleConfirm}
                disabled={postCatalogInlay.isPending}
              >
                {postCatalogInlay.isPending ? "Adding..." : "Add to Project"}
              </Button>
            </div>
          </div>
        </div>
      }
    >
      <div class="flex flex-col lg:flex-row gap-6">
        <FilterSidebar
          searchValue={search()}
          selectedCategory={category()}
          selectedTags={tags()}
          onSearchChange={(value) => {
            setSearch(value);
            setOffset(0);
          }}
          onCategoryChange={(value) => {
            setCategory(value);
            setOffset(0);
          }}
          onTagsChange={(newTags) => {
            setTags(newTags);
            setOffset(0);
          }}
        />

        <div class="flex-1 flex flex-col gap-6">
          <Show
            when={!query.isLoading && (query.data?.items.length ?? 0) > 0}
            fallback={
              <Show when={query.isLoading}>
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                  <For each={Array.from({ length: 6 })}>
                    {() => (
                      <div class="bg-gray-200 rounded-lg h-64 animate-pulse" />
                    )}
                  </For>
                </div>
              </Show>
            }
          >
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <For each={query.data?.items ?? []}>
                {(item) => (
                  <div
                    class="bg-white border border-gray-200 rounded-lg overflow-hidden hover:shadow-lg transition-shadow cursor-pointer"
                    onClick={() => setSelectedItem(item)}
                  >
                    <div class="bg-gray-50 p-4 flex items-center justify-center h-48 overflow-hidden">
                      <img
                        src={item.svg_url}
                        alt={item.name}
                        class="max-w-full max-h-full object-contain"
                      />
                    </div>
                    <div class="p-4 flex flex-col gap-2">
                      <code class="text-xs font-mono bg-gray-100 px-2 py-1 rounded w-fit">
                        {item.catalog_code}
                      </code>
                      <h3 class="font-semibold text-gray-900 text-sm line-clamp-2">
                        {item.name}
                      </h3>
                      <p class="text-xs text-gray-500">{item.category}</p>
                      <Button variant="outline" class="w-full text-xs mt-1">
                        Select
                      </Button>
                    </div>
                  </div>
                )}
              </For>
            </div>
          </Show>

          <Show when={!query.isLoading && query.data?.items.length === 0}>
            <div class="flex-1 flex items-center justify-center py-12">
              <div class="text-center">
                <h3 class="text-lg font-semibold text-gray-900">
                  No items found
                </h3>
                <p class="text-gray-600 mt-2">
                  Try adjusting your filters or search criteria
                </p>
              </div>
            </div>
          </Show>

          <Show
            when={
              !query.isLoading &&
              query.data &&
              query.data.items.length > 0 &&
              offset() + limit < query.data.total
            }
          >
            <div class="flex justify-center">
              <Button
                onClick={() => setOffset(offset() + limit)}
                variant="outline"
              >
                Load More
              </Button>
            </div>
          </Show>
        </div>
      </div>

      <div class="flex justify-center mt-6">
        <Button
          variant="outline"
          as={Link}
          to={`/projects/${props.projectUuid}`}
        >
          Cancel
        </Button>
      </div>
    </Show>
  );
}

interface CustomInlayFormProps {
  projectUuid: string;
  onSuccess: () => void;
}

function CustomInlayForm(props: CustomInlayFormProps) {
  const postCustomInlay = useMutation(postCustomInlayOpts);

  const customForm = createForm(() => ({
    defaultValues: {
      name: "",
      description: "",
      width: "",
      height: "",
    },
    validators: {
      onChange: z.object({
        name: z.string().min(1, "Name is required"),
        description: z.string().min(1, "Description is required"),
        width: z
          .string()
          .min(1, "Width is required")
          .refine(...zodStringNumber),
        height: z
          .string()
          .min(1, "Height is required")
          .refine(...zodStringNumber),
      }),
    },
    onSubmit: async ({ value }) => {
      postCustomInlay.mutate(
        {
          projectUuid: props.projectUuid,
          body: {
            name: value.name,
            description: value.description,
            requested_width: parseFloat(value.width),
            requested_height: parseFloat(value.height),
          },
        },
        {
          onSuccess() {
            showToast({
              title: "Inlay added",
              description: `${value.name} has been added to the project.`,
              variant: "success",
            });
            props.onSuccess();
          },
          onError(error) {
            if (isApiError(error)) {
              showToast({
                title: "Failed to add inlay",
                description: error?.data?.error ?? "Unknown error",
                variant: "error",
              });
            }
          },
        },
      );
    },
  }));

  return (
    <div class="mx-auto max-w-2xl p-4 sm:px-6 lg:px-0">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          customForm.handleSubmit();
        }}
      >
        <div class="flex gap-8 flex-col">
          <customForm.Field
            name="name"
            children={(field) => (
              <Form.TextField
                field={field}
                label="Inlay Name"
                placeholder="e.g. Memorial Rose Design"
              />
            )}
          />
          <customForm.Field
            name="description"
            children={(field) => (
              <Form.TextArea
                field={field}
                label="Describe what the design will be"
                placeholder="Describe the desired design in detail..."
              />
            )}
          />
          <div>
            <span class={textfieldLabel()}>
              Desired dimensions of the finished piece
            </span>
            <div class="flex items-center gap-4 mt-1">
              <customForm.Field
                name="width"
                children={(field) => (
                  <Form.TextField field={field} placeholder="Width (in)" />
                )}
              />
              <IoClose />
              <customForm.Field
                name="height"
                children={(field) => (
                  <Form.TextField field={field} placeholder="Height (in)" />
                )}
              />
            </div>
          </div>
          <div class="mx-auto flex gap-4">
            <Button
              variant="outline"
              as={Link}
              to={`/projects/${props.projectUuid}`}
              disabled={postCustomInlay.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={postCustomInlay.isPending}>
              {postCustomInlay.isPending ? "Adding..." : "Add to Project"}
            </Button>
          </div>
        </div>
      </form>
    </div>
  );
}
