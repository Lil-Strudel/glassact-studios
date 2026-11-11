import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import type { CatalogCompoundWrite } from "@glassact/data";
import { showToast } from "@glassact/ui";
import { postCompoundOpts } from "../../queries/compound";
import { CompoundEditor } from "../../components/admin/compound-editor";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/compounds_/create")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const postCompound = useMutation(postCompoundOpts);

  function save(body: CatalogCompoundWrite) {
    postCompound.mutate(body, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["compound"] });
        navigate({ to: "/admin/compounds" });
      },
      onError: (error) =>
        showToast({
          title: "Failed to create compound",
          description:
            error instanceof Error && isApiError(error)
              ? (error.data?.error ?? "Unknown error")
              : "Unknown error",
          variant: "error",
        }),
    });
  }

  return (
    <div class="flex flex-col gap-6">
      <h1 class="text-2xl font-bold text-gray-900">New compound</h1>
      <CompoundEditor isSaving={postCompound.isPending} onSave={save} />
    </div>
  );
}
