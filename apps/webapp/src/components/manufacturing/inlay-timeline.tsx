import { createMemo, createSignal, For, Match, Show, Switch } from "solid-js";
import { Badge, Button } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import {
  IoAlertCircleOutline,
  IoInformationCircleOutline,
} from "solid-icons/io";
import type { GET, InlayMilestone, InlayUpdate } from "@glassact/data";
import {
  getInlayMilestonesOpts,
  getInlayUpdatesOpts,
} from "../../queries/manufacturing";
import { stepLabel } from "./steps";
import { Can } from "../Can";
import { AddInlayUpdateForm } from "./add-inlay-update-form";

interface InlayTimelineProps {
  inlayUuid: string;
}

type TimelineItem =
  | { kind: "milestone"; time: string; milestone: GET<InlayMilestone> }
  | { kind: "update"; time: string; update: GET<InlayUpdate> };

function formatTime(time: string) {
  return new Date(time).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

function MilestoneRow(props: { milestone: GET<InlayMilestone> }) {
  return (
    <li class="flex gap-3">
      <div class="mt-1.5 flex-shrink-0">
        <span
          aria-hidden="true"
          class={`lozenge size-2 ${
            props.milestone.event_type === "reverted"
              ? "text-warning-foreground"
              : "text-gilt"
          }`}
        />
      </div>
      <div class="min-w-0 flex-1">
        <p class="text-sm text-foreground">
          <Show
            when={props.milestone.event_type === "reverted"}
            fallback={
              <>
                Entered{" "}
                <span class="font-medium">
                  {stepLabel(props.milestone.step)}
                </span>
              </>
            }
          >
            Reverted to{" "}
            <span class="font-medium">{stepLabel(props.milestone.step)}</span>
          </Show>
        </p>
        <p class="text-xs text-faint">
          {formatTime(props.milestone.event_time)}
        </p>
      </div>
    </li>
  );
}

function UpdateRow(props: { update: GET<InlayUpdate> }) {
  const isIssue = () => props.update.update_type === "issue";
  return (
    <li class="flex gap-3">
      <div class="flex-shrink-0 mt-0.5">
        <Show
          when={isIssue()}
          fallback={
            <span class="text-faint">
              <IoInformationCircleOutline size={18} />
            </span>
          }
        >
          <span class="text-warning-foreground">
            <IoAlertCircleOutline size={18} />
          </span>
        </Show>
      </div>
      <div
        class={`min-w-0 flex-1 rounded-lg border p-3 ${
          isIssue()
            ? "border-warning-foreground/30 bg-warning"
            : "border-border bg-secondary"
        }`}
      >
        <div class="flex items-center gap-2">
          <Badge
            variant={isIssue() ? "warning" : "outline"}
            class="capitalize"
          >
            {props.update.update_type}
          </Badge>
          <Show when={props.update.step}>
            <span class="text-xs text-faint">
              {stepLabel(props.update.step!)}
            </span>
          </Show>
        </div>
        <p class="mt-1 text-sm text-foreground break-words">
          {props.update.message}
        </p>
        <p class="mt-1 text-xs text-faint">
          {formatTime(props.update.created_at)}
        </p>
      </div>
    </li>
  );
}

export function InlayTimeline(props: InlayTimelineProps) {
  const [showAddForm, setShowAddForm] = createSignal(false);

  const milestonesQuery = useQuery(() =>
    getInlayMilestonesOpts(props.inlayUuid),
  );
  const updatesQuery = useQuery(() => getInlayUpdatesOpts(props.inlayUuid));

  const isLoading = () => milestonesQuery.isLoading || updatesQuery.isLoading;
  const isError = () => milestonesQuery.isError || updatesQuery.isError;

  const items = createMemo<TimelineItem[]>(() => {
    const milestones = (milestonesQuery.data ?? [])
      .filter((m) => m.event_type !== "exited")
      .map(
        (m): TimelineItem => ({
          kind: "milestone",
          time: m.event_time,
          milestone: m,
        }),
      );
    const updates = (updatesQuery.data ?? []).map(
      (u): TimelineItem => ({
        kind: "update",
        time: u.created_at,
        update: u,
      }),
    );
    return [...milestones, ...updates].sort(
      (a, b) => new Date(a.time).getTime() - new Date(b.time).getTime(),
    );
  });

  return (
    <div class="space-y-4 rounded-lg border border-border bg-card p-5">
      <div class="flex items-center justify-between">
        <h2 class="text-xl">Manufacturing timeline</h2>
      </div>

      <Switch>
        <Match when={isLoading()}>
          <div class="space-y-2">
            <For each={[1, 2, 3]}>
              {() => <div class="h-10 animate-pulse rounded bg-muted" />}
            </For>
          </div>
        </Match>

        <Match when={isError()}>
          <p class="flex items-center gap-2 text-sm text-error-foreground">
            <IoAlertCircleOutline size={16} aria-hidden="true" />
            Couldn't load the timeline.
          </p>
        </Match>

        <Match when={!isLoading() && !isError()}>
          <Show
            when={items().length > 0}
            fallback={
              <p class="py-4 text-center text-sm text-faint">
                No manufacturing activity yet.
              </p>
            }
          >
            <ol class="space-y-3">
              <For each={items()}>
                {(item) => (
                  <Switch>
                    <Match when={item.kind === "milestone" && item}>
                      {(m) => <MilestoneRow milestone={m().milestone} />}
                    </Match>
                    <Match when={item.kind === "update" && item}>
                      {(u) => <UpdateRow update={u().update} />}
                    </Match>
                  </Switch>
                )}
              </For>
            </ol>
          </Show>

          <Can permission="create_inlay_update">
            <Show
              when={showAddForm()}
              fallback={
                <Button
                  variant="secondary"
                  class="w-full"
                  onClick={() => setShowAddForm(true)}
                >
                  Add update
                </Button>
              }
            >
              <div class="space-y-3 rounded-lg border border-border bg-muted p-4">
                <p class="eyebrow">Add update</p>
                <AddInlayUpdateForm
                  inlayUuid={props.inlayUuid}
                  onSuccess={() => setShowAddForm(false)}
                  onCancel={() => setShowAddForm(false)}
                />
              </div>
            </Show>
          </Can>
        </Match>
      </Switch>
    </div>
  );
}
