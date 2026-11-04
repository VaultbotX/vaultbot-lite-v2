package genres

import (
	"reflect"
	"testing"
)

func TestNormalizeGenres(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "lowercases, trims, and deduplicates",
			input:    []string{"Pop", "  pop  ", "INDIE ROCK", "Indie Rock"},
			expected: []string{"pop", "indie rock"},
		},
		{
			name:     "filters year tags and validates length",
			input:    []string{"2020s", "1995", "", "synth-pop", "dance pop"},
			expected: []string{"synth-pop", "dance pop"},
		},
		{
			name:     "collapses whitespace and strips leading 'the '",
			input:    []string{"the  indie   pop", " alternative   rock "},
			expected: []string{"indie pop", "alternative rock"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeGenres(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NormalizeGenres(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
