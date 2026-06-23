import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { Show, Switch, Match, createMemo } from "solid-js";
import { Card, CardContent, showToast } from "@glassact/ui";
import { getCatalogItemOpts } from "../../queries/catalog";
import { getCatalogSvgTextOpts } from "../../queries/customize";
import { getGlassColorsOpts } from "../../queries/glass-colors";
import { getGroutsOpts } from "../../queries/grouts";
import { postCatalogInlayOpts } from "../../queries/inlay";
import { Customizer } from "../../components/customizer/customizer";
import { isApiError } from "../../utils/is-api-error";
import type { BakeResult } from "@glassact/data";

export const Route = createFileRoute(
  "/_app/projects_/$id/add-inlay/customize/$catalogUuid",
)({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const itemQuery = useQuery(() => getCatalogItemOpts(params().catalogUuid));
  const svgQuery = useQuery(() => getCatalogSvgTextOpts(params().catalogUuid));
  const glassQuery = useQuery(() => getGlassColorsOpts());
  const groutsQuery = useQuery(() => getGroutsOpts());

  const postCatalogInlay = useMutation(postCatalogInlayOpts);

  const isLoading = createMemo(
    () =>
      itemQuery.isLoading ||
      svgQuery.isLoading ||
      glassQuery.isLoading ||
      groutsQuery.isLoading,
  );

  const hasRegions = createMemo(() => {
    const regions = itemQuery.data?.manifest?.glass_regions;
    return !!regions && Object.keys(regions).length > 0;
  });

  function handleBakeComplete(result: BakeResult) {
    const item = itemQuery.data;
    if (!item) return;

    postCatalogInlay.mutate(
      {
        projectUuid: params().id,
        body: {
          name: item.name,
          catalog_item_id: item.id,
          customization_notes: "",
          customization: {
            baked_design_asset_url: result.design_asset_url,
            scale_factor: result.scale_factor,
            width: result.width,
            height: result.height,
            color_overrides: result.color_overrides ?? {},
          },
        },
      },
      {
        onSuccess() {
          showToast({
            title: "Customized inlay added",
            description: `${item.name} has been added to the project for internal review.`,
            variant: "success",
          });
          queryClient.invalidateQueries({
            queryKey: ["project", params().id, "inlays"],
          });
          navigate({ to: `/projects/${params().id}` });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to add inlay",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  return (
    <Switch>
      <Match when={isLoading()}>
        <div class="flex h-[60vh] items-center justify-center text-gray-500">
          Loading customizer…
        </div>
      </Match>

      <Match when={itemQuery.isError}>
        <Notice
          title="Couldn't load this design"
          body="Please try again in a moment."
        />
      </Match>

      <Match when={itemQuery.data && !hasRegions()}>
        <Notice
          title="This design isn't available to customize yet"
          body="Its artwork couldn't be prepared for recoloring. Please contact GlassAct if you'd like it enabled."
        />
      </Match>

      <Match
        when={
          itemQuery.data &&
          svgQuery.data &&
          glassQuery.data &&
          groutsQuery.data
        }
      >
        <Show when={itemQuery.data}>
          {(item) => (
            <Customizer
              item={item()}
              svgText={svgQuery.data!}
              glassColors={glassQuery.data!}
              grouts={groutsQuery.data!}
              onBakeComplete={handleBakeComplete}
            />
          )}
        </Show>
      </Match>
    </Switch>
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
