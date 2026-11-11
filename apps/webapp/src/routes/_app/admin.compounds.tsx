import { Link, createFileRoute } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { For, Match, Show, Switch } from "solid-js";
import type { CatalogCompound, GET } from "@glassact/data";
import { Button, showToast } from "@glassact/ui";
import { IoAddCircleOutline } from "solid-icons/io";
import {
  deleteCompoundOpts,
  getCompoundsOpts,
  patchCompoundOpts,
} from "../../queries/compound";
import { CompoundPreview } from "../../components/stone-designer/compound-preview";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/compounds")({
  component: RouteComponent,
});

function errorMessage(error: unknown) {
  return error instanceof Error && isApiError(error)
    ? (error.data?.error ?? "Unknown error")
    : "Unknown error";
}

function RouteComponent() {
  const queryClient = useQueryClient();
  const compoundsQuery = useQuery(() => getCompoundsOpts());
  const patchCompound = useMutation(patchCompoundOpts);
  const deleteCompound = useMutation(deleteCompoundOpts);

  const refresh = () => queryClient.invalidateQueries({ queryKey: ["compound"] });

  function toggleActive(compound: GET<CatalogCompound>) {
    patchCompound.mutate(
      { uuid: compound.uuid, body: { is_active: !compound.is_active, version: compound.version } },
      {
        onSuccess: refresh,
        onError: (error) =>
          showToast({ title: "Failed to update compound", description: errorMessage(error), variant: "error" }),
      },
    );
  }

  function remove(compound: GET<CatalogCompound>) {
    if (!window.confirm(`Delete "${compound.name}"? Inlays already added to projects are kept.`)) return;
    deleteCompound.mutate(compound.uuid, {
      onSuccess: refresh,
      onError: (error) =>
        showToast({ title: "Failed to delete compound", description: errorMessage(error), variant: "error" }),
    });
  }

  return (
    <div class="flex flex-col gap-6">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">Compound inlays</h1>
          <p class="mt-1 text-gray-600">
            Arranged sets of catalog inlays. Dealerships add a whole set at
            once; each inlay is ordered on its own and the set stays grouped
            in their stone design.
          </p>
        </div>
        <Button as={Link} to="/admin/compounds/create">
          <IoAddCircleOutline size={18} class="mr-1" />
          New compound
        </Button>
      </div>

      <Switch>
        <Match when={compoundsQuery.isLoading}>
          <div class="h-48 animate-pulse rounded-lg bg-gray-100" />
        </Match>
        <Match when={compoundsQuery.isError}>
          <p class="text-sm text-red-600">Couldn't load compounds.</p>
        </Match>
        <Match when={compoundsQuery.isSuccess && compoundsQuery.data.length === 0}>
          <p class="rounded-lg border border-dashed border-gray-300 p-8 text-center text-sm text-gray-500">
            No compounds yet.
          </p>
        </Match>
        <Match when={compoundsQuery.isSuccess}>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
            <For each={compoundsQuery.data}>
              {(compound) => (
                <div class="flex flex-col gap-3 rounded-lg border border-gray-200 p-4">
                  <CompoundPreview members={compound.members} class="h-40 w-full" />
                  <div>
                    <div class="flex items-center gap-2">
                      <h2 class="truncate font-medium text-gray-900">{compound.name}</h2>
                      <Show when={!compound.is_active}>
                        <span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">Disabled</span>
                      </Show>
                    </div>
                    <p class="text-xs text-gray-500">
                      {compound.members.map((m) => m.catalog_item.catalog_code).join(" · ")}
                    </p>
                    <Show when={compound.members.some((m) => !m.catalog_item.is_active)}>
                      <p class="mt-1 text-xs text-amber-700">
                        Hidden from dealerships: a member is disabled in the catalog.
                      </p>
                    </Show>
                  </div>
                  <div class="mt-auto flex gap-2">
                    <Button as={Link} to={`/admin/compounds/${compound.uuid}`} size="sm" variant="outline">
                      Edit
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => toggleActive(compound)}>
                      {compound.is_active ? "Disable" : "Enable"}
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => remove(compound)}>
                      Delete
                    </Button>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Match>
      </Switch>
    </div>
  );
}
