package devin

import (
	"testing"

	"connectrpc.com/connect"
)

func testCatalog(uids ...string) map[string]bool {
	set := make(map[string]bool, len(uids))
	for _, uid := range uids {
		set[uid] = true
	}
	return set
}

func TestRouteClaudeEffort(t *testing.T) {
	catalog := testCatalog(
		"claude-opus-5-low", "claude-opus-5-medium", "claude-opus-5-high", "claude-opus-5-xhigh", "claude-opus-5-max",
		"claude-opus-5-high-fast",
		"claude-opus-4-6", "claude-opus-4-6-thinking",
		"claude-fable-5-2-low", "claude-fable-5-2-high",
		"claude-5-fable-low", "claude-5-fable-high", "claude-5-fable-max",
	)
	cases := []struct {
		name, model, effort, want string
	}{
		{"explicit effort", "claude-opus-5", "max", "claude-opus-5-max"},
		{"effort is case and space tolerant", "claude-opus-5", " XHigh ", "claude-opus-5-xhigh"},
		{"missing effort defaults to high", "claude-opus-5", "", "claude-opus-5-high"},
		{"unknown effort defaults to high", "claude-opus-5", "turbo", "claude-opus-5-high"},
		{"missing tier falls back downward", "claude-fable-5-2", "medium", "claude-fable-5-2-low"},
		{"missing tier above falls back upward", "claude-fable-5-2", "max", "claude-fable-5-2-high"},
		{"model without tiers keeps base name", "claude-opus-4-6", "high", "claude-opus-4-6"},
		{"legacy suffixed uid passes through", "claude-opus-5-low", "max", "claude-opus-5-low"},
		{"public fable name routes to upstream 5-fable tiers", "claude-fable-5", "max", "claude-5-fable-max"},
		{"public fable name falls back downward", "claude-fable-5", "xhigh", "claude-5-fable-high"},
		{"upstream fable name still accepted", "claude-5-fable", "low", "claude-5-fable-low"},
	}
	for _, c := range cases {
		got, err := routeClaudeEffort(c.model, c.effort, false, catalog)
		if err != nil || got != c.want {
			t.Errorf("%s: routeClaudeEffort(%q,%q) = (%q,%v), want %q", c.name, c.model, c.effort, got, err, c.want)
		}
	}
}

func TestRouteClaudeEffortWithoutCatalogConcatenates(t *testing.T) {
	got, err := routeClaudeEffort("claude-sonnet-5", "low", false, nil)
	if err != nil || got != "claude-sonnet-5-low" {
		t.Fatalf("got (%q,%v), want claude-sonnet-5-low", got, err)
	}
}

func TestRouteClaudeEffortBlocksFastVariants(t *testing.T) {
	catalog := testCatalog("claude-opus-5-high-fast")
	for _, model := range []string{"claude-opus-5-high-fast", "claude-opus-5-fast", "swe-1-6-fast", "claude-opus-4-6-thinking", "claude-sonnet-4-6-thinking-1m", "claude-opus-4-6-1m"} {
		_, err := routeClaudeEffort(model, "high", false, catalog)
		if err == nil {
			t.Fatalf("routeClaudeEffort(%q) returned no error, want not-found", model)
		}
		if connect.CodeOf(err) != connect.CodeNotFound {
			t.Fatalf("routeClaudeEffort(%q) code = %v, want NotFound", model, connect.CodeOf(err))
		}
	}
}

func TestRouteClaudeEffortLongContext(t *testing.T) {
	catalog := testCatalog(
		"claude-opus-4-6", "claude-opus-4-6-1m",
		"claude-sonnet-4-6", "claude-sonnet-4-6-1m",
		"claude-opus-5-high", "claude-opus-5-max",
	)
	cases := []struct {
		name, model, effort string
		long                bool
		want                string
	}{
		{"declared 1m routes to 1m variant", "claude-opus-4-6", "high", true, "claude-opus-4-6-1m"},
		{"undeclared keeps base context", "claude-opus-4-6", "high", false, "claude-opus-4-6"},
		{"declared 1m on sonnet 4-6", "claude-sonnet-4-6", "", true, "claude-sonnet-4-6-1m"},
		{"tiered model already 1m, no variant", "claude-opus-5", "max", true, "claude-opus-5-max"},
		{"declared 1m without variant ignored", "claude-haiku-5-5", "low", true, "claude-haiku-5-5-low"},
	}
	for _, c := range cases {
		got, err := routeClaudeEffort(c.model, c.effort, c.long, catalog)
		if err != nil || got != c.want {
			t.Errorf("%s: got (%q,%v), want %q", c.name, got, err, c.want)
		}
	}
	if got, _ := routeClaudeEffort("claude-opus-4-6", "high", true, nil); got != "claude-opus-4-6-high" {
		t.Errorf("nil catalog with long context: got %q", got)
	}
}

func TestStripLongContextSuffix(t *testing.T) {
	if base, long := stripLongContextSuffix("claude-opus-4-6[1m]"); base != "claude-opus-4-6" || !long {
		t.Fatalf("got (%q,%v)", base, long)
	}
	if base, long := stripLongContextSuffix("claude-opus-4-6"); base != "claude-opus-4-6" || long {
		t.Fatalf("got (%q,%v)", base, long)
	}
}
