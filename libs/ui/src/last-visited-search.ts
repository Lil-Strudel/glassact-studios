import type { AnyRouter } from "@tanstack/solid-router";

const searchByPathname = new Map<string, Record<string, unknown>>();

function normalizePathname(pathname: string): string {
  return pathname.length > 1 ? pathname.replace(/\/+$/, "") : pathname;
}

export function trackLastVisitedSearch(router: AnyRouter): () => void {
  return router.subscribe("onResolved", ({ toLocation }) => {
    searchByPathname.set(
      normalizePathname(toLocation.pathname),
      toLocation.search,
    );
  });
}

// The stored value was produced by the destination route's own search
// validation, so handing it back typed as that route's search input is safe.
export function lastVisitedSearch<TSearch = Record<string, unknown>>(
  pathname: string,
): TSearch | undefined {
  return searchByPathname.get(normalizePathname(pathname)) as
    | TSearch
    | undefined;
}
