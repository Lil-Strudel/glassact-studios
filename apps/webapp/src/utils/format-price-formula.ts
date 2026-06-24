import type { PriceAdjustmentType } from "@glassact/data";
import { formatMoney } from "./format-money";

// Renders the pricing formula caption shown above a dollar amount, e.g.
// "PG1 + 20%", "PG1 - $5.00", or just "PG1" when there is no adjustment.
// Returns null when there is no price group to attribute the price to.
export function formatPriceFormula(
  priceGroupName: string | null | undefined,
  type: PriceAdjustmentType | null | undefined,
  value: number | null | undefined,
): string | null {
  if (!priceGroupName) {
    return null;
  }

  if (!type || type === "none" || !value) {
    return priceGroupName;
  }

  const sign = value < 0 ? "-" : "+";
  const magnitude = Math.abs(value);

  if (type === "percent") {
    return `${priceGroupName} ${sign} ${magnitude}%`;
  }

  // fixed: value is in cents
  return `${priceGroupName} ${sign} ${formatMoney(magnitude / 100)}`;
}
