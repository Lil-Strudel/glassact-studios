import { For, Show, createMemo } from "solid-js";
import { Badge } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import type { ColorOverrides, Manifest } from "@glassact/data";
import { getGlassColorsOpts } from "../../queries/glass-colors";
import { getGroutsOpts } from "../../queries/grouts";
import {
  customPieceCount,
  groupGlassId,
} from "../customizer/shared/resolution";

interface InlayGlassSummaryProps {
  manifest: Manifest | undefined;
  colorOverrides: ColorOverrides;
}

// Read-only view of the glass and grout an inlay was customized with. The
// customizer records only a changelist (`color_overrides`), so every colour has
// to be resolved back against the design's manifest defaults to be shown.
export function InlayGlassSummary(props: InlayGlassSummaryProps) {
  const glassQuery = useQuery(() => getGlassColorsOpts());
  const groutsQuery = useQuery(() => getGroutsOpts());

  const glassById = createMemo(
    () => new Map((glassQuery.data ?? []).map((glass) => [glass.id, glass])),
  );

  const groupRows = createMemo(() =>
    Object.entries(props.manifest?.glass_regions ?? {}).map(
      ([groupKey, region]) => {
        const glassId = groupGlassId(groupKey, props.colorOverrides, props.manifest);
        return {
          groupKey,
          count: region.count,
          customCount: customPieceCount(region.piece_ids, props.colorOverrides),
          glass: glassId != null ? glassById().get(glassId) : undefined,
          fallbackHex: region.source_hex ?? "#cccccc",
        };
      },
    ),
  );

  const grout = createMemo(() => {
    const groutId =
      props.colorOverrides.background?.grout_id ??
      props.manifest?.grout_region.grout_id;
    if (groutId == null) return undefined;
    return (groutsQuery.data ?? []).find((g) => g.id === groutId);
  });

  return (
    <Show when={groupRows().length > 0}>
      <div class="space-y-2">
        <p class="text-xs font-medium uppercase tracking-wide text-gray-500">
          Glass
        </p>
        <ul class="space-y-1.5">
          <For each={groupRows()}>
            {(row) => (
              <li class="flex items-center gap-2">
                <span
                  class="h-5 w-5 shrink-0 rounded border border-black/10"
                  style={{
                    "background-color": row.glass?.hex ?? row.fallbackHex,
                  }}
                />
                <span class="min-w-0 flex-1 truncate text-sm text-gray-800">
                  {row.glass?.name ?? "Original color"}
                </span>
                <span class="shrink-0 text-xs text-gray-500">
                  {row.count} pc{row.count === 1 ? "" : "s"}
                </span>
                <Show when={row.customCount > 0}>
                  <Badge
                    variant="warning"
                    class="shrink-0 rounded-full px-1.5 py-0.5 text-[10px]"
                  >
                    {row.customCount} custom
                  </Badge>
                </Show>
              </li>
            )}
          </For>
          <Show when={grout()}>
            {(groutColor) => (
              <li class="flex items-center gap-2 border-t pt-1.5">
                <span
                  class="h-5 w-5 shrink-0 rounded border border-black/10"
                  style={{ "background-color": groutColor().hex }}
                />
                <span class="min-w-0 flex-1 truncate text-sm text-gray-800">
                  {groutColor().name}
                </span>
                <span class="shrink-0 text-xs text-gray-500">grout</span>
              </li>
            )}
          </Show>
        </ul>
      </div>
    </Show>
  );
}
