import { createFileRoute, Link } from "@tanstack/solid-router";
import type { ProjectListItem, ProjectStatus } from "@glassact/data";
import { Button, Badge } from "@glassact/ui";
import {
  IoAddCircleOutline,
  IoAlertCircleOutline,
  IoChevronForward,
} from "solid-icons/io";
import { Component, createMemo, For, Match, Show, Switch } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { getProjectsOpts } from "../../queries/project";
import { useUserContext } from "../../providers/user";
import { Can } from "../../components/Can";
import { ProjectStatusBadge } from "../../components/project/status-badge";
import { PageHeader } from "@glassact/ui";

export const Route = createFileRoute("/_app/projects")({
  component: RouteComponent,
});

interface StatusGroup {
  title: string;
  description: string;
  statuses: ProjectStatus[];
}

const DEALERSHIP_GROUPS: StatusGroup[] = [
  {
    title: "Drafts",
    description: "Projects you're still preparing.",
    statuses: ["draft"],
  },
  {
    title: "Ordered",
    description: "Orders placed and awaiting production.",
    statuses: ["ordered"],
  },
  {
    title: "In Production",
    description: "Projects currently being manufactured or shipped.",
    statuses: ["in-production", "shipped", "invoiced"],
  },
  {
    title: "Completed",
    description: "Projects that are finished or cancelled.",
    statuses: ["completed", "cancelled"],
  },
];

const INTERNAL_GROUPS: StatusGroup[] = [
  {
    title: "Incoming",
    description:
      "New projects from dealerships that are still being prepared.",
    statuses: ["draft"],
  },
  {
    title: "Ordered",
    description: "Orders placed and awaiting production.",
    statuses: ["ordered"],
  },
  {
    title: "In Production",
    description: "Projects currently being manufactured or shipped.",
    statuses: ["in-production", "shipped", "invoiced"],
  },
  {
    title: "Completed",
    description: "Projects that are finished or cancelled.",
    statuses: ["completed", "cancelled"],
  },
];

function RouteComponent() {
  const { isDealership } = useUserContext();
  const query = useQuery(() => getProjectsOpts());

  const groups = createMemo(() =>
    isDealership() ? DEALERSHIP_GROUPS : INTERNAL_GROUPS,
  );

  function getByStatuses(statuses: ProjectStatus[]): ProjectListItem[] {
    if (!query.isSuccess) return [];
    return query.data.filter((project) => statuses.includes(project.status));
  }

  return (
    <div>
      <PageHeader
        eyebrow={isDealership() ? "Your orders" : "All dealerships"}
        title="Projects"
        actions={
          <Can permission="create_project">
            <Button as={Link} to="/projects/create-project">
              <IoAddCircleOutline size={18} class="mr-2" />
              Create new project
            </Button>
          </Can>
        }
      />

      <Switch>
        <Match when={query.isLoading}>
          <div class="mt-8 flex flex-col gap-10">
            <For each={[1, 2, 3]}>
              {() => (
                <div class="flex flex-col gap-3">
                  <div class="h-7 w-48 animate-pulse rounded bg-muted" />
                  <div class="h-28 animate-pulse rounded-lg border border-border bg-card" />
                </div>
              )}
            </For>
          </div>
        </Match>

        <Match when={query.isError}>
          <div class="mt-8 flex flex-col items-center gap-2 rounded-lg border border-dashed border-error-foreground/40 bg-error p-8 text-center">
            <p class="flex items-center gap-2 font-medium text-error-foreground">
              <IoAlertCircleOutline size={18} aria-hidden="true" />
              Couldn't load projects
            </p>
            <p class="text-sm text-muted-foreground">
              {query.error?.message ?? "An unexpected error occurred."}
            </p>
            <Button
              variant="secondary"
              class="mt-2"
              onClick={() => query.refetch()}
            >
              Retry
            </Button>
          </div>
        </Match>

        <Match when={query.isSuccess}>
          <div class="mt-8 flex flex-col gap-10">
            <For each={groups()}>
              {(group) => {
                const projects = () => getByStatuses(group.statuses);
                return (
                  <section class="flex flex-col gap-3">
                    <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
                      <h2 class="text-2xl">{group.title}</h2>
                      <span class="text-sm tabular-nums text-faint">
                        {projects().length}
                      </span>
                      <p class="w-full text-sm text-muted-foreground">
                        {group.description}
                      </p>
                    </div>

                    <Show
                      when={projects().length > 0}
                      fallback={
                        <SectionMessage
                          title={`No ${group.title.toLowerCase()} projects`}
                          description="Projects in this category will appear here."
                        />
                      }
                    >
                      <ul class="divide-y divide-border overflow-hidden rounded-lg border border-border bg-card">
                        <For each={projects()}>
                          {(project) => (
                            <li>
                              <ProjectRow project={project} />
                            </li>
                          )}
                        </For>
                      </ul>
                    </Show>
                  </section>
                );
              }}
            </For>
          </div>
        </Match>
      </Switch>
    </div>
  );
}

interface SectionMessageProps {
  title: string;
  description: string;
}

const SectionMessage: Component<SectionMessageProps> = (props) => {
  return (
    <div class="rounded-lg border border-dashed border-input px-6 py-6 text-center">
      <p class="text-muted-foreground">{props.title}</p>
      <p class="mt-1 text-sm text-faint">{props.description}</p>
    </div>
  );
};

interface ProjectRowProps {
  project: ProjectListItem;
}

const ProjectRow: Component<ProjectRowProps> = (props) => {
  const formattedDate = () => {
    const date = new Date(props.project.created_at);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  const summary = () => props.project.action_summary;

  return (
    <Link
      to="/projects/$id"
      params={{ id: props.project.uuid }}
      class="group grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-2 px-4 py-3.5 transition-colors hover:bg-accent/60 sm:grid-cols-[minmax(0,1fr)_auto_auto]"
    >
      <div class="flex min-w-0 flex-col gap-1">
        <div class="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-0.5">
          <span class="truncate font-display text-lg text-foreground group-hover:text-crimson-light">
            {props.project.name}
          </span>
          <Show when={props.project.internal_reference}>
            {(ref) => (
              <span class="truncate text-sm tabular-nums text-muted-foreground">
                {ref()}
              </span>
            )}
          </Show>
        </div>
        <div class="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-faint">
          <Show when={props.project.dealership_name}>
            {(name) => (
              <>
                <span class="text-muted-foreground">{name()}</span>
                <span aria-hidden="true">·</span>
              </>
            )}
          </Show>
          <span class="tabular-nums">Created {formattedDate()}</span>
        </div>
        <Show when={summary()}>
          {(s) => (
            <Show
              when={
                s().needs_internal_approval > 0 ||
                s().needs_proof > 0 ||
                s().awaiting_reply
              }
            >
              <div class="flex flex-wrap items-center gap-1.5 pt-1">
                <Show when={s().needs_internal_approval > 0}>
                  <Badge variant="gilt" lozenge>
                    Needs approval · {s().needs_internal_approval}
                  </Badge>
                </Show>
                <Show when={s().needs_proof > 0}>
                  <Badge variant="info" lozenge>
                    Needs proof · {s().needs_proof}
                  </Badge>
                </Show>
                <Show when={s().awaiting_reply}>
                  <Badge variant="crimson" lozenge>
                    Awaiting reply
                  </Badge>
                </Show>
              </div>
            </Show>
          )}
        </Show>
      </div>
      <ProjectStatusBadge status={props.project.status} class="justify-self-end" />
      <IoChevronForward
        size={16}
        class="hidden text-faint transition-colors group-hover:text-crimson-light sm:block"
        aria-hidden="true"
      />
    </Link>
  );
};
