import { Show } from "solid-js";
import { Link } from "@tanstack/solid-router";
import { PERMISSION_ACTIONS, type InlayWithInfo } from "@glassact/data";
import { Button } from "@glassact/ui";
import { Can } from "../Can";

interface CustomizeInlayLinkProps {
  projectUuid: string;
  inlay: InlayWithInfo;
  variant?: "outline" | "link";
  class?: string;
}

export function CustomizeInlayLink(props: CustomizeInlayLinkProps) {
  return (
    <Show when={props.inlay.type === "catalog"}>
      <Can permission={PERMISSION_ACTIONS.MANAGE_PROJECT}>
        <Button
          as={Link}
          to="/projects/$id/inlay/$inlayId/recustomize"
          params={{ id: props.projectUuid, inlayId: props.inlay.uuid }}
          search={{ returnTo: "designer" }}
          size="sm"
          variant={props.variant ?? "outline"}
          class={props.class}
        >
          {props.inlay.is_customized ? "Adjust design" : "Customize"}
        </Button>
      </Can>
    </Show>
  );
}
