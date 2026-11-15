import { For } from "solid-js";
import { cn } from "@glassact/ui";

interface LozengeStepperProps {
  labels: string[];
  currentIndex: number;
  class?: string;
}

export function LozengeStepper(props: LozengeStepperProps) {
  return (
    <ol
      class={cn(
        "grid auto-cols-[minmax(5rem,1fr)] grid-flow-col overflow-x-auto pb-1 pt-2",
        props.class,
      )}
    >
      <For each={props.labels}>
        {(label, index) => {
          const isComplete = () => index() < props.currentIndex;
          const isCurrent = () => index() === props.currentIndex;

          return (
            <li
              class="relative flex flex-col items-center gap-2 text-center"
              aria-current={isCurrent() ? "step" : undefined}
            >
              {index() > 0 && (
                <span
                  aria-hidden="true"
                  class={cn(
                    "absolute right-1/2 top-[7px] h-px w-full",
                    isComplete() || isCurrent() ? "bg-gilt" : "bg-input",
                  )}
                />
              )}
              <span
                aria-hidden="true"
                class={cn(
                  "relative z-10 size-3.5 rotate-45 border-2",
                  isComplete() && "border-gilt bg-gilt",
                  isCurrent() &&
                    "border-crimson-light bg-primary shadow-[0_0_0_4px_hsl(var(--crimson-light)/0.2)]",
                  !isComplete() && !isCurrent() && "border-input bg-background",
                )}
              />
              <span
                class={cn(
                  "px-0.5 text-xs leading-tight",
                  isComplete() && "text-muted-foreground",
                  isCurrent() && "font-semibold text-foreground",
                  !isComplete() && !isCurrent() && "text-faint",
                )}
              >
                {label}
              </span>
            </li>
          );
        }}
      </For>
    </ol>
  );
}
