import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { Match, Switch } from "solid-js";
import type { CatalogCompoundWrite } from "@glassact/data";
import { Alert, showToast } from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
import { PageHeader } from "@glassact/ui";
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
      <PageHeader eyebrow="Compounds" title="Edit compound" />
      <Switch>
        <Match when={compoundQuery.isLoading}>
          <div class="h-96 animate-pulse rounded-lg bg-card" />
        </Match>
        <Match when={compoundQuery.isError}>
          <Alert variant="destructive">
            <IoAlertCircleOutline size={18} />
            <p>Couldn't load this compound.</p>
          </Alert>
        </Match>
        <Match when={compoundQuery.data}>
          {(compound) => (
            <CompoundEditor
              initial={compound()}
              isSaving={putCompound.isPending}
              onSave={save}
            />
          )}
        </Match>
      </Switch>
    </div>
  );
}
