import { Show, createMemo } from "solid-js";
import type { ManufacturingStep } from "@glassact/data";
import { LozengeStepper } from "../project/lozenge-stepper";
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
   * estimate under the steps; the compact card version leaves it off.
   */
  orderedAt?: string | null;
}

const STEP_TRACKER_LABELS = STEP_ORDER.map((step) => STEP_SHORT_LABELS[step]);

export function ManufacturingTracker(props: ManufacturingTrackerProps) {
  const currentIndex = createMemo(() => STEP_ORDER.indexOf(props.currentStep));
  const estimate = createMemo(() =>
    estimateProjectCompletion(props.orderedAt ?? null),
  );

  return (
    <div class="w-full space-y-1.5">
      <LozengeStepper
        labels={STEP_TRACKER_LABELS}
        currentIndex={currentIndex()}
      />
      <Show when={estimate()}>
        {(range) => (
          <p class="text-xs text-faint">
            Estimated ready to ship: {formatEstimate(range())}
          </p>
        )}
      </Show>
    </div>
  );
}
