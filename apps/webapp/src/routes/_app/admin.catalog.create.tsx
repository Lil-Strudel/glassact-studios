import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { postCatalogOpts, postCatalogTagOpts } from "../../queries/catalog";
import { CatalogItem, POST } from "@glassact/data";
import { CatalogForm } from "./admin/catalog/catalog-form";
import { Show } from "solid-js";

export const Route = createFileRoute("/_app/admin/catalog/create")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const postMutation = useMutation(() => postCatalogOpts());

  const handleSubmit = async (data: any, tags: string[]) => {
    const submitData = data as any;
    postMutation.mutate(submitData, {
      onSuccess: async (result: any) => {
        const tagOpts = postCatalogTagOpts(result.uuid) as any;
        for (const tag of tags) {
          try {
            await tagOpts().mutationFn(tag);
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
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Create Catalog Item</h1>
        <p class="text-gray-600 mt-1">Add a new catalog item to the system</p>
      </div>

      <Show when={postMutation.isError}>
        <div class="bg-red-50 border border-red-200 rounded-md p-4">
          <p class="text-sm text-red-600">
            {postMutation.error instanceof Error
              ? postMutation.error.message
              : "Failed to create catalog item"}
          </p>
        </div>
      </Show>

      <CatalogForm
        onSubmit={handleSubmit}
        isLoading={postMutation.isPending}
        isEditMode={false}
      />
    </div>
  );
}
