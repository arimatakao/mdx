package mangadexapi

import (
	"testing"
)

func TestGetMangaDexPaths(t *testing.T) {
	tests := []struct {
		name            string
		link            string
		isExpectedEmpty bool
	}{
		{
			name:            "Valid MangaDex Link",
			link:            "https://mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370/this-gorilla-will-die-in-1-day",
			isExpectedEmpty: false,
		},
		{
			name:            "Invalid Link",
			link:            "invalid-link",
			isExpectedEmpty: true,
		},
		{
			name:            "Host Not Mangadex.org",
			link:            "https://example.com/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370/this-gorilla-will-die-in-1-day",
			isExpectedEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMangaDexPaths(tt.link)
			if tt.isExpectedEmpty && len(result) > 0 {
				t.Errorf("Test Case: %s. Expected empty slice, but got %v", tt.name, result)
			}
		})
	}
}

func TestGetMangaIdFromUrl(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected string
	}{
		{
			name:     "Valid MangaDex Link",
			link:     "https://mangadex.org/title/a3f91d0b-02f5-4a3d-a2d0-f0bde7152370/this-gorilla-will-die-in-1-day",
			expected: "a3f91d0b-02f5-4a3d-a2d0-f0bde7152370",
		},
		{
			name:     "Invalid Link",
			link:     "invalid-link",
			expected: "",
		},
		{
			name:     "Link with Incorrect Path",
			link:     "https://mangadex.org/chapter/abc-123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMangaIdFromUrl(tt.link)
			if result != tt.expected {
				t.Errorf("Test Case: %s. Expected %s, but got %s", tt.name, tt.expected, result)
			}
		})
	}
}

func TestGetChapterIdFromUrl(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected string
	}{
		{
			name:     "Valid MangaDex Link",
			link:     "https://mangadex.org/chapter/7c5d2aea-ea55-47d9-8c65-a33c9e92df70",
			expected: "7c5d2aea-ea55-47d9-8c65-a33c9e92df70",
		},
		{
			name:     "Invalid Link",
			link:     "invalid-link",
			expected: "",
		},
		{
			name:     "Link with Incorrect Path",
			link:     "https://mangadex.org/title/abc-123",
			expected: "",
		},
		{
			name:     "Link with Incorrect Host",
			link:     "https://notmangadex.org/title/abc-123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetChapterIdFromUrl(tt.link)
			if result != tt.expected {
				t.Errorf("Test Case: %s. Expected %s, but got %s", tt.name, tt.expected, result)
			}
		})
	}
}
