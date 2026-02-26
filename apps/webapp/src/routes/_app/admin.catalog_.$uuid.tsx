import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getCatalogItemOpts,
  getCatalogTagsOpts,
  patchCatalogOpts,
  postCatalogTagOpts,
  deleteCatalogTagOpts,
  deleteCatalogOpts,
} from "../../queries/catalog";
import { CatalogItem, PATCH, POST } from "@glassact/data";
import { CatalogForm } from "../../components/admin/catalog-form";
import { Show } from "solid-js";
import { Button } from "@glassact/ui";

export const Route = createFileRoute("/_app/admin/catalog_/$uuid")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const itemQuery = useQuery(() => getCatalogItemOpts(params().uuid));
  const tagsQuery = useQuery(() => getCatalogTagsOpts(params().uuid));
  const patchCatalog = useMutation(() => patchCatalogOpts());
  const deleteCatalog = useMutation(() => deleteCatalogOpts());
  const deleteTag = useMutation(() => deleteCatalogTagOpts());
  const postTag = useMutation(() => postCatalogTagOpts());

  const handleSubmit = async (data: POST<CatalogItem>, newTags: string[]) => {
    const currentTags = tagsQuery.data ?? [];

    patchCatalog.mutate(
      { uuid: params().uuid, body: data },
      {
        onSuccess: async () => {
          const tagsToRemove = currentTags.filter(
            (t: any) => !newTags.includes(t),
          );
          const tagsToAdd = newTags.filter(
            (t: any) => !currentTags.includes(t),
          );

          for (const tag of tagsToRemove) {
            try {
              deleteTag.mutateAsync({ uuid: params().uuid, tag });
            } catch (e) {
              console.error("Failed to remove tag:", tag, e);
            }
          }

          for (const tag of tagsToAdd) {
            try {
              postTag.mutateAsync({ uuid: params().uuid, tag });
            } catch (e) {
              console.error("Failed to add tag:", tag, e);
            }
          }

          queryClient.invalidateQueries({ queryKey: ["catalog"] });
          navigate({ to: "/admin/catalog" });
        },
        onError: (error) => {
          console.error("Failed to update catalog item:", error);
        },
      },
    );
  };

  const handleDelete = async () => {
    if (
      !window.confirm(
        "Are you sure you want to delete this catalog item? This action cannot be undone.",
      )
    ) {
      return;
    }

    deleteCatalog.mutate(params().uuid, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["catalog"] });
        navigate({ to: "/admin/catalog" });
      },
      onError: (error) => {
        console.error("Failed to delete catalog item:", error);
      },
    });
  };

  return (
    <div class="flex flex-col gap-6">
      <Show when={itemQuery.data}>
        {(item) => (
          <>
            <div class="flex items-center justify-between">
              <div>
                <h1 class="text-2xl font-bold text-gray-900">
                  Edit Catalog Item: {item().name}
                </h1>
                <p class="text-gray-600 mt-1">Code: {item().catalog_code}</p>
              </div>

              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteCatalog.isPending}
              >
                {deleteCatalog.isPending ? "Deleting..." : "Delete Item"}
              </Button>
            </div>

            <Show when={patchCatalog.isError}>
              <div class="bg-red-50 border border-red-200 rounded-md p-4">
                <p class="text-sm text-red-600">
                  {patchCatalog.error instanceof Error
                    ? patchCatalog.error.message
                    : "Failed to update catalog item"}
                </p>
              </div>
            </Show>

            <Show when={deleteCatalog.isError}>
              <div class="bg-red-50 border border-red-200 rounded-md p-4">
                <p class="text-sm text-red-600">
                  {deleteCatalog.error instanceof Error
                    ? deleteCatalog.error.message
                    : "Failed to delete catalog item"}
                </p>
              </div>
            </Show>

            <CatalogForm
              defaultValues={item()}
              onSubmit={handleSubmit}
              isLoading={patchCatalog.isPending}
              isEditMode={true}
            />
          </>
        )}
      </Show>

      <Show when={itemQuery.isLoading}>
        <div class="text-center py-12">
          <p class="text-gray-600">Loading...</p>
        </div>
      </Show>

      <Show when={itemQuery.isError}>
        <div class="bg-red-50 border border-red-200 rounded-md p-4">
          <p class="text-sm text-red-600">Failed to load catalog item</p>
        </div>
      </Show>
    </div>
  );
}
