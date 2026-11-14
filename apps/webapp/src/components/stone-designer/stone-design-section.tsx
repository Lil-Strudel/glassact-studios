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
    <section class="flex flex-col gap-4">
      <div class="flex items-center justify-between gap-4">
        <h2 class="text-2xl">Stone design</h2>
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
          <p class="rounded-lg border border-dashed border-input p-6 text-center text-sm text-muted-foreground">
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
              class="h-72 rounded-lg border border-border bg-muted transition-colors hover:border-gilt/50"
            />
          </Link>
        )}
      </Show>
    </section>
  );
}
