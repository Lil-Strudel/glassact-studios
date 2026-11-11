import { createMemo } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { browseCatalogOpts } from "../../queries/catalog-browse";

// Every active background fits in one page; the browse endpoint caps at 100.
const BACKGROUND_PAGE = 100;

export function useBackgroundItems() {
  const query = useQuery(() =>
    browseCatalogOpts({ kind: "background", limit: BACKGROUND_PAGE }),
  );
  const byUuid = createMemo(
    () => new Map((query.data?.items ?? []).map((item) => [item.uuid, item])),
  );
  return { query, byUuid };
}
