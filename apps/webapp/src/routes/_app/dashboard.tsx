import { createFileRoute, Link } from "@tanstack/solid-router";
import { useQuery } from "@tanstack/solid-query";
import {
  Component,
  createMemo,
  For,
  type JSX,
  Match,
  Show,
  Switch,
} from "solid-js";
import type {
  DealershipDashboard,
  GET,
  InternalDashboard,
  Project,
  ProjectStatus,
  StatusCount,
} from "@glassact/data";
import { IoAlertCircleOutline } from "solid-icons/io";
import { PageHeader } from "@glassact/ui";
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
  "invoiced",
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
      <PageHeader eyebrow="Dashboard" title={`Welcome back, ${props.userName}`} />

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
    <div class="mt-8 flex flex-col gap-10">
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

      <StatLedger class="sm:grid-cols-3">
        <StatCard label="Active orders" value={activeOrderCount().toString()} />
        <StatCard
          label="Pending approval"
          value={props.data.pending_approval_count.toString()}
        />
        <StatCard
          label="Outstanding amount"
          value={formatCents(props.data.outstanding_invoice_amount_cents)}
        />
      </StatLedger>

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
      <PageHeader eyebrow="GlassAct Studios" title="Operations overview" />

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
    <div class="mt-8 flex flex-col gap-10">
      <div class="flex flex-col gap-3">
        <Show when={props.data.pending_proof_count > 0}>
          <AlertBanner
            to="/projects"
            text={`${props.data.pending_proof_count} ${
              props.data.pending_proof_count === 1 ? "proof" : "proofs"
            } awaiting design`}
            action="View projects"
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

      <StatLedger class="sm:grid-cols-3">
        <StatCard
          label="Pending proofs"
          value={props.data.pending_proof_count.toString()}
        />
        <StatCard
          label="Outstanding amount"
          value={formatCents(props.data.outstanding_invoice_amount_cents)}
        />
        <StatCard label="New orders" value={orderedCount().toString()} />
      </StatLedger>

      <Show when={props.data.manufacturing_step_counts.length > 0}>
        <section class="flex flex-col gap-4">
          <h2 class="text-2xl">Manufacturing pipeline</h2>
          <ol class="flex flex-wrap items-stretch gap-2">
            <For each={props.data.manufacturing_step_counts}>
              {(entry) => (
                <li class="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-2.5">
                  <span class="lozenge text-gilt" aria-hidden="true" />
                  <span class="text-sm capitalize text-muted-foreground">
                    {entry.step.replace(/-/g, " ")}
                  </span>
                  <span class="font-display text-lg tabular-nums text-foreground">
                    {entry.count}
                  </span>
                </li>
              )}
            </For>
          </ol>
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
      class="group flex items-center justify-between gap-4 rounded-lg border border-gilt/45 bg-card px-4 py-3 transition-colors hover:bg-accent/60"
    >
      <div class="flex items-center gap-3">
        <span class="lozenge size-2 text-gilt" aria-hidden="true" />
        <span class="text-sm font-medium text-foreground">{props.text}</span>
      </div>
      <span class="text-sm font-medium text-crimson-light group-hover:underline">
        {props.action} <span aria-hidden="true">&rarr;</span>
      </span>
    </Link>
  );
};

const StatLedger: Component<{ class?: string; children: JSX.Element }> = (
  props,
) => {
  return (
    <div
      class={`grid grid-cols-1 divide-y divide-border overflow-hidden rounded-lg border border-border bg-card sm:divide-x sm:divide-y-0 ${props.class ?? ""}`}
    >
      {props.children}
    </div>
  );
};

interface StatCardProps {
  label: string;
  value: string;
  detail?: string;
}

const StatCard: Component<StatCardProps> = (props) => {
  return (
    <div class="flex flex-col gap-2 px-5 py-5">
      <p class="eyebrow">{props.label}</p>
      <p class="font-display text-3xl tabular-nums text-foreground">
        {props.value}
      </p>
      <Show when={props.detail}>
        <p class="text-xs text-faint">{props.detail}</p>
      </Show>
    </div>
  );
};

interface StatusBreakdownProps {
  counts: StatusCount[];
}

const StatusBreakdown: Component<StatusBreakdownProps> = (props) => {
  return (
    <section class="flex flex-col gap-4">
      <h2 class="text-2xl">Projects by status</h2>
      <Show
        when={props.counts.length > 0}
        fallback={<p class="text-sm text-faint">No projects to display.</p>}
      >
        <div class="flex flex-wrap gap-2">
          <For each={props.counts}>
            {(entry) => (
              <div class="flex items-center gap-3 rounded-lg border border-border bg-card py-2 pl-2 pr-4">
                <ProjectStatusBadge status={entry.status as ProjectStatus} />
                <span class="font-display text-lg tabular-nums text-foreground">
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
    <section class="flex flex-col gap-4">
      <div class="flex items-baseline justify-between gap-4">
        <h2 class="text-2xl">{props.title}</h2>
        <Link
          to="/projects"
          class="text-sm text-crimson-light hover:underline"
        >
          All projects <span aria-hidden="true">&rarr;</span>
        </Link>
      </div>
      <Show
        when={props.projects.length > 0}
        fallback={<p class="text-sm text-faint">No recent projects.</p>}
      >
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
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
      class="group flex flex-col gap-3 rounded-lg border border-border bg-card p-4 transition-colors hover:border-input hover:bg-accent/40"
    >
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="truncate font-display text-lg text-foreground group-hover:text-crimson-light">
            {props.project.name}
          </p>
          <Show when={props.project.internal_reference}>
            <p class="truncate text-xs tabular-nums text-faint">
              {props.project.internal_reference}
            </p>
          </Show>
        </div>
        <ProjectStatusBadge status={props.project.status} class="shrink-0" />
      </div>
      <p class="text-xs tabular-nums text-faint">{displayDate()}</p>
    </Link>
  );
};

const DashboardSkeleton: Component = () => {
  return (
    <div class="mt-8 flex flex-col gap-8">
      <div class="h-28 animate-pulse rounded-lg border border-border bg-card" />
      <div class="h-6 w-48 animate-pulse rounded bg-muted" />
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <For each={[1, 2, 3]}>
          {() => (
            <div class="h-24 animate-pulse rounded-lg border border-border bg-card" />
          )}
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
    <div class="mt-8 flex flex-col items-center gap-2 rounded-lg border border-dashed border-error-foreground/40 bg-error p-8 text-center">
      <p class="flex items-center gap-2 font-medium text-error-foreground">
        <IoAlertCircleOutline size={18} aria-hidden="true" />
        Couldn't load the dashboard
      </p>
      <p class="text-sm text-muted-foreground">
        {props.message ?? "An unexpected error occurred."}
      </p>
    </div>
  );
};
