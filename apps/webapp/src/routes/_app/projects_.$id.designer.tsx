import { Link, createFileRoute } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { Match, Show, Switch, createMemo } from "solid-js";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { Button, showToast } from "@glassact/ui";
import { IoArrowBackOutline } from "solid-icons/io";
import { getProjectOpts } from "../../queries/project";
import { getInlaysByProjectOpts } from "../../queries/inlay";
import { getStoneDesignOpts } from "../../queries/stone-design";
import { postProjectCompound } from "../../queries/compound";
import { useUserContext } from "../../providers/user";
import { isApiError } from "../../utils/is-api-error";
import { StoneDesigner } from "../../components/stone-designer/stone-designer";

export const Route = createFileRoute("/_app/projects_/$id/designer")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const queryClient = useQueryClient();
  const userContext = useUserContext();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlaysQuery = useQuery(() => getInlaysByProjectOpts(params().id));
  const designQuery = useQuery(() => ({
    ...getStoneDesignOpts(params().id),
    refetchOnWindowFocus: false,
  }));

  const isLoading = () =>
    projectQuery.isLoading || inlaysQuery.isLoading || designQuery.isLoading;
  const isError = () =>
    projectQuery.isError || inlaysQuery.isError || designQuery.isError;

  // Staff see the design but never edit it, and it freezes at order time.
  const isEditable = createMemo(
    () =>
      projectQuery.data?.status === "draft" &&
      userContext.isDealership() &&
      userContext.can(PERMISSION_ACTIONS.CREATE_PROJECT),
  );

  async function addCompound(compoundUuid: string) {
    try {
      await postProjectCompound({ projectUuid: params().id, compoundUuid });
      await queryClient.invalidateQueries({ queryKey: ["project", params().id, "inlays"] });
      showToast({
        title: "Compound added",
        description: "Its inlays are on the project and grouped on the stone.",
        variant: "success",
      });
    } catch (error) {
      showToast({
        title: "Failed to add compound",
        description:
          error instanceof Error && isApiError(error)
            ? (error.data?.error ?? "Unknown error")
            : "Unknown error",
        variant: "error",
      });
      throw error;
    }
  }

  return (
    <div class="flex h-[calc(100vh-8rem)] min-h-[640px] flex-col gap-4">
      <div class="flex flex-wrap items-center gap-3">
        <Button as={Link} to={`/projects/${params().id}`} variant="ghost" size="sm">
          <IoArrowBackOutline class="mr-1" size={16} />
          Back to project
        </Button>
        <h1 class="text-xl font-semibold text-gray-900">
          Stone design
          <Show when={projectQuery.data}>
            {(project) => <span class="font-normal text-gray-500"> · {project().name}</span>}
          </Show>
        </h1>
        <Show when={projectQuery.isSuccess && !isEditable()}>
          <span class="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs text-gray-600">
            {projectQuery.data?.status === "draft" ? "View only" : "Frozen with the order"}
          </span>
        </Show>
      </div>

      <Switch>
        <Match when={isLoading()}>
          <div class="flex-1 animate-pulse rounded-lg bg-gray-100" />
        </Match>
        <Match when={isError()}>
          <div class="rounded-lg border-2 border-dashed border-red-300 p-8 text-center text-sm text-red-700">
            Couldn't load the stone design.
          </div>
        </Match>
        <Match when={projectQuery.isSuccess && inlaysQuery.isSuccess && designQuery.isSuccess}>
          <StoneDesigner
            projectUuid={params().id}
            design={designQuery.data ?? null}
            inlays={inlaysQuery.data ?? []}
            isEditable={isEditable()}
            onAddCompound={addCompound}
          />
        </Match>
      </Switch>
    </div>
  );
}
