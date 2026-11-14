import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { lastAdminCatalogSearch } from "../../utils/list-search";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { postCatalogOpts } from "../../queries/catalog";
import { CatalogForm } from "../../components/admin/catalog-form";
import { Show } from "solid-js";
import type { CatalogWriteRequest } from "@glassact/data";
import { Alert, Button } from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
import { PageHeader } from "@glassact/ui";

export const Route = createFileRoute("/_app/admin/catalog_/create")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const postCatalog = useMutation(() => postCatalogOpts());

  const handleSubmit = async (req: CatalogWriteRequest) => {
    await postCatalog.mutateAsync(req, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["catalog"] });
        navigate({ to: "/admin/catalog", search: lastAdminCatalogSearch() });
      },
      onError: (error) => {
        console.error("Failed to create catalog item:", error);
      },
    });
  };

  return (
    <div class="flex flex-col gap-6">
      <PageHeader
        eyebrow="Catalog"
        title="Create Catalog Item"
        description="Add a new catalog item to the system"
        actions={
          <Button
            variant="outline"
            onClick={() =>
              navigate({
                to: "/admin/catalog",
                search: lastAdminCatalogSearch(),
              })
            }
          >
            Back to catalog
          </Button>
        }
      />

      <Show when={postCatalog.isError}>
        <Alert variant="destructive">
          <IoAlertCircleOutline size={18} />
          <p>
            {postCatalog.error instanceof Error
              ? postCatalog.error.message
              : "Failed to create catalog item"}
          </p>
        </Alert>
      </Show>

      <CatalogForm onSubmit={handleSubmit} isLoading={postCatalog.isPending} />
    </div>
  );
}
