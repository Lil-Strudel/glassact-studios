package data

import "math"

// InstallationKitPriceCents is the flat add-on price charged per inlay when a
// dealership opts into an installation kit at order time. Keep this in sync with
// INSTALLATION_KIT_PRICE_CENTS in libs/data/src/installation-kits.ts.
const InstallationKitPriceCents = 4900

// ComputeAdjustedPriceCents applies a proof's price adjustment to a price
// group's base price. For "percent", adjValue is percentage points
// (20 = +20%); for "fixed", adjValue is cents (1221 = +$12.21). Adjustments may
// be negative (discounts); the result is clamped at zero.
func ComputeAdjustedPriceCents(baseCents int, adjType PriceAdjustmentType, adjValue float64) int {
	switch adjType {
	case PriceAdjustmentTypes.Percent:
		adjusted := float64(baseCents) * (1 + adjValue/100)
		return max(0, int(math.Round(adjusted)))
	case PriceAdjustmentTypes.Fixed:
		return max(0, baseCents+int(math.Round(adjValue)))
	default:
		return baseCents
	}
}
