import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { Show, Switch, Match, createMemo } from "solid-js";
import { Breadcrumb, Card, CardContent, showToast } from "@glassact/ui";
import type { BakeResult, ColorOverrides } from "@glassact/data";
import { getCatalogItemOpts } from "../../queries/catalog";
import { getCatalogSvgTextOpts } from "../../queries/customize";
import { getGlassColorsOpts } from "../../queries/glass-colors";
import { getGroutsOpts } from "../../queries/grouts";
import { getInlayOpts, postRecustomizeInlayOpts } from "../../queries/inlay";
import { Customizer } from "../../components/customizer/customizer";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute(
  "/_app/projects_/$id/inlay/$inlayId/recustomize",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const inlayQuery = useQuery(() => getInlayOpts(params().inlayId));

  const catalogItemUuid = createMemo(
    () => inlayQuery.data?.catalog_item?.uuid ?? "",
  );

  const itemQuery = useQuery(() => ({
    ...getCatalogItemOpts(catalogItemUuid()),
    enabled: catalogItemUuid() !== "",
  }));
  const svgQuery = useQuery(() => ({
    ...getCatalogSvgTextOpts(catalogItemUuid()),
    enabled: catalogItemUuid() !== "",
  }));
  const glassQuery = useQuery(() => getGlassColorsOpts());
  const groutsQuery = useQuery(() => getGroutsOpts());

  const recustomize = useMutation(postRecustomizeInlayOpts);

  const isLoading = createMemo(
    () =>
      inlayQuery.isLoading ||
      itemQuery.isLoading ||
      svgQuery.isLoading ||
      glassQuery.isLoading ||
      groutsQuery.isLoading,
  );

  const hasRegions = createMemo(() => {
    const regions = itemQuery.data?.manifest?.glass_regions;
    return !!regions && Object.keys(regions).length > 0;
  });

  const isRecustomizable = createMemo(() => {
    const inlay = inlayQuery.data;
    return !!inlay && inlay.type === "catalog" && inlay.is_customized;
  });

  // Resume from the inlay's current coloring rather than the catalog defaults —
  // the point of adjusting is to tweak what's already there.
  const initialState = createMemo(() => {
    const proof = inlayQuery.data?.latest_proof ?? inlayQuery.data?.approved_proof;
    if (!proof) return undefined;
    return {
      overrides: (proof.color_overrides ?? {}) as ColorOverrides,
      width: proof.width,
    };
  });

  function handleBakeComplete(result: BakeResult) {
    recustomize.mutate(
      {
        uuid: params().inlayId,
        body: {
          baked_design_asset_url: result.design_asset_url,
          scale_factor: result.scale_factor,
          width: result.width,
          height: result.height,
          color_overrides: result.color_overrides ?? {},
        },
      },
      {
        onSuccess() {
          showToast({
            title: "Design updated",
            description:
              "The new coloring has been sent for internal pricing review.",
            variant: "success",
          });
          queryClient.invalidateQueries({
            queryKey: ["inlay", params().inlayId],
          });
          queryClient.invalidateQueries({
            queryKey: ["project", params().id, "inlays"],
          });
          queryClient.invalidateQueries({ queryKey: ["review-queue"] });
          navigate({
            to: "/projects/$id/inlay/$inlayId",
            params: { id: params().id, inlayId: params().inlayId },
          });
        },
        onError(error) {
          showToast({
            title: "Failed to update design",
            description: isApiError(error)
              ? (error.data?.error ?? "Unknown error")
              : "Unknown error",
            variant: "error",
          });
        },
      },
    );
  }

  return (
    <div class="space-y-6">
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          {
            title: "Project",
            to: `/projects/${params().id}`,
          },
          {
            title: inlayQuery.data?.name ?? "Inlay",
            to: `/projects/${params().id}/inlay/${params().inlayId}`,
          },
          {
            title: "Adjust design",
            to: `/projects/${params().id}/inlay/${params().inlayId}/recustomize`,
          },
        ]}
      />

      <Switch>
        <Match when={isLoading()}>
          <div class="flex h-[60vh] items-center justify-center text-gray-500">
            Loading customizer…
          </div>
        </Match>

        <Match when={inlayQuery.isError || itemQuery.isError}>
          <Notice
            title="Couldn't load this design"
            body="Please try again in a moment."
          />
        </Match>

        <Match when={!isRecustomizable()}>
          <Notice
            title="This inlay can't be adjusted"
            body="Only customized catalog inlays can be re-colored. Custom inlays go through a designer proof instead."
          />
        </Match>

        <Match when={itemQuery.data && !hasRegions()}>
          <Notice
            title="This design isn't available to customize"
            body="Its artwork couldn't be prepared for recoloring. Please contact GlassAct if you'd like it enabled."
          />
        </Match>

        <Match
          when={
            itemQuery.data && svgQuery.data && glassQuery.data && groutsQuery.data
          }
        >
          <Show when={itemQuery.data}>
            {(item) => (
              <Customizer
                item={item()}
                svgText={svgQuery.data!}
                glassColors={glassQuery.data!}
                grouts={groutsQuery.data!}
                initialState={initialState()}
                storageScope={`inlay:${params().inlayId}`}
                onBakeComplete={handleBakeComplete}
              />
            )}
          </Show>
        </Match>
      </Switch>
    </div>
  );
}

function Notice(props: { title: string; body: string }) {
  return (
    <Card class="mx-auto mt-12 max-w-md text-center">
      <CardContent class="pt-6">
        <h2 class="text-lg font-semibold text-gray-900">{props.title}</h2>
        <p class="mt-2 text-sm text-gray-600">{props.body}</p>
      </CardContent>
    </Card>
  );
}
