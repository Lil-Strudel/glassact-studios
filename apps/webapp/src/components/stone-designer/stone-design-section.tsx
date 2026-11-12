import { Show } from "solid-js";
import { Link } from "@tanstack/solid-router";
import { useQuery } from "@tanstack/solid-query";
import { PERMISSION_ACTIONS, type InlayWithInfo } from "@glassact/data";
import { Button } from "@glassact/ui";
import { getStoneDesignOpts } from "../../queries/stone-design";
import { useUserContext } from "../../providers/user";
import { StonePreview } from "./stone-preview";

interface StoneDesignSectionProps {
  projectUuid: string;
  inlays: InlayWithInfo[];
  isDraft: boolean;
}

export function StoneDesignSection(props: StoneDesignSectionProps) {
  const userContext = useUserContext();
  const designQuery = useQuery(() => getStoneDesignOpts(props.projectUuid));

  const canEdit = () =>
    props.isDraft &&
    userContext.isDealership() &&
    userContext.can(PERMISSION_ACTIONS.CREATE_PROJECT);

  return (
    <div class="mt-8">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-900">Stone design</h2>
        <Show when={designQuery.data || canEdit()}>
          <Button
            as={Link}
            to={`/projects/${props.projectUuid}/designer`}
            variant="outline"
            size="sm"
          >
            {canEdit() ? (designQuery.data ? "Open designer" : "Design the stone") : "View design"}
          </Button>
        </Show>
      </div>
      <Show
        when={designQuery.data}
        fallback={
          <p class="rounded-lg border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500">
            <Show when={canEdit()} fallback="No stone has been designed for this project.">
              See every inlay on the actual stone: pick a shape and granite,
              arrange the inlays, add a sandblasted background and lettering.
            </Show>
          </p>
        }
      >
        {(design) => (
          <Link to="/projects/$id/designer" params={{ id: props.projectUuid }} class="block">
            <StonePreview
              design={design()}
              inlays={props.inlays}
              class="h-72 rounded-lg border border-gray-200 bg-gray-50 transition hover:border-gray-400"
            />
          </Link>
        )}
      </Show>
    </div>
  );
}
