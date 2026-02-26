import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { postCatalogOpts, postCatalogTagOpts } from "../../queries/catalog";
import { CatalogForm } from "../../components/admin/catalog-form";
import { Show } from "solid-js";
import { CatalogItem, POST } from "@glassact/data";

export const Route = createFileRoute("/_app/admin/catalog_/create")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const postCatalog = useMutation(() => postCatalogOpts());
  const postTag = useMutation(() => postCatalogTagOpts());

  const handleSubmit = async (data: POST<CatalogItem>, tags: string[]) => {
    console.log("here: ", data);
    postCatalog.mutate(data, {
      onSuccess: async (result) => {
        for (const tag of tags) {
          try {
            await postTag.mutateAsync({ uuid: result.uuid, tag });
          } catch (e) {
            console.error("Failed to add tag:", tag, e);
          }
        }

        queryClient.invalidateQueries({ queryKey: ["catalog"] });
        navigate({ to: "/admin/catalog" });
      },
      onError: (error) => {
        console.error("Failed to create catalog item:", error);
      },
    });
  };

  return (
    <div class="flex flex-col gap-6">
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">Create Catalog Item</h1>
          <p class="text-gray-600 mt-1">Add a new catalog item to the system</p>
        </div>
        <button
          onClick={() => navigate({ to: "/admin/catalog" })}
          class="px-4 py-2 bg-gray-200 hover:bg-gray-300 text-gray-800 rounded-md transition-colors"
        >
          Back to catalog
        </button>
      </div>

      <Show when={postCatalog.isError}>
        <div class="bg-red-50 border border-red-200 rounded-md p-4">
          <p class="text-sm text-red-600">
            {postCatalog.error instanceof Error
              ? postCatalog.error.message
              : "Failed to create catalog item"}
          </p>
        </div>
      </Show>

      <CatalogForm
        onSubmit={handleSubmit}
        isLoading={postCatalog.isPending}
        isEditMode={false}
      />
    </div>
  );
}
