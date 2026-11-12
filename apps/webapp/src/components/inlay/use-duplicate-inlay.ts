import { useMutation, useQueryClient } from "@tanstack/solid-query";
import type { Accessor } from "solid-js";
import type { GET, Inlay } from "@glassact/data";
import { showToast } from "@glassact/ui";
import { postDuplicateInlayOpts } from "../../queries/inlay";
import { isApiError } from "../../utils/is-api-error";

export interface DuplicateInlayHooks {
  before?: () => Promise<void>;
  after?: (duplicate: GET<Inlay>) => void | Promise<void>;
}

export function useDuplicateInlay(projectUuid: Accessor<string>) {
  const queryClient = useQueryClient();
  const mutation = useMutation(postDuplicateInlayOpts);

  async function duplicate(inlayUuid: string, hooks: DuplicateInlayHooks = {}) {
    try {
      await hooks.before?.();
      const created = await mutation.mutateAsync(inlayUuid);
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["project", projectUuid(), "inlays"],
        }),
        queryClient.invalidateQueries({
          queryKey: ["project", projectUuid(), "stone-design"],
        }),
      ]);
      await hooks.after?.(created);
      showToast({
        title: "Inlay duplicated",
        description: `Added ${created.name} to the project.`,
        variant: "success",
      });
    } catch (error) {
      showToast({
        title: "Couldn't duplicate the inlay",
        description:
          error instanceof Error && isApiError(error)
            ? (error.data.error ?? "Unknown error")
            : "Unknown error",
        variant: "error",
      });
    }
  }

  return { duplicate, isPending: () => mutation.isPending };
}
