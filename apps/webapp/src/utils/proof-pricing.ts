import type { PriceAdjustmentType } from "@glassact/data";

// Options for the price-adjustment type select.
export const PRICE_ADJUSTMENT_OPTIONS: {
  value: PriceAdjustmentType;
  label: string;
}[] = [
  { value: "none", label: "No adjustment" },
  { value: "percent", label: "Percent (%)" },
  { value: "fixed", label: "Fixed amount ($)" },
];

// In forms the user enters the adjustment in display units (percentage points
// for "percent", dollars for "fixed"). The API/data model stores "fixed" in
// cents, so convert before sending.
export function adjustmentValueToCents(
  type: PriceAdjustmentType,
  displayValue: number,
): number {
  return type === "fixed" ? Math.round(displayValue * 100) : displayValue;
}

// Inverse of adjustmentValueToCents, for pre-filling a form from stored values.
export function adjustmentValueFromCents(
  type: PriceAdjustmentType,
  storedValue: number,
): number {
  return type === "fixed" ? storedValue / 100 : storedValue;
}

// Mirrors the backend data.ComputeAdjustedPriceCents. `storedValue` is in the
// stored units (percentage points for "percent", cents for "fixed"). Used for
// live price previews; the authoritative price still comes from the API.
export function computeAdjustedPriceCents(
  baseCents: number,
  type: PriceAdjustmentType,
  storedValue: number,
): number {
  if (type === "percent") {
    return Math.max(0, Math.round(baseCents * (1 + storedValue / 100)));
  }
  if (type === "fixed") {
    return Math.max(0, baseCents + Math.round(storedValue));
  }
  return baseCents;
}
