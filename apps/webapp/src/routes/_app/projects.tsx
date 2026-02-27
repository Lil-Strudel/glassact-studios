import { createFileRoute, Link } from "@tanstack/solid-router";
import type { GET, Project, ProjectStatus } from "@glassact/data";
import {
  Button,
  Breadcrumb,
  Badge,
  Card,
  CardHeader,
  CardTitle,
} from "@glassact/ui";
import { IoAddCircleOutline } from "solid-icons/io";
import { Component, createMemo, For, Match, Show, Switch } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { getProjectsOpts } from "../../queries/project";
import { useUserContext } from "../../providers/user";
import { Can } from "../../components/Can";
import { ProjectStatusBadge } from "../../components/project/status-badge";

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
    title: "Needs Action",
    description:
      "Projects with proofs awaiting approval or invoices waiting to be paid.",
    statuses: ["pending-approval", "approved", "invoiced"],
  },
  {
    title: "Pending",
    description: "Projects that are being prepared or designed.",
    statuses: ["draft", "designing"],
  },
  {
    title: "Active",
    description: "Projects currently in progress.",
    statuses: ["ordered", "in-production", "shipped", "delivered"],
  },
  {
    title: "Completed",
    description: "Projects that are finished or cancelled.",
    statuses: ["completed", "cancelled"],
  },
];

const INTERNAL_GROUPS: StatusGroup[] = [
  {
    title: "Needs Action",
    description: "Projects requiring design work or delivery confirmation.",
    statuses: ["designing", "delivered"],
  },
  {
    title: "Pending",
    description: "Projects awaiting approval or already approved.",
    statuses: ["pending-approval", "approved"],
  },
  {
    title: "Active",
    description: "Projects currently in progress.",
    statuses: ["ordered", "in-production", "shipped", "invoiced"],
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

  function getByStatuses(statuses: ProjectStatus[]): GET<Project>[] {
    if (!query.isSuccess) return [];
    return query.data.filter((project) => statuses.includes(project.status));
  }

  return (
    <div>
      <Breadcrumb crumbs={[{ title: "Projects", to: "/projects" }]} />
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          Projects
        </h1>
        <Can permission="create_project">
          <Button as={Link} to="/projects/create-project">
            Create New Project
            <IoAddCircleOutline size={20} class="ml-2" />
          </Button>
        </Can>
      </div>

      <Switch>
        <Match when={query.isLoading}>
          <div class="flex flex-col gap-12 mt-8">
            <For each={[1, 2, 3]}>
              {() => (
                <div>
                  <div class="h-7 w-48 bg-gray-200 rounded animate-pulse" />
                  <div class="h-4 w-72 bg-gray-100 rounded animate-pulse mt-2" />
                  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
                    <For each={[1, 2, 3]}>
                      {() => (
                        <div class="h-32 bg-gray-200 rounded-lg animate-pulse" />
                      )}
                    </For>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Match>

        <Match when={query.isError}>
          <div class="mt-8 border-2 border-dashed border-red-300 rounded-xl p-8 text-center">
            <p class="text-red-600 font-medium">Failed to load projects</p>
            <p class="text-gray-500 text-sm mt-1">
              {query.error?.message ?? "An unexpected error occurred."}
            </p>
            <Button
              variant="outline"
              class="mt-4"
              onClick={() => query.refetch()}
            >
              Retry
            </Button>
          </div>
        </Match>

        <Match when={query.isSuccess}>
          <div class="flex flex-col gap-12 mt-8">
            <For each={groups()}>
              {(group) => {
                const projects = () => getByStatuses(group.statuses);
                return (
                  <section>
                    <div class="flex items-center gap-3">
                      <h2 class="text-xl font-bold tracking-tight text-gray-900">
                        {group.title}
                      </h2>
                      <Badge variant="secondary">{projects().length}</Badge>
                    </div>
                    <p class="mt-1 text-sm text-gray-500">
                      {group.description}
                    </p>

                    <Show
                      when={projects().length > 0}
                      fallback={
                        <SectionMessage
                          title={`No ${group.title.toLowerCase()} projects`}
                          description="Projects in this category will appear here."
                        />
                      }
                    >
                      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
                        <For each={projects()}>
                          {(project) => <ProjectCard project={project} />}
                        </For>
                      </div>
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
    <div class="border-2 border-dashed border-gray-300 rounded-xl p-8 mt-4">
      <div class="text-center">
        <p class="text-gray-400 text-lg font-medium">{props.title}</p>
        <p class="text-gray-400 text-sm mt-2">{props.description}</p>
      </div>
    </div>
  );
};

interface ProjectCardProps {
  project: GET<Project>;
}

const ProjectCard: Component<ProjectCardProps> = (props) => {
  const formattedDate = () => {
    const date = new Date(props.project.created_at);
    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  return (
    <Card class="hover:shadow-md transition-shadow">
      <CardHeader class="flex flex-row items-start justify-between gap-4 space-y-0">
        <div class="flex flex-col gap-2 min-w-0">
          <CardTitle class="text-base truncate">{props.project.name}</CardTitle>
          <div class="flex items-center gap-2 flex-wrap">
            <ProjectStatusBadge status={props.project.status} />
          </div>
          <p class="text-xs text-gray-500">{formattedDate()}</p>
        </div>
        <Button
          as={Link}
          to={`/projects/${props.project.uuid}`}
          variant="outline"
          size="sm"
        >
          View
        </Button>
      </CardHeader>
    </Card>
  );
};
