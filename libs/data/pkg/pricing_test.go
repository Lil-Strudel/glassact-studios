package data

import "testing"

func TestComputeAdjustedPriceCents(t *testing.T) {
	tests := []struct {
		name     string
		base     int
		adjType  PriceAdjustmentType
		adjValue float64
		want     int
	}{
		{"none ignores value", 10000, PriceAdjustmentTypes.None, 50, 10000},
		{"percent markup", 10000, PriceAdjustmentTypes.Percent, 20, 12000},
		{"percent discount", 10000, PriceAdjustmentTypes.Percent, -10, 9000},
		{"percent rounds to nearest cent", 9999, PriceAdjustmentTypes.Percent, 20, 11999},
		{"fixed markup in cents", 10000, PriceAdjustmentTypes.Fixed, 1221, 11221},
		{"fixed discount in cents", 10000, PriceAdjustmentTypes.Fixed, -2500, 7500},
		{"percent discount clamps at zero", 10000, PriceAdjustmentTypes.Percent, -150, 0},
		{"fixed discount clamps at zero", 1000, PriceAdjustmentTypes.Fixed, -5000, 0},
		{"unknown type falls back to base", 10000, PriceAdjustmentType("bogus"), 99, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeAdjustedPriceCents(tt.base, tt.adjType, tt.adjValue)
			if got != tt.want {
				t.Errorf("ComputeAdjustedPriceCents(%d, %q, %v) = %d, want %d",
					tt.base, tt.adjType, tt.adjValue, got, tt.want)
			}
		})
	}
}
