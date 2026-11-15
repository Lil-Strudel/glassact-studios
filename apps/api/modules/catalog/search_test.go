package catalog

import (
	"testing"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

func TestMatchesSearch(t *testing.T) {
	mustang := &data.CatalogItem{Name: "Mustang", CatalogCode: "B-OUT-0012", Tags: []string{"sports car", "vehicle"}}
	cardinal := &data.CatalogItem{Name: "Perched Bird", CatalogCode: "A-BRD-0003L", Tags: []string{"cardinal", "red bird"}}

	tests := []struct {
		name   string
		item   *data.CatalogItem
		search string
		want   bool
	}{
		{"empty search matches everything", mustang, "", true},
		{"tag word matches item without a name match", mustang, "car", true},
		{"tag word match ignores case", mustang, "CAR", true},
		{"tag substring does not match", cardinal, "car", false},
		{"name substring matches", mustang, "musta", true},
		{"code substring matches", cardinal, "brd-0003", true},
		{"multiple terms all matching", cardinal, "red bird", true},
		{"multiple terms require every term", cardinal, "red car", false},
		{"terms can mix name and tag", mustang, "mustang vehicle", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesSearch(tt.item, tt.search); got != tt.want {
				t.Errorf("matchesSearch(%q, %q) = %v, want %v", tt.item.Name, tt.search, got, tt.want)
			}
		})
	}
}
