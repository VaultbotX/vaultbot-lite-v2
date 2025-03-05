package spotify

import "testing"

func TestParseTrackId_ParsesAlphanumericText(t *testing.T) {
	// Arrange
	text := "1234567890"
	expected := "1234567890"

	// Act
	actual := ParseTrackId(text)

	// Assert
	if actual == nil {
		t.Errorf("Expected %s, got nil", expected)
	}

	if actual.String() != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}
