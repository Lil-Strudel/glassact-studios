import { Show } from "solid-js";
import { Badge } from "@glassact/ui";
import { IoSwapHorizontal, IoSwapVertical } from "solid-icons/io";
import { mirrorDescription, type MirrorFlags } from "../../utils/inlay-mirror";

interface MirrorBadgeProps {
  inlay: MirrorFlags;
}

export function MirrorBadge(props: MirrorBadgeProps) {
  return (
    <Show when={mirrorDescription(props.inlay)}>
      {(description) => (
        <Badge variant="warning" title={description()}>
          <Show when={props.inlay.mirror_x}>
            <IoSwapHorizontal size={12} class="mr-1" />
          </Show>
          <Show when={props.inlay.mirror_y}>
            <IoSwapVertical size={12} class="mr-1" />
          </Show>
          {description()}
        </Badge>
      )}
    </Show>
  );
}
