import type { ManufacturingStep } from "@glassact/data";

export const STEP_ORDER: ManufacturingStep[] = [
  "ordered",
  "materials-prep",
  "cutting",
  "fire-polish",
  "packaging",
  "ready-to-ship",
];

export const STEP_LABELS: Record<ManufacturingStep, string> = {
  ordered: "Ordered",
  "materials-prep": "Prepping Materials",
  cutting: "Cutting",
  "fire-polish": "Fire Polish",
  packaging: "Packaging",
  "ready-to-ship": "Ready to Ship",
};

// Abbreviated labels for the compact dot tracker, where horizontal space is tight.
export const STEP_SHORT_LABELS: Record<ManufacturingStep, string> = {
  ordered: "Ordered",
  "materials-prep": "Materials",
  cutting: "Cutting",
  "fire-polish": "Polish",
  packaging: "Packaging",
  "ready-to-ship": "Ready to Ship",
};

export function stepLabel(step: string): string {
  return STEP_LABELS[step as ManufacturingStep] ?? step;
}

/**
 * PLACEHOLDER TIMELINES — these are not measured, they are guesses.
 *
 * This map is the single source for every duration and date estimate shown to
 * dealerships. Replace these numbers with the real ones and every tooltip,
 * duration phrase and projected date range updates with them. Nothing else
 * needs editing.
 *
 * Values are business days (weekends are skipped when projecting dates).
 */
export const STEP_DURATIONS: Record<
  ManufacturingStep,
  { minDays: number; maxDays: number }
> = {
  ordered: { minDays: 1, maxDays: 2 },
  "materials-prep": { minDays: 2, maxDays: 4 },
  cutting: { minDays: 3, maxDays: 5 },
  "fire-polish": { minDays: 2, maxDays: 3 },
  packaging: { minDays: 1, maxDays: 2 },
  "ready-to-ship": { minDays: 1, maxDays: 2 },
};

export function addBusinessDays(from: Date, days: number): Date {
  const result = new Date(from.getTime());
  let remaining = days;

  while (remaining > 0) {
    result.setDate(result.getDate() + 1);
    const day = result.getDay();
    if (day !== 0 && day !== 6) {
      remaining -= 1;
    }
  }

  return result;
}

export interface StepEstimate {
  minDays: number;
  maxDays: number;
  minDate: Date;
  maxDate: Date;
}

/**
 * Projects how long it takes to finish `throughStep`, counting from the start of
 * `fromStep`. Anchored at `today` rather than at the milestone entry time — the
 * tracker only knows the current step, and "from now" is the honest reading for
 * someone answering a customer's question.
 */
export function estimateCompletionRange(
  fromStep: ManufacturingStep,
  throughStep: ManufacturingStep,
  today: Date = new Date(),
): StepEstimate | null {
  const start = STEP_ORDER.indexOf(fromStep);
  const end = STEP_ORDER.indexOf(throughStep);
  if (start === -1 || end === -1 || end < start) return null;

  let minDays = 0;
  let maxDays = 0;
  for (const step of STEP_ORDER.slice(start, end + 1)) {
    minDays += STEP_DURATIONS[step].minDays;
    maxDays += STEP_DURATIONS[step].maxDays;
  }

  return {
    minDays,
    maxDays,
    minDate: addBusinessDays(today, minDays),
    maxDate: addBusinessDays(today, maxDays),
  };
}

export function formatStepDuration(step: ManufacturingStep): string {
  const { minDays, maxDays } = STEP_DURATIONS[step];
  if (minDays === maxDays) {
    return `typically ${minDays} business day${minDays === 1 ? "" : "s"}`;
  }
  return `typically ${minDays}–${maxDays} business days`;
}

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
});

/**
 * Renders an estimate as a rough span plus the concrete dates it lands on,
 * e.g. "2–3 weeks (May 20 – Jun 3)". Short estimates read better in days.
 */
export function formatEstimate(estimate: StepEstimate): string {
  const dates = `${dateFormatter.format(estimate.minDate)} – ${dateFormatter.format(estimate.maxDate)}`;

  if (estimate.maxDays < 10) {
    const days =
      estimate.minDays === estimate.maxDays
        ? `${estimate.minDays} business day${estimate.minDays === 1 ? "" : "s"}`
        : `${estimate.minDays}–${estimate.maxDays} business days`;
    return `${days} (${dates})`;
  }

  const minWeeks = Math.max(1, Math.round(estimate.minDays / 5));
  const maxWeeks = Math.max(minWeeks, Math.round(estimate.maxDays / 5));
  const weeks =
    minWeeks === maxWeeks
      ? `${minWeeks} week${minWeeks === 1 ? "" : "s"}`
      : `${minWeeks}–${maxWeeks} weeks`;

  return `${weeks} (${dates})`;
}
