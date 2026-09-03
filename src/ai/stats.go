package ai

import (
	"fmt"
	"time"
	"unicode/utf8"
)

// EstimateOutputTokens provides a deterministic fallback when a provider does
// not include completion usage. Four characters per token is a compact,
// provider-independent approximation for generated text.
func EstimateOutputTokens(text string) int {
	characters := utf8.RuneCountInString(text)
	if characters == 0 {
		return 0
	}
	return (characters + 3) / 4
}

func FormatResponseStatistics(outputTokens int, elapsed time.Duration) string {
	if outputTokens <= 0 || elapsed <= 0 {
		return ""
	}

	seconds := int(elapsed.Seconds())
	if seconds < 1 {
		seconds = 1
	}

	throughput := outputTokens / seconds
	if seconds >= 60 {
		return fmt.Sprintf("%dt/s %dm %ds", throughput, seconds/60, seconds%60)
	}
	return fmt.Sprintf("%dt/s %ds", throughput, seconds)
}
