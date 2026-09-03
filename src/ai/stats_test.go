package ai

import (
	"testing"
	"time"
)

func TestEstimateOutputTokensAndFormatStatistics(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int
	}{
		{"", 0},
		{"abc", 1},
		{"hello", 2},
		{"éééé", 1},
		{"12345678", 2},
	} {
		if got := EstimateOutputTokens(tc.text); got != tc.want {
			t.Errorf("EstimateOutputTokens(%q) = %d, want %d", tc.text, got, tc.want)
		}
	}

	for _, tc := range []struct {
		name    string
		tokens  int
		elapsed time.Duration
		want    string
	}{
		{"unusable tokens", 0, time.Second, ""},
		{"zero duration", 4, 0, ""},
		{"sub-second guard", 5, 500 * time.Millisecond, "5t/s 1s"},
		{"seconds", 10, 3 * time.Second, "3t/s 3s"},
		{"minutes and seconds", 300, 65 * time.Second, "4t/s 1m 5s"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatResponseStatistics(tc.tokens, tc.elapsed); got != tc.want {
				t.Fatalf("FormatResponseStatistics(%d, %s) = %q, want %q", tc.tokens, tc.elapsed, got, tc.want)
			}
		})
	}
}
