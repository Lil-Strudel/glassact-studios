import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { lastAdminCatalogSearch } from "../../utils/list-search";
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
import { Alert, Button } from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
import { PageHeader } from "@glassact/ui";

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
          navigate({ to: "/admin/catalog", search: lastAdminCatalogSearch() });
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
        navigate({ to: "/admin/catalog", search: lastAdminCatalogSearch() });
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
            <PageHeader
              eyebrow="Edit Catalog Item"
              title={item().name}
              description={
                <>
                  Code:{" "}
                  <span class="font-mono tabular-nums text-foreground">
                    {item().catalog_code}
                  </span>
                </>
              }
              actions={
                <Button
                  variant="destructive"
                  onClick={handleDelete}
                  disabled={deleteCatalog.isPending}
                >
                  {deleteCatalog.isPending ? "Deleting..." : "Delete Item"}
                </Button>
              }
            />

            <Show when={putCatalog.isError}>
              <Alert variant="destructive">
                <IoAlertCircleOutline size={18} />
                <p>
                  {putCatalog.error instanceof Error
                    ? putCatalog.error.message
                    : "Failed to update catalog item"}
                </p>
              </Alert>
            </Show>

            <Show when={deleteCatalog.isError}>
              <Alert variant="destructive">
                <IoAlertCircleOutline size={18} />
                <p>
                  {deleteCatalog.error instanceof Error
                    ? deleteCatalog.error.message
                    : "Failed to delete catalog item"}
                </p>
              </Alert>
            </Show>

            <Show
              when={editState()}
              fallback={<p class="text-sm text-faint">Loading editor...</p>}
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
        <p class="py-12 text-center text-sm text-faint">Loading...</p>
      </Show>

      <Show when={itemQuery.isError}>
        <Alert variant="destructive">
          <IoAlertCircleOutline size={18} />
          <p>Failed to load catalog item</p>
        </Alert>
      </Show>
    </div>
  );
}
