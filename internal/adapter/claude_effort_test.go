package adapter

import "testing"

func TestSplitClaudeEffort(t *testing.T) {
	cases := []struct {
		uid, base, level string
		ok               bool
	}{
		{"claude-opus-5-high", "claude-opus-5", "high", true},
		{"claude-fable-5-2-xhigh", "claude-fable-5-2", "xhigh", true},
		{"claude-opus-5", "", "", false},
		{"claude-opus-4-6-thinking", "", "", false},
		{"swe-2-max", "", "", false},
		{"claude-opus-5-high-fast", "", "", false},
	}
	for _, c := range cases {
		base, level, ok := SplitClaudeEffort(c.uid)
		if base != c.base || level != c.level || ok != c.ok {
			t.Errorf("SplitClaudeEffort(%q) = (%q,%q,%v), want (%q,%q,%v)", c.uid, base, level, ok, c.base, c.level, c.ok)
		}
	}
}

func TestPublicModelsFoldsEffortAndBlocksFast(t *testing.T) {
	in := []ModelInfo{
		{ID: "claude-opus-5-high"},
		{ID: "claude-opus-5-low"},
		{ID: "claude-opus-5-high-fast"},
		{ID: "claude-5-fable-high"},
		{ID: "claude-5-fable-low"},
		{ID: "claude-opus-4-6"},
		{ID: "claude-opus-4-6-thinking"},
		{ID: "claude-opus-4-6-1m"},
		{ID: "claude-sonnet-4-6-thinking-1m"},
		{ID: "glm-5-2-1m"},
		{ID: "swe-2-max"},
		{ID: "swe-1-6-fast"},
	}
	var got []string
	for _, m := range PublicModels(in) {
		got = append(got, m.ID)
	}
	want := []string{"claude-opus-5", "claude-fable-5", "claude-opus-4-6", "glm-5-2-1m", "swe-2-max"}
	if len(got) != len(want) {
		t.Fatalf("PublicModels = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PublicModels = %v, want %v", got, want)
		}
	}
}
