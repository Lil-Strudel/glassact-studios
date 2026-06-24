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
import { createSignal, For, Show } from "solid-js";
import { z } from "zod";
import { zodStringNumber } from "../../utils/zod-string-number";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { browseCatalogOpts } from "../../queries/catalog-browse";
import { useDebounce } from "../../hooks/use-debounce";
import { FilterSidebar } from "../../components/catalog/filter-sidebar";
import { postCatalogInlayOpts, postCustomInlayOpts } from "../../queries/inlay";
import { getProjectOpts } from "../../queries/project";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/projects_/$id/add-inlay/")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const projectQuery = useQuery(() => getProjectOpts(params().id));

  const navigate = useNavigate();

  const queryClient = useQueryClient();
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
          {
            title: projectQuery.data?.name ?? "Project",
            to: `/projects/${params().id}`,
          },
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

  function addAsIs(item: { id: number; name: string }) {
    postCatalogInlay.mutate(
      {
        projectUuid: props.projectUuid,
        body: {
          name: item.name,
          catalog_item_id: item.id,
          customization_notes: "",
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
    <>
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
                  <div class="bg-white border border-gray-200 rounded-lg overflow-hidden hover:shadow-lg transition-shadow flex flex-col">
                    <div class="bg-gray-50 p-4 flex items-center justify-center h-48 overflow-hidden">
                      <img
                        src={item.svg_url}
                        alt={item.name}
                        class="max-w-full max-h-full object-contain"
                      />
                    </div>
                    <div class="p-4 flex flex-col gap-2 flex-1">
                      <code class="text-xs font-mono bg-gray-100 px-2 py-1 rounded w-fit">
                        {item.catalog_code}
                      </code>
                      <h3 class="font-semibold text-gray-900 text-sm line-clamp-2">
                        {item.name}
                      </h3>
                      <p class="text-xs text-gray-500">{item.category}</p>
                      <div class="flex flex-col gap-2 mt-auto pt-2">
                        <Button
                          variant="outline"
                          class="w-full text-xs"
                          onClick={() => addAsIs(item)}
                          disabled={postCatalogInlay.isPending}
                        >
                          {postCatalogInlay.isPending
                            ? "Adding..."
                            : "Add as-is"}
                        </Button>
                        <Button
                          as={Link}
                          to={`/projects/${props.projectUuid}/add-inlay/customize/${item.uuid}`}
                          class="w-full text-xs"
                        >
                          Customize
                        </Button>
                      </div>
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
    </>
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
