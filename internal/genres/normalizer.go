package genres

import (
	"regexp"
	"strings"
)

var excludedTags = map[string]struct{}{
	"seen live":             {},
	"favorites":             {},
	"favourite":             {},
	"favorite":              {},
	"male vocalists":        {},
	"female vocalists":      {},
	"female vocals":         {},
	"male vocals":           {},
	"beautiful":             {},
	"awesome":               {},
	"love":                  {},
	"cool":                  {},
	"under 2000 listeners": {},
	"all":                   {},
	// Common language / nationality tags that are not musical genres
	"english":    {},
	"american":   {},
	"british":    {},
	"canadian":   {},
	"australian": {},
	"german":     {},
	"french":     {},
	"italian":    {},
	"spanish":    {},
	"swedish":    {},
	"norwegian":  {},
	"danish":     {},
	"finnish":    {},
	"japanese":   {},
	"korean":     {},
	"chinese":    {},
	"irish":      {},
	"scottish":   {},
	"welsh":      {},
	"russian":    {},
	"polish":     {},
}

var yearRegex = regexp.MustCompile(`^(19|20)\d{2}s?$`)
var whitespaceRegex = regexp.MustCompile(`\s+`)

// NormalizeGenres cleans, trims, deduplicates, and filters non-genre tags.
func NormalizeGenres(rawGenres []string) []string {
	seen := make(map[string]struct{})
	var result []string

	for _, raw := range rawGenres {
		cleaned := strings.TrimSpace(strings.ToLower(raw))
		cleaned = whitespaceRegex.ReplaceAllString(cleaned, " ")
		cleaned = strings.TrimPrefix(cleaned, "the ")

		if len(cleaned) == 0 || len(cleaned) > 64 {
			continue
		}

		if _, excluded := excludedTags[cleaned]; excluded {
			continue
		}

		if yearRegex.MatchString(cleaned) {
			continue
		}

		if _, exists := seen[cleaned]; !exists {
			seen[cleaned] = struct{}{}
			result = append(result, cleaned)
		}
	}

	return result
}
