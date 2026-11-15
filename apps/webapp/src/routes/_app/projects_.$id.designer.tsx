import { createFileRoute } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { Match, Show, Switch, createMemo } from "solid-js";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { Badge, Breadcrumb, Button, showToast } from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
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

  function refetchAll() {
    projectQuery.refetch();
    inlaysQuery.refetch();
    designQuery.refetch();
  }

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
    <div class="flex h-[var(--app-pane-height)] min-h-[640px] flex-col gap-4">
      <div class="flex flex-wrap items-center gap-3">
        <Breadcrumb
          crumbs={[
            { title: "Projects", to: "/projects" },
            {
              title: projectQuery.data?.name ?? "Project",
              to: `/projects/${params().id}`,
            },
            { title: "Stone design", to: `/projects/${params().id}/designer` },
          ]}
        />
        <Show when={projectQuery.isSuccess && !isEditable()}>
          <Badge variant="secondary" lozenge>
            {projectQuery.data?.status === "draft" ? "View only" : "Frozen with the order"}
          </Badge>
        </Show>
      </div>

      <Switch>
        <Match when={isLoading()}>
          <div class="flex-1 animate-pulse rounded-lg border border-border bg-card" />
        </Match>
        <Match when={isError()}>
          <div class="flex flex-col items-center gap-2 rounded-lg border border-dashed border-error-foreground/40 bg-error p-8 text-center">
            <p class="flex items-center gap-2 font-medium text-error-foreground">
              <IoAlertCircleOutline size={18} aria-hidden="true" />
              Couldn't load the stone design
            </p>
            <Button variant="secondary" class="mt-2" onClick={refetchAll}>
              Try again
            </Button>
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
