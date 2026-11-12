import { Link, createFileRoute, stripSearchParams } from "@tanstack/solid-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { For, Match, Show, Switch, createMemo } from "solid-js";
import type { GET, SandblastBackground } from "@glassact/data";
import { Button, showToast } from "@glassact/ui";
import { IoAddCircleOutline } from "solid-icons/io";
import { getBackgroundsOpts, patchBackgroundOpts } from "../../queries/background";
import { useDebouncedSearchInput } from "../../hooks/use-debounced-search-input";
import { matchesBackgroundSearch } from "../../components/catalog/background-search";
import { listSearchDefaults, listSearchSchema, type ListSearchInput } from "../../utils/list-search";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/backgrounds")({
  validateSearch: listSearchSchema,
  search: { middlewares: [stripSearchParams(listSearchDefaults)] },
  component: RouteComponent,
});

function RouteComponent() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const queryClient = useQueryClient();
  const backgroundsQuery = useQuery(() => getBackgroundsOpts());
  const patchBackground = useMutation(patchBackgroundOpts);

  function updateSearch(patch: ListSearchInput) {
    navigate({ search: (prev) => ({ ...prev, ...patch }), replace: true, resetScroll: false });
  }

  const [searchText, setSearchText] = useDebouncedSearchInput(
    () => search().q,
    (q) => updateSearch({ q, page: 1 }),
  );

  const visible = createMemo(() =>
    (backgroundsQuery.data ?? []).filter((background) => matchesBackgroundSearch(background, search().q)),
  );

  function toggleActive(background: GET<SandblastBackground>) {
    patchBackground.mutate(
      { uuid: background.uuid, body: { is_active: !background.is_active, version: background.version } },
      {
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ["background"] }),
        onError: (error) =>
          showToast({
            title: "Failed to update background",
            description:
              error instanceof Error && isApiError(error) ? (error.data?.error ?? "Unknown error") : "Unknown error",
            variant: "error",
          }),
      },
    );
  }

  return (
    <div class="flex flex-col gap-6">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">Sandblast backgrounds</h1>
          <p class="mt-1 text-gray-600">
            Masks dealerships place on the stone face and download for their
            sandblaster. Related inlays are suggested alongside each one.
          </p>
        </div>
        <Button as={Link} to="/admin/backgrounds/create">
          <IoAddCircleOutline size={18} class="mr-1" />
          New background
        </Button>
      </div>

      <input
        placeholder="Search by name, code or tag..."
        value={searchText()}
        onInput={(e) => setSearchText(e.currentTarget.value)}
        class="w-full max-w-sm rounded-md border border-gray-300 px-3 py-2 text-sm"
      />

      <Switch>
        <Match when={backgroundsQuery.isLoading}>
          <div class="h-48 animate-pulse rounded-lg bg-gray-100" />
        </Match>
        <Match when={backgroundsQuery.isError}>
          <p class="text-sm text-red-600">Couldn't load backgrounds.</p>
        </Match>
        <Match when={backgroundsQuery.isSuccess && visible().length === 0}>
          <p class="rounded-lg border border-dashed border-gray-300 p-8 text-center text-sm text-gray-500">
            {search().q ? "No backgrounds match that search." : "No backgrounds yet."}
          </p>
        </Match>
        <Match when={backgroundsQuery.isSuccess}>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
            <For each={visible()}>
              {(background) => (
                <div class="flex flex-col gap-3 rounded-lg border border-gray-200 p-4">
                  <div class="flex h-32 items-center justify-center rounded bg-gray-50 p-2">
                    <img
                      src={background.svg_url}
                      alt={background.name}
                      loading="lazy"
                      class="max-h-full max-w-full object-contain brightness-0"
                    />
                  </div>
                  <div>
                    <div class="flex items-center gap-2">
                      <h2 class="truncate font-medium text-gray-900">{background.name}</h2>
                      <Show when={!background.is_active}>
                        <span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">Disabled</span>
                      </Show>
                    </div>
                    <p class="text-xs text-gray-500">
                      {background.code} · {background.related_items.length} related{" "}
                      {background.related_items.length === 1 ? "inlay" : "inlays"}
                    </p>
                  </div>
                  <div class="mt-auto flex gap-2">
                    <Button as={Link} to={`/admin/backgrounds/${background.uuid}`} size="sm" variant="outline">
                      Edit
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => toggleActive(background)}>
                      {background.is_active ? "Disable" : "Enable"}
                    </Button>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Match>
      </Switch>
    </div>
  );
}
