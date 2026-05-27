import { createFileRoute } from "@tanstack/solid-router";
import { useQuery } from "@tanstack/solid-query";
import { Show, Switch, Match, createMemo } from "solid-js";
import { Card, CardContent } from "@glassact/ui";
import { getCatalogItemOpts } from "../../queries/catalog";
import { getCatalogSvgTextOpts } from "../../queries/customize";
import { getGlassColorsOpts } from "../../queries/glass-colors";
import { getGroutsOpts } from "../../queries/grouts";
import { Customizer } from "../../components/customizer/customizer";

export const Route = createFileRoute("/_app/catalog_/$uuid/customize")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();

  const itemQuery = useQuery(() => getCatalogItemOpts(params().uuid));
  const svgQuery = useQuery(() => getCatalogSvgTextOpts(params().uuid));
  const glassQuery = useQuery(() => getGlassColorsOpts());
  const groutsQuery = useQuery(() => getGroutsOpts());

  const isLoading = createMemo(
    () =>
      itemQuery.isLoading ||
      svgQuery.isLoading ||
      glassQuery.isLoading ||
      groutsQuery.isLoading,
  );

  const hasRegions = createMemo(() => {
    const regions = itemQuery.data?.manifest?.regions;
    return !!regions && Object.keys(regions).length > 0;
  });

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

      <Match when={itemQuery.data?.is_quarantined || !hasRegions()}>
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
