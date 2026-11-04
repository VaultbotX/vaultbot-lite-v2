package genres

import (
	"regexp"
	"strings"
)

var yearRegex = regexp.MustCompile(`^(19|20)\d{2}s?$`)
var whitespaceRegex = regexp.MustCompile(`\s+`)

// NormalizeGenres cleans, trims, deduplicates, and validates genre strings.
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
