import type { ManufacturingStep } from "@glassact/data";

export const STEP_ORDER: ManufacturingStep[] = [
  "ordered",
  "materials-prep",
  "manufacturing",
  "packaging",
  "ready-to-ship",
];

export const STEP_LABELS: Record<ManufacturingStep, string> = {
  ordered: "Ordered",
  "materials-prep": "Prepping Materials",
  manufacturing: "Manufacturing",
  packaging: "Packaging",
  "ready-to-ship": "Ready to Ship",
};

// Abbreviated labels for the compact dot tracker, where horizontal space is tight.
export const STEP_SHORT_LABELS: Record<ManufacturingStep, string> = {
  ordered: "Ordered",
  "materials-prep": "Materials",
  manufacturing: "Manufacturing",
  packaging: "Packaging",
  "ready-to-ship": "Ready to Ship",
};

export function stepLabel(step: string): string {
  return STEP_LABELS[step as ManufacturingStep] ?? step;
}

/**
 * How long a whole order takes, from the day it is placed to the day it is
 * ready to ship. Business days — weekends are skipped when projecting dates.
 *
 * Deliberately a single number for the whole job rather than a per-step
 * breakdown: individual steps vary far too much to quote, and a dealership
 * only ever needs to answer "when will it get here?".
 */
export const PROJECT_DURATION = { minDays: 14, maxDays: 18 };

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

export interface ProjectEstimate {
  minDays: number;
  maxDays: number;
  minDate: Date;
  maxDate: Date;
}

/**
 * Projects when an order will be ready to ship, counting from the day it was
 * placed. Returns null for an order with no placed date — an unordered project
 * has no clock to run.
 */
export function estimateProjectCompletion(
  orderedAt: string | null,
): ProjectEstimate | null {
  if (!orderedAt) return null;

  const start = new Date(orderedAt);
  if (Number.isNaN(start.getTime())) return null;

  const { minDays, maxDays } = PROJECT_DURATION;
  return {
    minDays,
    maxDays,
    minDate: addBusinessDays(start, minDays),
    maxDate: addBusinessDays(start, maxDays),
  };
}

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  day: "numeric",
});

/**
 * Renders an estimate as a business-day span plus the dates it lands on,
 * e.g. "14–18 business days (May 20 – May 26)".
 */
export function formatEstimate(estimate: ProjectEstimate): string {
  const dates = `${dateFormatter.format(estimate.minDate)} – ${dateFormatter.format(estimate.maxDate)}`;
  const days =
    estimate.minDays === estimate.maxDays
      ? `${estimate.minDays} business day${estimate.minDays === 1 ? "" : "s"}`
      : `${estimate.minDays}–${estimate.maxDays} business days`;

  return `${days} (${dates})`;
}
