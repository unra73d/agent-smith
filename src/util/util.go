// Package util implements global utility methods
package util

import "strings"

var thinkTags = []string{"think", "thinking"}

func CutThinking(text string) string {
	for _, tag := range thinkTags {
		if strings.HasPrefix(text, "<"+tag+">") {
			pos := strings.Index(text, "</"+tag+">")
			if pos != -1 {
				text = text[pos+len("</"+tag+">"):]
			}
			break
		}
	}
	return strings.TrimSpace(text)
}
