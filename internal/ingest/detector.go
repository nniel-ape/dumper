package ingest

import (
	"regexp"
	"strings"
	"unicode"
)

// wordSeparators matches common separators in technical terms
var topicPattern = regexp.MustCompile(`^[\p{L}\p{N}\-\.\/\+\#\s]+$`)

// IsShortTopicMessage detects if a message is a short topic suitable for web search.
// Returns true for messages like "kubernetes", "react hooks", "go 1.25"
// Returns false for questions, long sentences, or URLs.
func IsShortTopicMessage(text string) bool {
	text = strings.TrimSpace(text)

	// Check length bounds (2-50 chars)
	if len(text) < 2 || len(text) > 50 {
		return false
	}

	// Not a URL
	if IsURL(text) {
		return false
	}

	// No question marks (likely a question, should be a note)
	if strings.Contains(text, "?") {
		return false
	}

	// Count words (split on whitespace)
	words := strings.Fields(text)
	if len(words) < 1 || len(words) > 3 {
		return false
	}

	// Must match topic pattern (alphanumeric + common separators)
	if !topicPattern.MatchString(text) {
		return false
	}

	// Must contain at least one letter
	hasLetter := false
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return false
	}

	return true
}
