import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getCatalogItemOpts,
  getCatalogTagsOpts,
  putCatalogOpts,
  deleteCatalogOpts,
} from "../../queries/catalog";
import { getCatalogSvgTextOpts } from "../../queries/customize";
import type { CatalogWriteRequest } from "@glassact/data";
import {
  CatalogForm,
  type CatalogFormEditState,
} from "../../components/admin/catalog-form";
import { createMemo, Show } from "solid-js";
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
  const svgQuery = useQuery(() => getCatalogSvgTextOpts(params().uuid));
  const putCatalog = useMutation(() => putCatalogOpts());
  const deleteCatalog = useMutation(() => deleteCatalogOpts());

  // The editor needs the stored SVG text + the finalized manifest before it can
  // render. Tags are carried in the write request, so no separate tag mutations.
  const editState = createMemo<CatalogFormEditState | null>(() => {
    const item = itemQuery.data;
    const svgText = svgQuery.data;
    if (!item || !item.manifest || svgText == null) return null;
    return {
      item,
      svgText,
      manifest: item.manifest,
      tags: tagsQuery.data ?? [],
    };
  });

  const handleSubmit = async (req: CatalogWriteRequest) => {
    await putCatalog.mutateAsync(
      { uuid: params().uuid, body: req },
      {
        onSuccess: () => {
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

            <Show when={putCatalog.isError}>
              <div class="bg-red-50 border border-red-200 rounded-md p-4">
                <p class="text-sm text-red-600">
                  {putCatalog.error instanceof Error
                    ? putCatalog.error.message
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

            <Show
              when={editState()}
              fallback={
                <p class="text-gray-600">Loading editor...</p>
              }
            >
              {(state) => (
                <CatalogForm
                  edit={state()}
                  onSubmit={handleSubmit}
                  isLoading={putCatalog.isPending}
                />
              )}
            </Show>
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
