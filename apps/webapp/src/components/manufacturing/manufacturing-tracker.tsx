import { For, Show, createMemo } from "solid-js";
import { cn, Tooltip, TooltipContent, TooltipTrigger } from "@glassact/ui";
import type { ManufacturingStep } from "@glassact/data";
import {
  STEP_ORDER,
  STEP_SHORT_LABELS,
  estimateProjectCompletion,
  formatEstimate,
} from "./steps";

interface ManufacturingTrackerProps {
  currentStep: ManufacturingStep;
  /**
   * When the order was placed. Supplying it renders the whole-order completion
   * estimate under the dots; the compact card version leaves it off.
   */
  orderedAt?: string | null;
}

export function ManufacturingTracker(props: ManufacturingTrackerProps) {
  const currentIdx = createMemo(() => STEP_ORDER.indexOf(props.currentStep));
  const estimate = createMemo(() =>
    estimateProjectCompletion(props.orderedAt ?? null),
  );

  return (
    <div class="w-full space-y-1.5">
      <div class="flex items-center gap-0.5 w-full">
      <For each={STEP_ORDER}>
        {(step, index) => {
          const isComplete = () => index() < currentIdx();
          const isCurrent = () => index() === currentIdx();
          const isFuture = () => index() > currentIdx();

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
                    {isComplete()
                      ? "Completed"
                      : isCurrent()
                        ? "In progress"
                        : "Not started"}
                  </p>
                </TooltipContent>
              </Tooltip>
            </div>
          );
        }}
      </For>
      </div>

      <Show when={estimate()}>
        {(range) => (
          <p class="text-xs text-gray-500">
            Estimated ready to ship: {formatEstimate(range())}
          </p>
        )}
      </Show>
    </div>
  );
}
