import { createFileRoute, Link } from "@tanstack/solid-router";
import { useQuery } from "@tanstack/solid-query";
import { Component, createMemo, For, Match, Show, Switch } from "solid-js";
import type {
  DealershipDashboard,
  GET,
  InternalDashboard,
  Project,
  ProjectStatus,
  StatusCount,
} from "@glassact/data";
import { Badge, Card, CardContent, CardHeader, CardTitle } from "@glassact/ui";
import {
  getDealershipDashboardOpts,
  getInternalDashboardOpts,
} from "../../queries/dashboard";
import { useUserContext } from "../../providers/user";
import { ProjectStatusBadge } from "../../components/project/status-badge";

export const Route = createFileRoute("/_app/dashboard")({
  component: RouteComponent,
});

const ACTIVE_ORDER_STATUSES: ProjectStatus[] = [
  "ordered",
  "in-production",
  "shipped",
  "delivered",
];

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
});

function formatCents(cents: number): string {
  return currencyFormatter.format(cents / 100);
}

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function sumStatusCounts(
  counts: StatusCount[],
  statuses: string[],
): number {
  return counts
    .filter((entry) => statuses.includes(entry.status))
    .reduce((sum, entry) => sum + entry.count, 0);
}

function RouteComponent() {
  const { user, isDealership } = useUserContext();

  const userName = createMemo(() => user()?.name ?? "");

  return (
    <Switch>
      <Match when={isDealership()}>
        <DealershipDashboardView userName={userName()} />
      </Match>
      <Match when={!isDealership()}>
        <InternalDashboardView />
      </Match>
    </Switch>
  );
}

interface DealershipDashboardViewProps {
  userName: string;
}

const DealershipDashboardView: Component<DealershipDashboardViewProps> = (
  props,
) => {
  const query = useQuery(() => getDealershipDashboardOpts());

  return (
    <div>
      <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
        Welcome back, {props.userName}
      </h1>

      <Switch>
        <Match when={query.isLoading}>
          <DashboardSkeleton />
        </Match>
        <Match when={query.isError}>
          <DashboardError message={query.error?.message} />
        </Match>
        <Match when={query.isSuccess && query.data}>
          {(data) => <DealershipDashboardBody data={data()} />}
        </Match>
      </Switch>
    </div>
  );
};

interface DealershipDashboardBodyProps {
  data: DealershipDashboard;
}

const DealershipDashboardBody: Component<DealershipDashboardBodyProps> = (
  props,
) => {
  const activeOrderCount = createMemo(() =>
    sumStatusCounts(props.data.project_status_counts, ACTIVE_ORDER_STATUSES),
  );

  return (
    <div class="flex flex-col gap-8 mt-8">
      <div class="flex flex-col gap-3">
        <Show when={props.data.pending_approval_count > 0}>
          <AlertBanner
            to="/projects"
            text={`${props.data.pending_approval_count} ${
              props.data.pending_approval_count === 1 ? "proof" : "proofs"
            } awaiting your approval`}
            action="Review"
          />
        </Show>
        <Show when={props.data.outstanding_invoice_count > 0}>
          <AlertBanner
            to="/projects"
            text={`${props.data.outstanding_invoice_count} outstanding ${
              props.data.outstanding_invoice_count === 1
                ? "invoice"
                : "invoices"
            }`}
            action="View"
          />
        </Show>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard label="Active Orders" value={activeOrderCount().toString()} />
        <StatCard
          label="Pending Approval"
          value={props.data.pending_approval_count.toString()}
        />
        <StatCard
          label="Outstanding Amount"
          value={formatCents(props.data.outstanding_invoice_amount_cents)}
        />
      </div>

      <StatusBreakdown counts={props.data.project_status_counts} />

      <RecentProjectsList
        title="Recent Projects"
        projects={props.data.recent_projects}
        dateField="updated_at"
      />
    </div>
  );
};

const InternalDashboardView: Component = () => {
  const query = useQuery(() => getInternalDashboardOpts());

  return (
    <div>
      <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
        Operations Overview
      </h1>

      <Switch>
        <Match when={query.isLoading}>
          <DashboardSkeleton />
        </Match>
        <Match when={query.isError}>
          <DashboardError message={query.error?.message} />
        </Match>
        <Match when={query.isSuccess && query.data}>
          {(data) => <InternalDashboardBody data={data()} />}
        </Match>
      </Switch>
    </div>
  );
};

interface InternalDashboardBodyProps {
  data: InternalDashboard;
}

const InternalDashboardBody: Component<InternalDashboardBodyProps> = (
  props,
) => {
  const orderedCount = createMemo(() =>
    sumStatusCounts(props.data.project_status_counts, ["ordered"]),
  );

  return (
    <div class="flex flex-col gap-8 mt-8">
      <div class="flex flex-col gap-3">
        <Show when={props.data.hard_blocker_count > 0}>
          <AlertBanner
            to="/inlays"
            text={`${props.data.hard_blocker_count} hard ${
              props.data.hard_blocker_count === 1 ? "blocker" : "blockers"
            } need attention`}
            action="View Kanban"
          />
        </Show>
        <Show when={props.data.pending_proof_count > 0}>
          <AlertBanner
            to="/projects"
            text={`${props.data.pending_proof_count} ${
              props.data.pending_proof_count === 1 ? "proof" : "proofs"
            } awaiting design`}
            action="View Projects"
          />
        </Show>
        <Show when={orderedCount() > 0}>
          <AlertBanner
            to="/inlays"
            text={`${orderedCount()} new ${
              orderedCount() === 1 ? "order" : "orders"
            } ready for production`}
            action="Kanban"
          />
        </Show>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label="Pending Proofs"
          value={props.data.pending_proof_count.toString()}
        />
        <StatCard
          label="Active Blockers"
          value={props.data.active_blocker_count.toString()}
          detail={`${props.data.active_blocker_count} total, ${props.data.hard_blocker_count} hard`}
        />
        <StatCard
          label="Outstanding Amount"
          value={formatCents(props.data.outstanding_invoice_amount_cents)}
        />
        <StatCard label="New Orders" value={orderedCount().toString()} />
      </div>

      <Show when={props.data.manufacturing_step_counts.length > 0}>
        <section>
          <h2 class="text-xl font-bold tracking-tight text-gray-900">
            Manufacturing Pipeline
          </h2>
          <div class="flex flex-wrap gap-3 mt-4">
            <For each={props.data.manufacturing_step_counts}>
              {(entry) => (
                <div class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2">
                  <span class="text-sm font-medium text-gray-700 capitalize">
                    {entry.step.replace(/-/g, " ")}
                  </span>
                  <Badge variant="secondary">{entry.count}</Badge>
                </div>
              )}
            </For>
          </div>
        </section>
      </Show>

      <StatusBreakdown counts={props.data.project_status_counts} />

      <RecentProjectsList
        title="Recent Orders"
        projects={props.data.recent_projects}
        dateField="ordered_at"
      />
    </div>
  );
};

interface AlertBannerProps {
  to: string;
  text: string;
  action: string;
}

const AlertBanner: Component<AlertBannerProps> = (props) => {
  return (
    <Link
      to={props.to}
      class="flex items-center justify-between gap-4 rounded-lg border border-amber-300 bg-amber-50 px-4 py-3 hover:bg-amber-100 transition-colors"
    >
      <div class="flex items-center gap-3">
        <span
          aria-hidden="true"
          class="flex h-8 w-8 items-center justify-center rounded-full bg-amber-200 text-amber-800 font-bold"
        >
          !
        </span>
        <span class="text-sm font-medium text-amber-900">{props.text}</span>
      </div>
      <span class="text-sm font-semibold text-amber-900">
        {props.action} →
      </span>
    </Link>
  );
};

interface StatCardProps {
  label: string;
  value: string;
  detail?: string;
}

const StatCard: Component<StatCardProps> = (props) => {
  return (
    <Card>
      <CardHeader>
        <CardTitle class="text-sm font-medium text-gray-500">
          {props.label}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-2xl font-bold text-gray-900">{props.value}</p>
        <Show when={props.detail}>
          <p class="text-xs text-gray-500 mt-1">{props.detail}</p>
        </Show>
      </CardContent>
    </Card>
  );
};

interface StatusBreakdownProps {
  counts: StatusCount[];
}

const StatusBreakdown: Component<StatusBreakdownProps> = (props) => {
  return (
    <section>
      <h2 class="text-xl font-bold tracking-tight text-gray-900">
        Project Status Breakdown
      </h2>
      <Show
        when={props.counts.length > 0}
        fallback={
          <p class="text-sm text-gray-500 mt-2">No projects to display.</p>
        }
      >
        <div class="flex flex-wrap gap-2 mt-4">
          <For each={props.counts}>
            {(entry) => (
              <div class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2">
                <ProjectStatusBadge status={entry.status as ProjectStatus} />
                <span class="text-sm font-semibold text-gray-900">
                  {entry.count}
                </span>
              </div>
            )}
          </For>
        </div>
      </Show>
    </section>
  );
};

interface RecentProjectsListProps {
  title: string;
  projects: GET<Project>[];
  dateField: "updated_at" | "ordered_at";
}

const RecentProjectsList: Component<RecentProjectsListProps> = (props) => {
  return (
    <section>
      <h2 class="text-xl font-bold tracking-tight text-gray-900">
        {props.title}
      </h2>
      <Show
        when={props.projects.length > 0}
        fallback={
          <p class="text-sm text-gray-500 mt-2">No recent projects.</p>
        }
      >
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
          <For each={props.projects}>
            {(project) => (
              <RecentProjectCard
                project={project}
                dateField={props.dateField}
              />
            )}
          </For>
        </div>
      </Show>
    </section>
  );
};

interface RecentProjectCardProps {
  project: GET<Project>;
  dateField: "updated_at" | "ordered_at";
}

const RecentProjectCard: Component<RecentProjectCardProps> = (props) => {
  const displayDate = createMemo(() => {
    const raw = props.project[props.dateField];
    return raw ? formatDate(raw) : "—";
  });

  return (
    <Link
      to="/projects/$id"
      params={{ id: props.project.uuid }}
      class="block hover:shadow-md transition-shadow"
    >
      <Card>
        <CardHeader>
          <CardTitle class="text-base truncate">
            {props.project.name}
          </CardTitle>
          <div class="flex items-center gap-2 flex-wrap mt-2">
            <ProjectStatusBadge status={props.project.status} />
          </div>
          <p class="text-xs text-gray-500 mt-2">{displayDate()}</p>
        </CardHeader>
      </Card>
    </Link>
  );
};

const DashboardSkeleton: Component = () => {
  return (
    <div class="flex flex-col gap-8 mt-8">
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <For each={[1, 2, 3]}>
          {() => <div class="h-28 bg-gray-200 rounded-xl animate-pulse" />}
        </For>
      </div>
      <div class="h-6 w-48 bg-gray-200 rounded animate-pulse" />
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <For each={[1, 2, 3]}>
          {() => <div class="h-32 bg-gray-200 rounded-xl animate-pulse" />}
        </For>
      </div>
    </div>
  );
};

interface DashboardErrorProps {
  message: string | undefined;
}

const DashboardError: Component<DashboardErrorProps> = (props) => {
  return (
    <div class="mt-8 border-2 border-dashed border-red-300 rounded-xl p-8 text-center">
      <p class="text-red-600 font-medium">Failed to load dashboard</p>
      <p class="text-gray-500 text-sm mt-1">
        {props.message ?? "An unexpected error occurred."}
      </p>
    </div>
  );
};
