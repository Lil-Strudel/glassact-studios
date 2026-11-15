package catalog

import (
	"strings"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

// Tags match on whole words only: a substring match would make "car" also
// return cardinal, carnation and cartoon. Every term must match, so extra
// words narrow the results.
func matchesSearch(item *data.CatalogItem, search string) bool {
	name := strings.ToLower(item.Name)
	code := strings.ToLower(item.CatalogCode)

	tagWords := make(map[string]bool)
	for _, tag := range item.Tags {
		for _, word := range strings.Fields(strings.ToLower(tag)) {
			tagWords[word] = true
		}
	}

	for _, term := range strings.Fields(strings.ToLower(search)) {
		if !strings.Contains(name, term) && !strings.Contains(code, term) && !tagWords[term] {
			return false
		}
	}

	return true
}
