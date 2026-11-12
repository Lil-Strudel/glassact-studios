import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { Button, showToast } from "@glassact/ui";
import {
  IoCopyOutline,
  IoSwapHorizontal,
  IoSwapVertical,
} from "solid-icons/io";
import { Can } from "../Can";
import { patchInlayOpts } from "../../queries/inlay";
import { isApiError } from "../../utils/is-api-error";
import type { MirrorFlags } from "../../utils/inlay-mirror";
import {
  useDuplicateInlay,
  type DuplicateInlayHooks,
} from "./use-duplicate-inlay";

type MirrorAxis = "mirror_x" | "mirror_y";

interface InlayArrangeActionsProps {
  projectUuid: string;
  inlay: MirrorFlags & { uuid: string };
  duplicateHooks?: DuplicateInlayHooks;
  class?: string;
}

export function InlayArrangeActions(props: InlayArrangeActionsProps) {
  const queryClient = useQueryClient();
  const mirror = useMutation(patchInlayOpts);
  const duplicator = useDuplicateInlay(() => props.projectUuid);

  function toggleMirror(axis: MirrorAxis) {
    mirror.mutate(
      { uuid: props.inlay.uuid, body: { [axis]: !props.inlay[axis] } },
      {
        onSuccess: () =>
          Promise.all([
            queryClient.invalidateQueries({
              queryKey: ["project", props.projectUuid, "inlays"],
            }),
            queryClient.invalidateQueries({
              queryKey: ["inlay", props.inlay.uuid],
            }),
          ]),
        onError: (error) =>
          showToast({
            title: "Couldn't mirror the inlay",
            description: isApiError(error)
              ? (error.data.error ?? "Unknown error")
              : "Unknown error",
            variant: "error",
          }),
      },
    );
  }

  return (
    <div class={props.class ?? "flex flex-wrap items-center gap-2"}>
      <Can permission={PERMISSION_ACTIONS.MANAGE_PROJECT}>
        <Button
          size="sm"
          variant={props.inlay.mirror_x ? "secondary" : "outline"}
          aria-pressed={props.inlay.mirror_x}
          title="Flip the design so it faces the other way"
          disabled={mirror.isPending}
          onClick={() => toggleMirror("mirror_x")}
        >
          <IoSwapHorizontal size={16} class="mr-1" />
          Mirror ↔
        </Button>
        <Button
          size="sm"
          variant={props.inlay.mirror_y ? "secondary" : "outline"}
          aria-pressed={props.inlay.mirror_y}
          title="Flip the design top to bottom"
          disabled={mirror.isPending}
          onClick={() => toggleMirror("mirror_y")}
        >
          <IoSwapVertical size={16} class="mr-1" />
          Mirror ↕
        </Button>
      </Can>
      <Can permission={PERMISSION_ACTIONS.CREATE_PROJECT}>
        <Button
          size="sm"
          variant="outline"
          disabled={duplicator.isPending()}
          onClick={() =>
            void duplicator.duplicate(props.inlay.uuid, props.duplicateHooks)
          }
        >
          <IoCopyOutline size={16} class="mr-1" />
          {duplicator.isPending() ? "Duplicating..." : "Duplicate"}
        </Button>
      </Can>
    </div>
  );
}
