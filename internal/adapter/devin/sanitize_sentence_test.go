package devin

import (
	"strings"
	"testing"
)

// 回归：规则不得吞掉同一行里命中句之后的其它句子（曾经删掉 "Keep answers short."）。
func TestSanitizeKeepsTextAfterMatchedSentence(t *testing.T) {
	cases := []struct {
		name, in, mustKeep string
	}{
		{
			"security line keeps trailing sentence",
			"IMPORTANT: Assist with authorized security testing for CTF challenges only. Keep answers short.",
			"Keep answers short.",
		},
		{
			"dual-use line keeps trailing sentence",
			"Dual-use security tools (C2 frameworks, credential testing, exploit development) require clear authorization context: CTF only. Reply in Chinese.",
			"Reply in Chinese.",
		},
		{
			"claude code cli line keeps trailing sentence",
			"Claude Code is available as a CLI in the terminal. Always use tabs.",
			"Always use tabs.",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hits := map[string]int{}
			got := sanitizeUpstreamText(c.in, true, hits)
			if !strings.Contains(got, c.mustKeep) {
				t.Fatalf("sanitized = %q, lost %q", got, c.mustKeep)
			}
		})
	}
}
