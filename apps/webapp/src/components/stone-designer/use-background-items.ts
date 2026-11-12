import { createMemo } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { browseBackgroundsOpts } from "../../queries/background";

export function useBackgroundItems() {
  const query = useQuery(() => browseBackgroundsOpts());
  const byUuid = createMemo(
    () => new Map((query.data ?? []).map((background) => [background.uuid, background])),
  );
  return { query, byUuid };
}
