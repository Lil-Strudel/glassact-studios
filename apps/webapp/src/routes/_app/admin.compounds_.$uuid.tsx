import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { Match, Switch } from "solid-js";
import type { CatalogCompoundWrite } from "@glassact/data";
import { showToast } from "@glassact/ui";
import { getCompoundOpts, putCompoundOpts } from "../../queries/compound";
import { CompoundEditor } from "../../components/admin/compound-editor";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/compounds_/$uuid")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const compoundQuery = useQuery(() => ({
    ...getCompoundOpts(params().uuid),
    refetchOnWindowFocus: false,
  }));
  const putCompound = useMutation(putCompoundOpts);

  function save(body: CatalogCompoundWrite) {
    const compound = compoundQuery.data;
    if (!compound) return;
    putCompound.mutate(
      { uuid: compound.uuid, body: { ...body, version: compound.version } },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ["compound"] });
          navigate({ to: "/admin/compounds" });
        },
        onError: (error) =>
          showToast({
            title: "Failed to save compound",
            description:
              error instanceof Error && isApiError(error)
                ? (error.data?.error ?? "Unknown error")
                : "Unknown error",
            variant: "error",
          }),
      },
    );
  }

  return (
    <div class="flex flex-col gap-6">
      <h1 class="text-2xl font-bold text-gray-900">Edit compound</h1>
      <Switch>
        <Match when={compoundQuery.isLoading}>
          <div class="h-96 animate-pulse rounded-lg bg-gray-100" />
        </Match>
        <Match when={compoundQuery.isError}>
          <p class="text-sm text-red-600">Couldn't load this compound.</p>
        </Match>
        <Match when={compoundQuery.data}>
          {(compound) => (
            <CompoundEditor initial={compound()} isSaving={putCompound.isPending} onSave={save} />
          )}
        </Match>
      </Switch>
    </div>
  );
}
