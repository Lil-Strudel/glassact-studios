import { createEffect } from "solid-js";
import { createStore } from "solid-js/store";
import { useQueryClient } from "@tanstack/solid-query";
import type { StoneElement } from "@glassact/data";
import { getCatalogSvgTextOpts } from "../../queries/customize";
import { parseBackgroundArt, type BackgroundArt } from "./background-art";

// Fetches and parses the artwork of every background placed on the stone, so
// the canvas can resolve a catalog uuid to drawable markup synchronously.
export function useBackgroundArt(elements: () => StoneElement[]) {
  const queryClient = useQueryClient();
  const [art, setArt] = createStore<Record<string, BackgroundArt | null>>({});

  function ensure(uuid: string) {
    if (uuid in art) return;
    setArt(uuid, null);
    queryClient
      .fetchQuery(getCatalogSvgTextOpts(uuid))
      .then((text) => setArt(uuid, parseBackgroundArt(text)))
      .catch(() => setArt(uuid, null));
  }

  createEffect(() => {
    for (const el of elements()) {
      if (el.type === "background") ensure(el.catalog_item_uuid);
    }
  });

  return (uuid: string) => art[uuid] ?? undefined;
}
