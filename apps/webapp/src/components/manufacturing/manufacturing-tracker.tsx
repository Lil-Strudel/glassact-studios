import { For, Show, createMemo } from "solid-js";
import { cn, Tooltip, TooltipContent, TooltipTrigger } from "@glassact/ui";
import type { ManufacturingStep } from "@glassact/data";
import {
  STEP_ORDER,
  STEP_SHORT_LABELS,
  estimateCompletionRange,
  formatEstimate,
  formatStepDuration,
} from "./steps";

interface ManufacturingTrackerProps {
  currentStep: ManufacturingStep;
}

export function ManufacturingTracker(props: ManufacturingTrackerProps) {
  const currentIdx = createMemo(() => STEP_ORDER.indexOf(props.currentStep));

  return (
    <div class="flex items-center gap-0.5 w-full">
      <For each={STEP_ORDER}>
        {(step, index) => {
          const isComplete = () => index() < currentIdx();
          const isCurrent = () => index() === currentIdx();
          const isFuture = () => index() > currentIdx();

          // Cumulative estimate to finish this step, counting from the step the
          // inlay is on now. Past steps have already happened, so they get no
          // projection.
          const estimate = createMemo(() =>
            isComplete()
              ? null
              : estimateCompletionRange(props.currentStep, step),
          );

          return (
            <div class="flex items-center flex-1 min-w-0">
              {index() > 0 && (
                <div
                  class={cn(
                    "h-0.5 flex-1",
                    isComplete() || isCurrent() ? "bg-primary" : "bg-gray-200",
                  )}
                />
              )}

              <Tooltip>
                {/* `as="span"` because this tracker also renders inside a
                    project-card <Link>, and Kobalte's default <button> trigger
                    would nest an interactive element in an anchor. */}
                <TooltipTrigger
                  as="span"
                  class="relative flex-shrink-0 cursor-default p-1.5"
                  aria-label={STEP_SHORT_LABELS[step]}
                >
                  <span
                    class={cn(
                      "block w-3 h-3 rounded-full border-2",
                      isComplete() && "bg-primary border-primary",
                      isCurrent() &&
                        "bg-primary border-primary ring-2 ring-primary/30 ring-offset-1",
                      isFuture() && "bg-white border-gray-300",
                    )}
                  />
                </TooltipTrigger>
                <TooltipContent class="max-w-xs text-left">
                  <p class="font-medium">{STEP_SHORT_LABELS[step]}</p>
                  <p class="text-primary-foreground/80">
                    {formatStepDuration(step)}
                  </p>
                  <Show
                    when={estimate()}
                    fallback={
                      <p class="text-primary-foreground/80">Completed</p>
                    }
                  >
                    {(range) => (
                      <p class="text-primary-foreground/80">
                        Est. done in {formatEstimate(range())}
                      </p>
                    )}
                  </Show>
                </TooltipContent>
              </Tooltip>
            </div>
          );
        }}
      </For>
    </div>
  );
}
