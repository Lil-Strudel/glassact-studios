import { Show } from "solid-js";
import { Link } from "@tanstack/solid-router";
import { IoAlertCircleOutline } from "solid-icons/io";
import type { GET, InlayUpdate } from "@glassact/data";
import { stepLabel } from "./steps";
import { formatTimestamp } from "../../utils/format-timestamp";

interface InlayIssueStripProps {
  latestIssue: GET<InlayUpdate>;
  issueCount: number;
  projectUuid: string;
  inlayUuid: string;
}

export function InlayIssueStrip(props: InlayIssueStripProps) {
  const earlierIssueCount = () => props.issueCount - 1;

  return (
    <div class="flex gap-3 border-t border-warning-foreground/30 bg-warning px-4 py-3">
      <IoAlertCircleOutline
        size={18}
        aria-hidden="true"
        class="mt-0.5 flex-shrink-0 text-warning-foreground"
      />
      <div class="flex min-w-0 flex-1 flex-col gap-1">
        <p class="eyebrow text-warning-foreground">Latest issue</p>
        <p class="break-words text-sm text-foreground">
          {props.latestIssue.message}
        </p>
        <p class="flex flex-wrap gap-x-2 text-xs text-faint">
          <Show when={props.latestIssue.step}>
            {(step) => <span>{stepLabel(step())}</span>}
          </Show>
          <span>{formatTimestamp(props.latestIssue.created_at)}</span>
          <Show when={earlierIssueCount() > 0}>
            <Link
              to="/projects/$id/inlay/$inlayId"
              params={{ id: props.projectUuid, inlayId: props.inlayUuid }}
              class="text-crimson-light hover:underline"
            >
              {earlierIssueCount() === 1
                ? "1 earlier issue"
                : `${earlierIssueCount()} earlier issues`}
            </Link>
          </Show>
        </p>
      </div>
    </div>
  );
}
