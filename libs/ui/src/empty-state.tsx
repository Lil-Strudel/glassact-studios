import { Show, type JSX } from "solid-js";
import { cn } from "./cn";

interface EmptyStateProps {
  title: string;
  description?: JSX.Element;
  action?: JSX.Element;
  class?: string;
}

export function EmptyState(props: EmptyStateProps) {
  return (
    <div
      class={cn(
        "flex flex-col items-center gap-3 px-6 py-12 text-center",
        props.class,
      )}
    >
      <div class="ornament-rule w-40" aria-hidden="true">
        <span class="lozenge" />
      </div>
      <p class="font-display text-xl text-foreground">{props.title}</p>
      <Show when={props.description}>
        <p class="max-w-prose text-sm text-muted-foreground">
          {props.description}
        </p>
      </Show>
      <Show when={props.action}>
        <div class="pt-1">{props.action}</div>
      </Show>
    </div>
  );
}
