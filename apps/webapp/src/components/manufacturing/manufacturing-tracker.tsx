import { For } from "solid-js";
import { cn } from "@glassact/ui";
import type { ManufacturingStep } from "@glassact/data";

interface ManufacturingTrackerProps {
  currentStep: ManufacturingStep;
}

const STEPS: { id: ManufacturingStep; label: string }[] = [
  { id: "ordered", label: "Ordered" },
  { id: "materials-prep", label: "Materials" },
  { id: "cutting", label: "Cutting" },
  { id: "fire-polish", label: "Polish" },
  { id: "packaging", label: "Packaging" },
  { id: "ready-to-ship", label: "Ready to Ship" },
];

const STEP_ORDER = STEPS.map((s) => s.id);

export function ManufacturingTracker(props: ManufacturingTrackerProps) {
  const currentIdx = () => STEP_ORDER.indexOf(props.currentStep);

  return (
    <div class="flex items-center gap-0.5 w-full">
      <For each={STEPS}>
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

              <div class="relative flex-shrink-0">
                <div
                  class={cn(
                    "w-3 h-3 rounded-full border-2",
                    isComplete() && "bg-primary border-primary",
                    isCurrent() &&
                      "bg-primary border-primary ring-2 ring-primary/30 ring-offset-1",
                    isFuture() && "bg-white border-gray-300",
                  )}
                  title={step.label}
                />
              </div>
            </div>
          );
        }}
      </For>
    </div>
  );
}
