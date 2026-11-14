import { Show, type JSX } from "solid-js";

interface PageHeaderProps {
  eyebrow?: string;
  title: JSX.Element;
  description?: JSX.Element;
  actions?: JSX.Element;
}

export function PageHeader(props: PageHeaderProps) {
  return (
    <header class="flex flex-wrap items-end justify-between gap-x-6 gap-y-4 border-b border-border pb-5">
      <div class="min-w-0 space-y-1.5">
        <Show when={props.eyebrow}>
          <p class="eyebrow">{props.eyebrow}</p>
        </Show>
        <h1 class="text-3xl sm:text-4xl">{props.title}</h1>
        <Show when={props.description}>
          <p class="max-w-prose text-muted-foreground">{props.description}</p>
        </Show>
      </div>
      <Show when={props.actions}>
        <div class="flex flex-wrap items-center gap-2">{props.actions}</div>
      </Show>
    </header>
  );
}
