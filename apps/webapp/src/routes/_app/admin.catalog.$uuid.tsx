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
import { CatalogItem, PATCH } from "@glassact/data";
import { CatalogForm } from "./admin/catalog/catalog-form";
import { Show, createSignal, createMemo } from "solid-js";
import { Button } from "@glassact/ui";

export const Route = createFileRoute("/_app/admin/catalog/$uuid")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [isDeleting, setIsDeleting] = createSignal(false);

  const uuid = createMemo(() => (params as any).uuid);

  const itemQuery = useQuery(getCatalogItemOpts(uuid()) as any);
  const tagsQuery = useQuery(getCatalogTagsOpts(uuid()) as any);
  const patchMutation = useMutation(() => patchCatalogOpts(uuid()));
  const deleteMutation = useMutation(() => deleteCatalogOpts(uuid()));

  const handleSubmit = async (data: PATCH<CatalogItem>, newTags: string[]) => {
    const currentTags = (tagsQuery.data ?? []) as any;

    patchMutation.mutate(data, {
      onSuccess: async () => {
        const tagsToRemove = currentTags.filter(
          (t: any) => !newTags.includes(t),
        );
        const tagsToAdd = newTags.filter((t: any) => !currentTags.includes(t));

        for (const tag of tagsToRemove) {
          try {
            const deleteOpts = deleteCatalogTagOpts(uuid(), tag) as any;
            await deleteOpts().mutationFn();
          } catch (e) {
            console.error("Failed to remove tag:", tag, e);
          }
        }

        for (const tag of tagsToAdd) {
          try {
            const addOpts = postCatalogTagOpts(uuid()) as any;
            await addOpts().mutationFn(tag);
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
    });
  };

  const handleDelete = async () => {
    if (
      !window.confirm(
        "Are you sure you want to delete this catalog item? This action cannot be undone.",
      )
    ) {
      return;
    }

    setIsDeleting(true);
    deleteMutation.mutate(undefined, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["catalog"] });
        navigate({ to: "/admin/catalog" });
      },
      onError: (error) => {
        console.error("Failed to delete catalog item:", error);
        setIsDeleting(false);
      },
    });
  };

  return (
    <div class="flex flex-col gap-6">
      <Show when={itemQuery.data as any}>
        {(item) => (
          <>
            <div class="flex items-center justify-between">
              <div>
                <h1 class="text-2xl font-bold text-gray-900">
                  Edit Catalog Item: {(item() as any).name}
                </h1>
                <p class="text-gray-600 mt-1">
                  Code: {(item() as any).catalog_code}
                </p>
              </div>

              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={isDeleting()}
              >
                {isDeleting() ? "Deleting..." : "Delete Item"}
              </Button>
            </div>

            <Show when={patchMutation.isError}>
              <div class="bg-red-50 border border-red-200 rounded-md p-4">
                <p class="text-sm text-red-600">
                  {patchMutation.error instanceof Error
                    ? patchMutation.error.message
                    : "Failed to update catalog item"}
                </p>
              </div>
            </Show>

            <Show when={deleteMutation.isError}>
              <div class="bg-red-50 border border-red-200 rounded-md p-4">
                <p class="text-sm text-red-600">
                  {deleteMutation.error instanceof Error
                    ? deleteMutation.error.message
                    : "Failed to delete catalog item"}
                </p>
              </div>
            </Show>

            <CatalogForm
              defaultValues={item() as any}
              onSubmit={handleSubmit}
              isLoading={patchMutation.isPending}
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
