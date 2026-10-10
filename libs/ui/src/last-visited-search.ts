import type { AnyRouter } from "@tanstack/solid-router";

const searchByPathname = new Map<string, Record<string, unknown>>();

function normalizePathname(pathname: string): string {
  return pathname.length > 1 ? pathname.replace(/\/+$/, "") : pathname;
}

// onResolved only fires after a pending phase, which search-only changes on
// routes without loaders never enter; onBeforeLoad fires for every navigation.
export function trackLastVisitedSearch(router: AnyRouter): () => void {
  return router.subscribe("onBeforeLoad", ({ toLocation }) => {
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
