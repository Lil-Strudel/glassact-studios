import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import type { SandblastBackgroundWrite } from "@glassact/data";
import { showToast } from "@glassact/ui";
import { getBackgroundsOpts, postBackgroundOpts } from "../../queries/background";
import { BackgroundForm } from "../../components/admin/background-form";
import { backgroundTagOptions } from "../../components/catalog/background-search";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/backgrounds_/create")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const backgroundsQuery = useQuery(() => getBackgroundsOpts());
  const postBackground = useMutation(postBackgroundOpts);

  function save(body: SandblastBackgroundWrite) {
    postBackground.mutate(body, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["background"] });
        navigate({ to: "/admin/backgrounds" });
      },
      onError: (error) =>
        showToast({
          title: "Failed to create background",
          description:
            error instanceof Error && isApiError(error) ? (error.data?.error ?? "Unknown error") : "Unknown error",
          variant: "error",
        }),
    });
  }

  return (
    <div class="flex flex-col gap-6">
      <h1 class="text-2xl font-bold text-gray-900">New background</h1>
      <BackgroundForm
        tagOptions={backgroundTagOptions(backgroundsQuery.data ?? [])}
        isSaving={postBackground.isPending}
        onSave={save}
      />
    </div>
  );
}
