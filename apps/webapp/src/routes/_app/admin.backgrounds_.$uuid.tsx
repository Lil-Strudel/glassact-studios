import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { Match, Switch } from "solid-js";
import type { SandblastBackgroundWrite } from "@glassact/data";
import { Alert, showToast } from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
import { PageHeader } from "@glassact/ui";
import {
  getBackgroundOpts,
  getBackgroundsOpts,
  putBackgroundOpts,
} from "../../queries/background";
import { BackgroundForm } from "../../components/admin/background-form";
import { backgroundTagOptions } from "../../components/catalog/background-search";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/backgrounds_/$uuid")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const backgroundQuery = useQuery(() => ({
    ...getBackgroundOpts(params().uuid),
    refetchOnWindowFocus: false,
  }));
  const backgroundsQuery = useQuery(() => getBackgroundsOpts());
  const putBackground = useMutation(putBackgroundOpts);

  function save(body: SandblastBackgroundWrite) {
    const background = backgroundQuery.data;
    if (!background) return;
    putBackground.mutate(
      { uuid: background.uuid, body: { ...body, version: background.version } },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ["background"] });
          navigate({ to: "/admin/backgrounds" });
        },
        onError: (error) =>
          showToast({
            title: "Failed to save background",
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
      <PageHeader eyebrow="Backgrounds" title="Edit background" />
      <Switch>
        <Match when={backgroundQuery.isLoading}>
          <div class="h-96 animate-pulse rounded-lg bg-card" />
        </Match>
        <Match when={backgroundQuery.isError}>
          <Alert variant="destructive">
            <IoAlertCircleOutline size={18} />
            <p>Couldn't load this background.</p>
          </Alert>
        </Match>
        <Match when={backgroundQuery.data}>
          {(background) => (
            <BackgroundForm
              initial={background()}
              tagOptions={backgroundTagOptions(backgroundsQuery.data ?? [])}
              isSaving={putBackground.isPending}
              onSave={save}
            />
          )}
        </Match>
      </Switch>
    </div>
  );
}
