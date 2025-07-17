package ai

// ModerateTranscript mocks AI moderation.
// Replace with actual OpenAI logic if available.
func ModerateTranscript(transcript, speaker string) string {
	if transcript == "" {
		return ""
	}
	if containsBadWords(transcript) {
		return "Please use respectful language."
	}
	return ""
}

func containsBadWords(text string) bool {
	// Placeholder bad word check
	badWords := []string{"stupid", "hate", "idiot"}
	for _, word := range badWords {
		if contains(text, word) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || contains(s[1:], substr))
}
